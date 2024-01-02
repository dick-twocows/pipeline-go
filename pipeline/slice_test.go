package pipeline

import (
	"testing"
)

func TestEmpty(t *testing.T) {
	pipeline := NewPipeline()
	defer pipeline.Cancel()

	slice := EmptySlice[int](pipeline)
	<-ForEachStdOutV[int](pipeline, slice).Output()

	StdOutConsumer[int](pipeline, int(slice.Count.Load()))
}

func TestSlicer(t *testing.T) {
	pipeline := NewPipeline()
	defer pipeline.Cancel()

	slice := Slice[int](pipeline, []int{0, 1, 2, 3, 4})

	<-ForEachStdOutV[int](pipeline, slice).Output()

	StdOutConsumer[int](pipeline, int(slice.Count.Load()))
}
