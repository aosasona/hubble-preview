package apperrors

import (
	"strings"

	"github.com/rs/zerolog/log"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
)

func ErrorHandler(err error) (robin.Serializable, int) {
	var (
		defaultMessage = "An unknown error occurred, we have been notified and are working on it, please try again later."
		message        robin.Serializable
		code           = 500
	)

	switch e := err.(type) {
	case *Error:
		log.Err(err).Str("source", "apperrors").Str("kind", "base").Msg("An app error occurred")
		message, code = robin.ErrorString(uppercaseFirstLetter(e.Message())), e.Code()

	case *ValidationError:
		log.Err(err).Str("source", "apperrors").Str("kind", "validation").Msg("A validation error occurred")
		message, code = e.Serialize(), e.Code()

	case *AuthorizationError:
		log.Err(err).Str("source", "apperrors").Str("kind", "authorization").Msg("An authorization error occurred")
		message, code = e.Serialize(), e.Code()

	case *seer.Seer:
		l := log.Err(e).
			Str("source", "seer").
			Str("stacktrace", e.ErrorWithStackTrace())

		if e.OriginalError() != nil {
			l.Str("original_error", e.OriginalError().Error())
		}

		l.Msg("A seer error occurred")

		message, code = robin.ErrorString(uppercaseFirstLetter(e.Message())), e.Code()

	case *robin.Error:
		log.Err(err).Str("source", "robin").Msg("A robin error occurred")
		message = robin.ErrorString(defaultMessage)
		if e.Code == 404 {
			message, code = robin.ErrorString(uppercaseFirstLetter(e.Message)), 404
		}

	default:
		log.Err(err).Str("source", "unknown").Msg("An unknown error occurred")
		message = robin.ErrorString(defaultMessage)
	}

	return message, code
}

func uppercaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
