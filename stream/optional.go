package stream

import "errors"

type Optional[T any] interface {
	Get() (T, error)
	OK() bool
}

type optional[T any] struct {
	value T
	ok    bool
}

func (optional *optional[T]) Get() (T, error) {
	if optional.ok {
		return optional.value, nil
	}

	return *new(T), errors.New("Optional value not ok")
}

func (optional *optional[T]) OK() bool {
	return optional.ok
}

func NewOptional[T any](t T) optional[T] {
	return optional[T]{t, true}
}

func EmptyOptional[T any]() optional[T] {
	return optional[T]{*new(T), false}
}

func OptionalOr[T any](optional Optional[T], or func() Optional[T]) Optional[T] {
	if optional.OK() {
		return optional
	}
	return or()
}
