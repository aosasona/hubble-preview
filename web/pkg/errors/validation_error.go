package apperrors

import (
	"fmt"

	"github.com/goccy/go-json"

	"go.trulyao.dev/robin"
)

type ErrorMap map[string][]string

// This is the more specialized error type that will be used for validation errors.
type ValidationError struct {
	// Errs is a map of field names to a list of error messages.
	// Example: {"email": ["is required", "must be a valid email"]}
	Errs ErrorMap `json:"errors"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	errs := map[string]any{
		"type":   "validation-error",
		"errors": e.Errs,
	}

	jsonBytes, err := json.Marshal(errs)
	if err != nil {
		return "Unserializable validation error"
	}

	return string(jsonBytes)
}

func (e *ValidationError) Serialize() robin.Serializable {
	return json.RawMessage(e.Error())
}

func (e *ValidationError) Code() int {
	return 400
}

func (e *ValidationError) String() string {
	return fmt.Sprintf("[validation-error] %s", e.Error())
}

// Add adds a new error message to the field.
func (e *ValidationError) Add(field, message string) {
	if e.Errs == nil {
		e.Errs = make(map[string][]string)
	}

	e.Errs[field] = append(e.Errs[field], message)
}

func (e *ValidationError) Merge(errs ErrorMap) {
	for field, messages := range errs {
		for _, message := range messages {
			e.Add(field, message)
		}
	}
}

// NewValidationError creates a new validation error with the provided errors if any
// NOTE: this performs a merge if more than one error map is provided
func NewValidationError(errs ...ErrorMap) *ValidationError {
	e := &ValidationError{}

	if len(errs) > 0 {
		e.Errs = errs[0]

		// Merge additional errors
		for _, err := range errs[1:] {
			e.Merge(err)
		}
	}

	return e
}
