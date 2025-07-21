package ratelimit

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/kv"
)

const KeySeparator = ":_:"

type (
	Limit struct {
		MaxRequests int
		Interval    time.Duration
	}

	State struct {
		Counter int       `json:"counter"`
		LastHit time.Time `json:"last_hit"`
	}

	RateLimiter interface {
		// Increment increments the counter for the given key and returns the new value.
		Increment(key string) (int, error)

		// Reset resets the counter for the given key.
		Reset(key string) error

		// GetState gets the state for the given key.
		GetState(key string) (State, error)

		// GetResetTime gets the time when the rate limit will reset for the given key.
		GetResetTime(key string) (time.Time, error)

		// HasReachedLimit checks if the rate limit has been reached for the given key.
		HasReachedLimit(key string) (bool, error)

		// GetLimit gets the rate limit for the given key.
		GetLimit(key string) (Limit, error)

		// SetLimits sets the rate limits for the rate limiter.
		SetLimits(limits map[string]Limit)

		// Persist persists the rate limiter state.
		Persist() error

		// Close closes the rate limiter.
		Close() error
	}

	DefaultRateLimiter struct {
		scope  string
		store  kv.Store
		limits map[string]Limit
		state  map[string]State
		mu     sync.RWMutex
	}

	Deps struct {
		Store  kv.Store
		Limits map[string]Limit
		Scope  string
	}
)

func (r *DefaultRateLimiter) SetLimits(limits map[string]Limit) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.limits = limits
}

func (r *DefaultRateLimiter) Increment(key string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.GetLimit(key)
	if err != nil {
		return 0, err
	}

	state, ok := r.state[key]
	if !ok {
		state = State{}
	}

	state.Counter++
	state.LastHit = time.Now()

	r.state[key] = state
	return state.Counter, nil
}

func (r *DefaultRateLimiter) GetState(key string) (State, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.state[key]
	if !ok {
		return State{}, fmt.Errorf("state not found for key: %s", key)
	}

	return state, nil
}

func (r *DefaultRateLimiter) GetLimit(key string) (Limit, error) {
	// Extract the limit
	var (
		limit Limit
		ok    bool
	)

	// Check if the key has an identifier
	key, _ = ExtractKeyParts(key)

	if limit, ok = r.limits[key]; !ok {
		// Check for the "*" default rate limit
		if limit, ok = r.limits["*"]; !ok {
			return Limit{}, errors.New("rate limit not found")
		}
	}

	return limit, nil
}

func (r *DefaultRateLimiter) GetResetTime(key string) (time.Time, error) {
	state, err := r.GetState(key)
	if err != nil {
		return time.Time{}, err
	}

	limit, err := r.GetLimit(key)
	if err != nil {
		return time.Time{}, err
	}

	resetTime := state.LastHit.Add(limit.Interval)
	return resetTime, nil
}

func (r *DefaultRateLimiter) HasReachedLimit(key string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, ok := r.state[key]
	if !ok {
		return false, nil
	}

	limit, err := r.GetLimit(key)
	if err != nil {
		return false, err
	}

	hasCrossedLimit := (state.Counter >= limit.MaxRequests)
	isWithinInterval := (time.Since(state.LastHit) < limit.Interval)

	// If the counter has crossed the limit and the interval has not passed
	if hasCrossedLimit && isWithinInterval {
		return true, nil
	}

	// Reset the counter if the interval has passed
	if !isWithinInterval {
		state.Counter = 0
		r.state[key] = state
	}

	return false, nil
}

func (r *DefaultRateLimiter) Reset(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.state, key)
	return nil
}

func (r *DefaultRateLimiter) Persist() error {
	log.Info().Msg("persisting rate limiter state")
	err := r.store.SetJson(kv.KeyRateLimiterState(r.scope), r.state)
	return err
}

func (r *DefaultRateLimiter) Close() error {
	return r.Persist()
}

func NewDefaultRateLimiter(deps Deps) *DefaultRateLimiter {
	state := make(map[string]State)

	scope := "app"
	if deps.Scope != "" {
		scope = deps.Scope
	}

	// Attempt to load the state from the store
	if deps.Store != nil {
		log.Info().Msg("loading rate limiter state from store")
		if err := deps.Store.GetJson(kv.KeyRateLimiterState(scope), &state); err != nil {
			log.Error().Err(err).Msg("failed to load rate limiter state from KV store")
			state = make(map[string]State)
		}
	}

	return &DefaultRateLimiter{
		store:  deps.Store,
		state:  state,
		limits: deps.Limits,
		scope:  scope,
	}
}

// WithIdentifier returns a key with an identifier appended to it.
func WithIdentifier(key, identifier string) string {
	return fmt.Sprintf("%s%s%s", key, KeySeparator, identifier)
}

func ExtractKeyParts(complexKey string) (key string, identifier string) {
	if index := strings.LastIndex(complexKey, KeySeparator); index != -1 {
		return complexKey[:index], complexKey[index+len(KeySeparator):]
	}

	return complexKey, ""
}

var _ RateLimiter = (*DefaultRateLimiter)(nil)
