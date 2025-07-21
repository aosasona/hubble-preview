package otp

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewhartstonge/argon2"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/kv"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
)

type Manager interface {
	// GenerateToken generates a scoped token for a user.
	GenerateToken(tokenType TokenType, identifier any) (string, error)

	// FindAccountVerificationToken finds an email verification token by user ID.
	FindToken(tokenType TokenType, identifier any) (token Token, err error)

	// VerifyToken verifies a token.
	VerifyToken(VerifyTokenParams) error

	// DeleteToken deletes a user token by user ID.
	DeleteToken(tokenType TokenType, identifier any) error

	// GetTokenTimeToLive returns the time to live for a token type.
	GetTokenTimeToLive(tokenType TokenType) (time.Duration, error)

	// getKey returns the key for a token.
	getKey(tokenType TokenType, identifier any) (kv.KeyContainer, error)

	// getTokenConfig returns the token configuration for a token type.
	getTokenConfig(tokenType TokenType) (TokenConfig, error)
}

type tokenRepo struct {
	store kv.Store
}

type Token struct {
	Hash     []byte `json:"hash"`
	Attempts int    `json:"attempts"`
}

// VerifyTokenParams contains the parameters required to verify a token.
type VerifyTokenParams struct {
	// The type of token to verify (email verification, password reset, etc.)
	Type TokenType

	// The identifier associated with the token
	Identifier any

	// The token to verify (i.e. the user-provided token)
	ProvidedToken string

	// RetainAfterVerified determines if the token should be retained after verification.
	RetainAfterVerified bool
}

func NewManager(store kv.Store) Manager {
	return &tokenRepo{store: store}
}

// DeleteToken removes a token from the store.
func (t *tokenRepo) DeleteToken(tokenType TokenType, identifier any) error {
	key, err := t.getKey(tokenType, identifier)
	if err != nil {
		return err
	}

	return t.store.Delete(key)
}

// FindToken finds a token by the provided identifier.
func (t *tokenRepo) FindToken(tokenType TokenType, identifier any) (Token, error) {
	key, err := t.getKey(tokenType, identifier)
	if err != nil {
		return Token{}, err
	}

	// Check if the token has expired
	expired, err := t.store.ExpiresAt(key)
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return Token{}, apperrors.New(
				"your token might have expired or is invalid, please request a new one",
				http.StatusBadRequest,
			)
		}

		return Token{}, err
	}

	// The token ideally should not be present when it has expired, but we will assume expiry one second ahead and return an error if it has expired then
	if expired.Before(time.Now().Add(time.Second)) {
		return Token{}, apperrors.New("verification token has expired", http.StatusGone)
	}

	token := Token{}
	if err = t.store.GetJson(key, &token); err != nil {
		return Token{}, err
	}

	return token, nil
}

// GenerateToken generates a token for a certain identifier.
func (t *tokenRepo) GenerateToken(tokenType TokenType, identifier any) (string, error) {
	var (
		key         kv.KeyContainer
		tokenConfig TokenConfig
		err         error
	)

	if key, err = t.getKey(tokenType, identifier); err != nil {
		return "", err
	}

	if tokenConfig, err = t.getTokenConfig(tokenType); err != nil {
		return "", err
	}

	// Ensure that the user does not have a valid token that was generated before the MinResendInterval
	expiresAt, err := t.store.ExpiresAt(key)
	if err != nil && !errors.Is(err, kv.ErrKeyNotFound) && !errors.Is(err, kv.ErrNoExpiry) {
		return "", err
	}

	// If the token has not been around for MinResendInterval, we don't want to generate a new one
	expectedMinDurationLeft := tokenConfig.TimeToLive - tokenConfig.MinResendInterval
	if expiresAt.After(time.Now().Add(expectedMinDurationLeft)) {
		tryAfter := expiresAt.Add(-expectedMinDurationLeft)
		return "", apperrors.New(
			fmt.Sprintf(
				"please wait %s before requesting a new token",
				lib.ToHumanReadableDuration(time.Until(tryAfter).Truncate(time.Second)),
			),
			http.StatusTooEarly,
		)
	}

	// Generate a new token
	token, err := lib.GenerateToken(string(tokenType))
	if err != nil {
		return "", err
	}

	value := Token{Hash: token.EncodedByte(), Attempts: 0}
	if err := t.store.SetJsonWithTTL(key, value, tokenConfig.TimeToLive); err != nil {
		return "", err
	}

	return token.String(), nil
}

// VerifyToken verifies a token.
func (t *tokenRepo) VerifyToken(params VerifyTokenParams) error {
	savedToken, err := t.FindToken(params.Type, params.Identifier)
	if err != nil {
		return err
	}

	// Load the max attempts from the token config
	tokenConfig, err := t.getTokenConfig(params.Type)
	if err != nil {
		return seer.Wrap("get_token_config", err)
	}

	maxAttempts := DefaultMaxAttempts
	if tokenConfig.MaxAttempts > 0 {
		maxAttempts = tokenConfig.MaxAttempts
	}

	// Ensure we haven't exceeded the maximum number of attempts
	if savedToken.Attempts >= maxAttempts {
		return apperrors.New(
			"maximum number of attempts exceeded, please request a new token",
			http.StatusTooManyRequests,
		)
	}

	// Increment the number of attempts
	defer func() {
		savedToken.Attempts++
		key, err := t.getKey(params.Type, params.Identifier)
		if err != nil {
			log.Error().Err(err).Msg("failed to get key for incrementing token attempts")
			return
		}

		if err := t.store.SetJsonWithTTL(key, savedToken, tokenConfig.TimeToLive); err != nil {
			log.Error().Err(err).Msg("failed to update token attempts")
		}
	}()

	// Compare hashes (we only store hashed tokens)
	ok, err := argon2.VerifyEncoded([]byte(params.ProvidedToken), []byte(savedToken.Hash))
	if err != nil {
		return seer.Wrap("verify_token", err)
	}
	if !ok {
		return apperrors.New("invalid token provided", http.StatusUnauthorized)
	}

	// Delete the token after successful verification unless specified otherwise
	if !params.RetainAfterVerified {
		go func() {
			if err := t.DeleteToken(params.Type, params.Identifier); err != nil {
				log.Error().Err(err).Msg("failed to delete token")
			}
		}()
	}

	return nil
}

// GetTokenTimeToLive returns the time to live for a token type.
func (t *tokenRepo) GetTokenTimeToLive(tokenType TokenType) (time.Duration, error) {
	config, err := t.getTokenConfig(tokenType)
	if err != nil {
		return 0, seer.Wrap("get_token_time_to_live", err)
	}

	return config.TimeToLive, nil
}

func (t *tokenRepo) getKey(tokenType TokenType, identifier any) (kv.KeyContainer, error) {
	config, ok := tokenConfigs[tokenType]
	if !ok {
		return kv.NoopKey(), fmt.Errorf("invalid token type: %s", tokenType)
	}

	return config.KeyHandler(identifier)
}

func (t *tokenRepo) getTokenConfig(tokenType TokenType) (TokenConfig, error) {
	config, ok := tokenConfigs[tokenType]
	if !ok {
		return TokenConfig{}, fmt.Errorf("invalid token type: %s", tokenType)
	}

	return config, nil
}
