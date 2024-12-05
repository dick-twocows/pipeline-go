package main

import (
	"fmt"
	"testing"
)

func TestForEach(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	consumer := func(t int) error {
		fmt.Printf("%v\n", t)
		return nil
	}

	forEach := NewIgnoreResultTerminal[int](stream, slice, consumer, NilTerminalFinally)
	forEach.Start()
	WaitForTerminalResult[int, struct{}](forEach)
}

func TestConsumeToStdOut(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	forEach := NewConsumeToStdOut[int](stream, slice)
	forEach.Start()
	WaitForTerminalResult[int, struct{}](forEach)
}

func TestCount(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	count := NewCount[int](stream, slice)
	count.Start()
	fmt.Printf("count result %v\n", WaitForTerminalResult[int, int](count))
}

func TestMin(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	terminal := NewMin[int](stream, slice)
	terminal.Start()
	fmt.Printf("min result %v\n", WaitForTerminalResult[int, int](terminal))
}

func TestMax(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	terminal := NewMax[int](stream, slice)
	terminal.Start()
	fmt.Printf("max result %v\n", WaitForTerminalResult[int, int](terminal))
}
