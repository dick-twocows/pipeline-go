package v2

import (
	"fmt"
	"sync"
	"testing"

	"golang.org/x/exp/rand"
)

func TestPipeline(t *testing.T) {
	slice := NewSliceSource[int]([]int{1, 2, 3, 4, 5})

	p1 := NewPeek[int](slice, func(t int) { fmt.Printf("p1 %v\n", t) })

	m := NewMap[int, int](p1, func(t int) int { return t * t })

	forEach := NewForEach[int](m, func(t int) { fmt.Printf("for each %v\n", t) })

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach)[0].Get())
}

func TestSupplier(t *testing.T) {
	supplier := NewSupplierSource[int](func() int { return rand.Intn(9) })

	limit := NewLimit[int](supplier, 10)

	forEach := NewForEach[int](limit, func(t int) { fmt.Printf("for each %v\n", t) })

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach)[0].Get())
}

func TestFilter(t *testing.T) {
	supplier := NewSliceSource[int]([]int{1, 2, 3, 4, 5})

	filter := NewFilter[int](supplier, func(t int) bool { return t%2 == 0 })

	forEach := NewForEach[int](filter, func(t int) { fmt.Printf("for each %v\n", t) })

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach)[0].Get())
}

func TestTee(t *testing.T) {
	supplier := NewSliceSource[int]([]int{1, 2, 3, 4, 5})

	tee := NewTeeSource[int](supplier, 2, 0)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		forEach := NewForEach[int](tee[0], func(t int) { fmt.Printf("0 for each %v\n", t) })

		fmt.Printf("0 %v\n", *WaitForTerminal[int](forEach)[0].Get())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		forEach := NewForEach[int](tee[1], func(t int) { fmt.Printf("1 for each %v\n", t) })

		fmt.Printf("1 %v\n", *WaitForTerminal[int](forEach)[0].Get())
	}()

	wg.Wait()
	fmt.Println("OK")
}

func TestTee2(t *testing.T) {
	supplier := NewSliceSource[int]([]int{1, 2, 3, 4, 5})

	tee := NewTeeSource[int](supplier, 2, 0)

	consumer := func(t int) { fmt.Printf("0 for each %v\n", t) }

	result := WaitForTerminals[int]([]Source[int]{NewForEach[int](tee[0], consumer), NewForEach[int](tee[1], consumer)})

	fmt.Printf("%v\n", len(result))

	fmt.Printf("%v\n", len(result[0]))
	// fmt.Printf("%v\n", *result[0][0].Get())

	fmt.Printf("%v\n", len(result[1]))
	// fmt.Printf("%v\n", *result[1][0].Get())

	fmt.Println("OK")
}

func TestTee3(t *testing.T) {
	supplier := NewSliceSource[int]([]int{1, 2, 3, 4, 5})

	tee := NewTeeSource[int](supplier, 10, 0)

	result := WaitForTerminals[int]([]Source[int]{tee[0], tee[1]})
	fmt.Printf("%v\n", len(result))

	fmt.Println("OK")
}
