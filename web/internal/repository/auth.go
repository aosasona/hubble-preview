package repository

import (
	"errors"
	"net/http"
	"time"

	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/models"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
)

type AuthRepository interface {
	// GenerateAuthSession generates a new session for the given user ID.
	GenerateAuthSession(userID int32) (models.AuthSession, error)

	// LookupAuthSession looks up a session by token.
	LookupAuthSession(token string) (models.AuthSession, error)

	// RevokeAuthSession revokes a session by token.
	RevokeAuthSession(token string) error
}

type authRepo struct {
	*baseRepo
}

const (
	// AuthSessionTTL is the time-to-live for a session - defaults to 30 days
	AuthSessionTTL = time.Hour * 24 * 30
)

// GenerateAuthSession generates a new session for the given user ID.
func (a *authRepo) GenerateAuthSession(userID int32) (models.AuthSession, error) {
	var session models.AuthSession

	token, err := lib.GenerateAuthToken()
	if err != nil {
		return session, err
	}

	session.UserID = userID
	session.Token = token.String()
	session.Suspended = false
	session.TTL = AuthSessionTTL
	session.IssuedAt = time.Now()
	session.ExpiresAt = time.Now().Add(AuthSessionTTL)

	// Save to store
	if err := a.store.SetJsonWithTTL(kv.KeyAuthSession(token.String()), session, AuthSessionTTL); err != nil {
		return session, err
	}

	return session, nil
}

// LookupAuthSession looks up a session by token.
func (a *authRepo) LookupAuthSession(token string) (models.AuthSession, error) {
	session := models.AuthSession{}

	err := a.store.GetJson(kv.KeyAuthSession(token), &session)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return session, apperrors.NewAuthorizationError(
				"This session might have expired, please sign in again",
				http.StatusUnauthorized,
			)
		}

		return session, err
	}

	return session, nil
}

// RevokeAuthSession revokes a session.
func (a *authRepo) RevokeAuthSession(token string) error {
	return a.store.Delete(kv.KeyAuthSession(token))
}

var _ AuthRepository = (*authRepo)(nil)
