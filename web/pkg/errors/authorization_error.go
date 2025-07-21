package apperrors

import (
	"net/http"

	"github.com/goccy/go-json"
	"go.trulyao.dev/robin"
)

type AuthorizationError struct {
	// message is the error message.
	message string

	// Code is the HTTP status code.
	code int
}

var (
	ErrUnauthorized = NewAuthorizationError(
		"you need to be signed in to continue",
		http.StatusUnauthorized,
	)
	ErrIncompleteSession = NewAuthorizationError(
		"incomplete session data, please sign in again",
		http.StatusUnauthorized,
	)
	ErrSessionExpired = NewAuthorizationError(
		"session is invalid or expired",
		http.StatusUnauthorized,
	)
)

func NewAuthorizationError(message string, code int) *AuthorizationError {
	return &AuthorizationError{
		message: message,
		code:    code,
	}
}

func (e *AuthorizationError) Error() string {
	errorResponse := map[string]string{
		"type":    "authz-error",
		"message": uppercaseFirstLetter(e.message),
	}

	jsonBytes, err := json.Marshal(errorResponse)
	if err != nil {
		return "Unserializable authentication error"
	}

	return string(jsonBytes)
}

func (e *AuthorizationError) Serialize() robin.Serializable {
	return json.RawMessage(e.Error())
}

func (e *AuthorizationError) Code() int {
	return e.code
}
