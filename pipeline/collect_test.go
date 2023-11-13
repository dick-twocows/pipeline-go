package pipeline

import (
	"fmt"
	"testing"
	"time"
)

func TestToSlice(t *testing.T) {
	pipeline := Background()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	result, err := ToSlice[int](pipeline, slice)

	fmt.Printf("%v\n%v\n", result, err)
}

func TestGroupBy(t *testing.T) {
	pipeline := Background()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4})

	result, err := GroupBy[int, string](pipeline, slice, func(t int) (string, error) {
		if t%2 == 0 {
			return "even", nil
		} else {
			return "odd", nil
		}
	})

	fmt.Printf("%v\n%v\n", result, err)
}

func TestTo(t *testing.T) {
	pipeline := Background()

	slice := Slice[int](pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})

	supplier := func() ([]int, error) {
		fmt.Printf("supplier\n")
		return []int{}, nil
	}

	accumulator := func(c []int, t int) ([]int, error) {
		fmt.Printf("accumulator [%v] [%v]\n", c, t)
		time.Sleep(500 * time.Millisecond)
		return append(c, t), nil
	}

	combiner := func(a []int, b []int) ([]int, error) {
		fmt.Printf("combiner [%v] [%v]\n", a, b)

		if len(a) > 0 && len(b) == 0 {
			return a, nil
		}

		if len(a) == 0 && len(b) > 0 {
			return b, nil
		}

		return append(a, b...), nil
	}

	finisher := func(a []int) ([]int, error) {
		fmt.Printf("finisher\n")
		r := []int{}
		for _, t := range a {
			r = append(r, t)
		}
		return r, nil
	}

	result, err := To[int, []int, []int](pipeline, slice, supplier, accumulator, combiner, finisher)

	fmt.Printf("%v\n%v\n", result, err)
}
