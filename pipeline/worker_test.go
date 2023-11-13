package pipeline

import (
	"fmt"
	"testing"
	"time"
)

func TestWorker(t *testing.T) {
	pipeline := Background()

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	f := func(t int) error {
		fmt.Printf("%v\n", t)
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// <-group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().SequentialWorker())

	// <-group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().ParallelWorkers())

	progress := group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().SequentialWorker().WithProgress())

	StdOutV[groupProgress[int]](pipeline, progress)

	fmt.Printf("OK\n")
}

func TestIdleWorker(t *testing.T) {
	pipeline := Background()

	data := []int{0, 1, 3}

	f := func(t int) error {
		fmt.Printf("%v\n", t)
		return nil
	}

	progress := group[int](pipeline, Throttle[int](pipeline, Slice[int](pipeline, data), 2*time.Second), f, *GroupOptions().SequentialWorker().WithIdleWorkerDuration(60 * time.Second).WithProgress())

	StdOutV[groupProgress[int]](pipeline, progress)

	fmt.Printf("OK\n")
}
