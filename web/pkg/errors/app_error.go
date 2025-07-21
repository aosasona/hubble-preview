package apperrors

import "fmt"

// This is the default generic application error type that will be used for errors that are intended to make their way to the client.
type Error struct {
	message string
	code    int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.message
}

// Code returns the error code.
func (e *Error) Code() int {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) String() string {
	return fmt.Sprintf("[app-error:%d] %s", e.code, e.message)
}

// New creates a new application error with the provided message and code.
func New(message string, code int) *Error {
	return &Error{message: message, code: code}
}

func BadRequest(message string) *Error {
	return New(message, 400)
}

func Unauthorized(message string) *Error {
	return New(message, 401)
}

func Forbidden(message string) *Error {
	return New(message, 403)
}

func ServerError(message string) *Error {
	return New(message, 500)
}
