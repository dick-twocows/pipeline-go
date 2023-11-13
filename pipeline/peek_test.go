package pipeline

import (
	"fmt"
	"testing"
)

func TestPeek(t *testing.T) {
	pipeline := Background()

	defer pipeline.Cancel()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4, 5})

	consumer := func(t int) error {
		fmt.Printf("%v\n", t)
		return nil
	}

	peek := Peek[int](pipeline, slice, consumer)

	Drop[int](pipeline, peek)
}
