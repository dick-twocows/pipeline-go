package pipeline

import (
	"fmt"
	"testing"
)

func TestFilter(t *testing.T) {
	pipeline := Background()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	filter := Filter[int](pipeline, slice, func(t int) (bool, error) { return t%2 == 0, nil })

	ForEach[int](pipeline, filter, func(t int) error { fmt.Printf("%v\n", t); return nil })
}
