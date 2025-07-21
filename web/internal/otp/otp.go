package otp

import (
	"fmt"
	"time"

	"go.trulyao.dev/hubble/web/internal/kv"
)

type TokenType string

const (
	// TokenVerification is used for email verification (signup/sign in)
	TokenUserAccountVerification TokenType = "email_verification"
	// TokenPasswordReset is used for password reset
	TokenUserPasswordReset TokenType = "password_reset"
	// TokenMfa is used for email multi-factor authentication requests
	TokenUserMfaVerification TokenType = "mfa"
	// TokenEmailChange is used for email change requests
	TokenUserEmailChange TokenType = "email_change"
)

// When the MaxAttempts is set to 0, this default value is used.
const DefaultMaxAttempts = 5

type (
	KeyHandler func(identifier any) (kv.KeyContainer, error)

	TokenConfig struct {
		// TimeToLive is the duration a token is valid for.
		TimeToLive time.Duration

		/*
			MinResendInterval is the minimum time between resending a token i.e. the amount of time that must have elapsed before the same token can be regenerated and resent

			For example, if our TimeToLive is 10m and the MinResendInterval is 4m, a user can request a new token if the current one has been in the store for AT LEAST 4 minutes.
		*/
		MinResendInterval time.Duration

		// MaxAttempts is the maximum number of attempts a user can make to verify a token.
		MaxAttempts int

		// KeyHandler is a function that returns a key container for a given identifier.
		KeyHandler KeyHandler
	}
)

var tokenConfigs = map[TokenType]TokenConfig{
	TokenUserAccountVerification: {
		TimeToLive:        time.Minute * 10,
		MinResendInterval: time.Minute * 5,
		KeyHandler:        require(kv.KeyEmailVerificationToken),
	},
	TokenUserPasswordReset: {
		TimeToLive:        time.Minute * 10,
		MinResendInterval: time.Minute * 5,
		KeyHandler:        require(kv.KeyPasswordResetToken),
	},
	TokenUserMfaVerification: {
		TimeToLive:        time.Minute * 5,
		MinResendInterval: time.Minute * 2,
		KeyHandler:        require(kv.KeyEmailMfaToken),
	},
	TokenUserEmailChange: {
		TimeToLive:        time.Minute * 10,
		MinResendInterval: time.Minute * 5,
		KeyHandler:        require(kv.KeyEmailChangeToken),
	},
}

func require[T comparable](
	container func(T) kv.KeyContainer,
) func(any) (kv.KeyContainer, error) {
	return func(identifier any) (kv.KeyContainer, error) {
		keyIdentifier, ok := identifier.(T)
		if !ok {
			return kv.NoopKey(), fmt.Errorf(
				"invalid identifier type, expected `%T` for %s",
				keyIdentifier,
				TokenUserAccountVerification,
			)
		}

		return container(keyIdentifier), nil
	}
}
