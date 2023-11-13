package pipeline

type optional[T any] struct {
	value   T
	defined bool
}

func (o optional[T]) Value() T {
	return o.value
}

func (o optional[T]) ValueOK() (T, bool) {
	return o.value, o.defined
}

func Optional[T any](value T) optional[T] {
	return optional[T]{value, true}
}

func Empty[T any]() optional[T] {
	return *new(optional[T])
}

// Convenience to return an optional[T] where the defined is based on a loop calculation.
// See return from Count/Min/Max.
func Raw[T any](value T, defined bool) optional[T] {
	return optional[T]{value, defined}
}
