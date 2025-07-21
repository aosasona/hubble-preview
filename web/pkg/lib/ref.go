package lib

import (
	"fmt"
)

type refType interface {
	string | int | int32 | int64 | float32 | float64 | bool
}

// Ref returns a pointer to a string.
func Ref[T refType](s T) *T { return &s }

// Deref returns the value of a reference or alternatively, a fallback value.
func Deref[T refType](s *T, fallback T) T {
	if s == nil {
		return fallback
	}

	return *s
}

// MustDeref returns the value of a reference or panics if the reference is nil.
func MustDeref[T refType](s *T) T {
	if s == nil {
		panic(fmt.Sprintf("reference is nil, expected %T", s))
	}

	return *s
}
