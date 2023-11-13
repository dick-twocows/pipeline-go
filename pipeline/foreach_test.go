package pipeline

import (
	"fmt"
	"testing"
)

func TestForEach(t *testing.T) {
	p := Background()

	slice := Slice[int](p, []int{0, 1, 2, 3, 4, 5})

	consumer := func(t int) error {
		fmt.Printf("%v\n", t)
		return nil
	}

	ForEach[int](p, slice, consumer)
}

func TestCount(t *testing.T) {
	p := Background()
	c, _ := Count[int](p, Slice[int](p, []int{0, 1, 2, 3, 4, 5}))
	if c.Value() != 6 {
		t.Fatal("count")
	}
}

func TestMin(t *testing.T) {
	p := Background()
	c, _ := Min[int](p, Slice[int](p, []int{0, 1, 2, 3, 4, 5}))
	if c.Value() != 0 {
		t.Fatal("min")
	}
}

func TestMax(t *testing.T) {
	p := Background()
	c, _ := Max[int](p, Slice[int](p, []int{0, 1, 2, 3, 4, 5}))
	if c.Value() != 5 {
		t.Fatal("min")
	}
}
