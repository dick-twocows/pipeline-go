package stream

import (
	"fmt"
)

func NewCount[T any](stream Stream, in Source[T]) *forEach[T, int] {
	count := 0
	consumer := func(_ T) error {
		count++
		fmt.Printf("consumer count %v\n", count)
		return nil
	}

	var forEach *forEach[T, int]
	finally := func() error {
		fmt.Printf("Count finally %v\n", count)
		return forEach.SendResult(NewOptional(count))
	}
	forEach = NewForEach[T, int](stream, in, consumer, finally)

	return forEach
}
