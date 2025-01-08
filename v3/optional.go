package v3

import "fmt"

type Optional[T any] interface {
	Get() *T
	OK() bool
}

type optional[T any] struct {
	t  *T
	ok bool
}

func (optional *optional[T]) Get() *T {
	if !optional.ok {
		panic("t not defined")
	}
	return optional.t
}

func (optional *optional[T]) OK() bool {
	return optional.ok
}

func (optional *optional[T]) String() string {
	return fmt.Sprintf("%t %[1]T[%[1]v]", optional.ok, *optional.t)
}

// Return a new Optional[T] with t=*T and ok=true.
func NewOptional[T any](t *T) *optional[T] {
	return &optional[T]{t, true}
}

// Return a new empty optional T, with *T == nil and ok == false.
// If you are regularly returning an empty optional T consider creating it once and returning it rather than calling this function repeatedly...
func EmptyOptional[T any]() *optional[T] {
	return &optional[T]{nil, false}
}

var (
	EmptyBoolOptional   = EmptyOptional[bool]()
	EmptyIntOptional    = EmptyOptional[int]()
	EmptyStringOptional = EmptyOptional[string]()
)
