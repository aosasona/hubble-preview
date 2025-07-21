package ratelimit_test

import (
	"testing"
	"time"

	ratelimit "go.trulyao.dev/hubble/web/internal/ratelimit"
)

const (
	TestProcedure = "testProcedure"
	Interval      = 3 * time.Second
)

func Test_WithIdentifier(t *testing.T) {
	// Rate limit instance
	rateLimiter := ratelimit.NewDefaultRateLimiter(ratelimit.Deps{
		Store: nil,
		Limits: map[string]ratelimit.Limit{
			TestProcedure: {MaxRequests: 20, Interval: Interval},
		},
		Scope: "test",
	})

	key := ratelimit.WithIdentifier(TestProcedure, "some_identifier")

	// Check limit before incrementing
	_, err := rateLimiter.HasReachedLimit(key)
	if err != nil {
		t.Errorf("failed to check rate limit before increment: %v", err)
	}

	// Simulate some limit increments
	for i := 0; i < 20; i++ {
		if _, err := rateLimiter.Increment(key); err != nil {
			t.Errorf("failed to increment counter: %v", err)
		}
	}

	// Check for an actual state
	state, err := rateLimiter.GetState(key)
	if err != nil {
		t.Errorf("failed to get state: %v", err)
	}

	if state.Counter != 20 {
		t.Errorf("expected counter to be 20, got %d", state.Counter)
	}
}

func Test_ComplexKeys(t *testing.T) {
	key := ratelimit.WithIdentifier("procedure", "identifier")
	procedure, identifier := ratelimit.ExtractKeyParts(key)

	if procedure != "procedure" {
		t.Errorf("expected procedure to be 'procedure', got '%s'", procedure)
	}

	if identifier != "identifier" {
		t.Errorf("expected identifier to be 'identifier', got '%s'", identifier)
	}
}

func Test_DefaultRateLimiter(t *testing.T) {
	// Rate limit instance
	rateLimiter := ratelimit.NewDefaultRateLimiter(ratelimit.Deps{
		Store: nil,
		Limits: map[string]ratelimit.Limit{
			TestProcedure: {MaxRequests: 20, Interval: Interval},
		},
		Scope: "test",
	})

	// Check limit before incrementing
	reached, err := rateLimiter.HasReachedLimit(TestProcedure)
	if err != nil {
		t.Errorf("failed to check rate limit before increment: %v", err)
	}

	if reached {
		t.Errorf("expected rate limit to not be reached")
	}

	// Simulate 20 requests in {Interval} seconds
	for i := 0; i < 20; i++ {
		if _, err := rateLimiter.Increment(TestProcedure); err != nil {
			t.Errorf("failed to increment counter: %v", err)
		}
	}

	// Check if the rate limit has been exceeded
	reached, err = rateLimiter.HasReachedLimit(TestProcedure)
	if err != nil {
		t.Errorf("failed to check rate limit before sleep: %v", err)
	}

	if !reached {
		t.Errorf("expected rate limit to be reached")
	}

	// Check for an unknown procedure
	_, err = rateLimiter.Increment("unknownProcedure")
	if err == nil {
		t.Errorf("expected error for unknown procedure")
	}

	// Wait for {Interval} seconds
	time.Sleep(Interval)

	// Check if the rate limit has been reset
	reached, err = rateLimiter.HasReachedLimit(TestProcedure)
	if err != nil {
		t.Errorf("failed to check rate limit after sleep: %v", err)
	}

	if reached {
		t.Errorf("expected rate limit to be reset after %d seconds", Interval)
	}

	// Check the state for the procedure
	state, err := rateLimiter.GetState(TestProcedure)
	if err != nil {
		t.Errorf("unexpected error while checking state: %v", err)
	}

	if state.Counter != 0 {
		t.Errorf("expected counter to be 0, got %d", state.Counter)
	}
}
