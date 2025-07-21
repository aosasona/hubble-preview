package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/ratelimit"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/robin"
)

func (m *middleware) WithRateLimit(ctx *robin.Context) error {
	// Skip rate limiting for development mode
	if m.config.InDevelopment() {
		return nil
	}

	ipAddress, err := lib.GetRequestIP(ctx.Request())
	if err != nil {
		return err
	}

	key := ratelimit.WithIdentifier(ctx.ProcedureName(), ipAddress)

	reachedLimit, err := m.rateLimiter.HasReachedLimit(key)
	if err != nil {
		log.Error().Err(err).Msg("failed to check rate limit")
		return err
	}

	if reachedLimit {
		errorMessage := "rate limit exceeded for this action, try again later"

		resetTime, err := m.rateLimiter.GetResetTime(key)
		if err == nil {
			tryAgainIn := time.Until(resetTime)
			errorMessage = fmt.Sprintf(
				"rate limit exceeded, try again in %s",
				lib.ToHumanReadableDuration(tryAgainIn.Truncate(time.Second)),
			)
		}

		return apperrors.New(errorMessage, http.StatusTooManyRequests)
	}

	counter, err := m.rateLimiter.Increment(key)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment rate limit counter")
	}

	if m.config.Debug() {
		log.Debug().
			Str("procedure", ctx.ProcedureName()).
			Str("ip_address", ipAddress).
			Int("counter", counter).
			Msg("rate limit check")
	}

	return nil
}
