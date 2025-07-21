package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/robin"
)

// WithAuth is a middleware that requires the user to be authenticated
func (m *middleware) WithAuth(ctx *robin.Context) error {
	authCookie, err := ctx.Request().Cookie(authlib.CookieAuthSession)
	if err != nil {
		return apperrors.ErrUnauthorized
	}

	logout := func() {
		// Remove the cookie
		ctx.SetCookie(&http.Cookie{
			Name:     authlib.CookieAuthSession,
			Value:    "",
			HttpOnly: true,
			Secure:   m.config.InProduction() || m.config.InStaging(),
			SameSite: http.SameSiteDefaultMode,
			Expires:  time.Now().Add(-time.Hour),
		})
	}

	authToken, err := authlib.DecodeSignedCookie(authCookie, m.config.Keys.CookieSecret)
	if err != nil {
		log.Debug().Err(err).Msg("failed to decode auth cookie")
		logout()

		return apperrors.ErrSessionExpired
	}

	// Look through the user's session data
	session, err := m.repository.AuthRepository().LookupAuthSession(authToken.Value())
	if err != nil {
		log.Debug().Err(err).Msg("failed to lookup session, logging out")
		logout()

		return err
	}

	// Check if the session is still valid
	if session.ExpiresAt.Before(time.Now()) {
		log.Debug().Msg("session has expired, logging out")
		logout()

		// Delete the session if hasn't been deleted already
		if err := m.repository.AuthRepository().RevokeAuthSession(authToken.Value()); err != nil {
			return err
		}
	}

	// Set the user ID in the context
	ctx.Set(authlib.StateKeyUserID, session.UserID)
	ctx.Set(authlib.StateKeySession, session)

	return nil
}
