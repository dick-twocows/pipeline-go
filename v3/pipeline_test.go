package v3

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/exp/rand"
)

var ajSlice = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

var slice09 = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

var slice0To3 = []int{0, 1, 2, 3}

func stdOutConsumer[T any](t T) error {
	fmt.Printf("t [%v]\n", t)
	return nil
}

func TestPipeline(t *testing.T) {
	pipeline := NewPipeline()

	slice := NewSliceSource[int](pipeline, []int{1, 2, 3, 4, 5})

	consumer := func(t int) error {
		fmt.Printf("for each %v\n", t)

		return nil
	}

	forEach := NewForEachTerminal[int](pipeline, slice, consumer)

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach).Result()[0].Get())
}

func TestSupplier(t *testing.T) {
	pipeline := NewPipeline()

	supplier := NewSupplierSource[int](pipeline, func() (int, error) { return rand.Intn(9), nil })

	limit := NewLimit[int](pipeline, supplier, 10)

	forEach := NewForEachTerminal[int](pipeline, limit, func(t int) error { fmt.Printf("for each %v\n", t); return nil })

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach).Result()[0].Get())
}

func TestTimeout(t *testing.T) {
	pipeline := NewPipeline()

	f := func() (int, error) {
		s := rand.Intn(5)
		fmt.Printf("sleeping [%v]\n", s)
		time.Sleep(time.Second * time.Duration(s))
		return s, nil
	}

	supplier := NewSupplierSource[int](pipeline, f)

	// slice := NewSliceSource(pipeline, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	limit := NewLimit[int](pipeline, supplier, 10)

	timeout := NewReceiveTimeout(pipeline, limit, time.Duration(time.Second*2))

	forEach := NewForEachTerminal[int](pipeline, timeout, func(t int) error { fmt.Printf("for each %v\n", t); return nil })

	fmt.Printf("count %v\n", *(WaitForTerminal[int](forEach).Result()[0].Get()))
}

func TestPeriodIntermediate(t *testing.T) {
	pipeline := NewPipeline()

	f := func() (int, error) {
		s := rand.Intn(9)
		return s, nil
	}

	supplier := NewSupplierSource[int](pipeline, f)

	timeout := NewPeriodIntermediate(pipeline, supplier, time.Duration(time.Second*1))

	forEach := NewCountTerminal(pipeline, timeout)

	fmt.Printf("count %v\n", *(WaitForTerminal[int](forEach).Result()[0].Get()))
}

func TestFlowSequenceIntermediate(t *testing.T) {
	pipeline := NewPipeline()

	slice := NewSliceSource[string](pipeline, ajSlice)

	sequence := NewIntSequenceAddIntermediate(pipeline, slice)

	forEach := NewForEachTerminal(pipeline, sequence, TStdOutConsumer)

	fmt.Printf("terminal [%v]\n", WaitForTerminal[int](forEach))
}

func TestMapIntermediate(t *testing.T) {
	pipeline := NewPipeline()

	slice := NewSliceSource[int](pipeline, slice09)

	f := func(t int) (int, bool, error) {
		return t * 10, true, nil
	}
	mapper := NewMapperIntermediate[int, int](pipeline, slice, f)

	forEach := NewForEachTerminal(pipeline, mapper, TStdOutConsumer)

	fmt.Printf("terminal [%v]\n", WaitForTerminal[int](forEach))
}

func TestNested(t *testing.T) {
	pipeline := NewPipeline()

	slice1 := NewSliceSource[int](pipeline, slice0To3)
	slice2 := NewSliceSource[int](pipeline, slice0To3)

	slice := NewSliceSource[Source[int]](pipeline, []Source[int]{slice1, slice2})

	nested := NewNestedIntermediate(pipeline, slice)

	forEach := NewForEachTerminal[int](pipeline, nested, stdOutConsumer)

	fmt.Printf("%v\n", *WaitForTerminal[int](forEach).Result()[0].Get())
}
