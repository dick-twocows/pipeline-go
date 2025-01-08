package v1

import "errors"

type Optional[T any] interface {
	Get() (*T, error)
	Defined() bool
}

type optional[T any] struct {
	value   *T
	defined bool
}

func (optional *optional[T]) Get() (*T, error) {
	if optional.defined {
		return optional.value, nil
	}

	return new(T), errors.New("optional value not defined")
}

func (optional *optional[T]) Defined() bool {
	return optional.defined
}

func NewOptional[T any](t *T) optional[T] {
	return optional[T]{t, true}
}

func EmptyOptional[T any]() optional[T] {
	return optional[T]{new(T), false}
}

func OptionalOr[T any](optional Optional[T], or func() Optional[T]) Optional[T] {
	if optional.Defined() {
		return optional
	}
	return or()
}
