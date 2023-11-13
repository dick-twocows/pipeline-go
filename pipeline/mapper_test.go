package pipeline

import (
	"fmt"
	"testing"
)

func TestMapper(t *testing.T) {
	pipeline := Background()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	mapper := Mapper[int](pipeline, slice, func(t int) (int, error) { return t, nil }, *GroupOptions().ParallelWorkers())

	count, err := Count[int](pipeline, mapper)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("count [%v]\n", count)
}
