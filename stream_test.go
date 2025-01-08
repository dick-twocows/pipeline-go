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
	WaitForTerminal[int, struct{}](forEach)
}

func TestConsumeToStdOut(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	forEach := NewConsumeToStdOut[int](stream, slice)
	forEach.Start()
	WaitForTerminal[int, struct{}](forEach)
}

func TestCount(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	count := NewCount[int](stream, slice)
	count.Start()
	fmt.Printf("count result %v\n", WaitForTerminal[int, int](count))
}

func TestMin(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	terminal := NewMin[int](stream, slice)
	terminal.Start()
	fmt.Printf("min result %v\n", WaitForTerminal[int, int](terminal))
}

func TestMax(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	terminal := NewMax[int](stream, slice)
	terminal.Start()
	fmt.Printf("max result %v\n", WaitForTerminal[int, int](terminal))
}

func TestMapToString(t *testing.T) {
	stream := NewStream()

	slice := NewSlice(stream, []int{1, 2, 3, 4, 5})
	slice.Start()

	m := NewMapToString[int](stream, slice)
	m.Start()

	p := NewPeekToStdOut[string](stream, m)
	p.Start()

	terminal := NewConsumeToStdOut[string](stream, p)
	terminal.Start()
	fmt.Printf("result %v\n", WaitForTerminal[string, struct{}](terminal))
}
