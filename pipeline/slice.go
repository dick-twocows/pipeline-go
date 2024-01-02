package pipeline

import (
	"sync/atomic"
)

type slice[T any] struct {
	source[T]
	Count atomic.Int64
}

func Slice[T any](p Pipeline, i []T) *slice[T] {

	source := &slice[T]{*NewSource[T](p, "Slice"), atomic.Int64{}}

	go func() {
		defer func() {
			close(source.Output())
			p.FlowDone(source)
		}()

		for _, t := range i {
			select {
			case source.Output() <- t:
				source.Count.Add(1)
			case <-p.CTX().Done():
				return
			}
		}
	}()

	return source
}

// Convenience function to return an empty []T.
func EmptySlice[T any](p Pipeline) *slice[T] {
	return Slice[T](p, []T{})
}
