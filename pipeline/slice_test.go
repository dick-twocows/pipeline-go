package pipeline

import (
	"fmt"
	"sync"
	"testing"
)

func TestSlicer(t *testing.T) {
	pipeline := Background()

	output := Slice[int](pipeline, []int{0, 1, 2, 3, 4, 5})

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case t, ok := <-output:
				if !ok {
					return
				}
				fmt.Printf("%v\n", t)
			case <-pipeline.Done():
				return
			}
		}
	}()

	wg.Wait()
}
