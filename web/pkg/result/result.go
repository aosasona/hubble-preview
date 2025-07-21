package result

type Result[Value any] interface {
	IsOk() bool
	IsErr() bool
	Unwrap(Value) Value
	Err() error
}

// Ok represents a successful result.
type ok[Value any] struct {
	value Value
}

func (o ok[Value]) IsOk() bool {
	return true
}

func (o ok[Value]) IsErr() bool {
	return false
}

func (o ok[Value]) Value() Value {
	return o.value
}

func (o ok[Value]) Unwrap(fallback Value) Value {
	if o.IsErr() {
		return fallback
	}

	return o.value
}

func (o ok[_]) Err() error {
	return nil
}

// Err represents an error result.
type err[Value any] struct {
	value Value
	err   error
}

func (e err[_]) IsOk() bool {
	return false
}

func (e err[_]) IsErr() bool {
	return true
}

func (e err[_]) Err() error {
	return e.err
}

func (e err[Value]) Unwrap(fallback Value) Value {
	if e.IsOk() {
		return fallback
	}

	return e.value
}

// Ok returns a successful result.
func Ok[Value any](value Value) Result[Value] {
	return ok[Value]{value}
}

// Err returns an error result.
func Err[Value any](e error) Result[Value] {
	return err[Value]{err: e}
}

var (
	_ Result[any] = ok[any]{}
	_ Result[any] = err[any]{}
)
