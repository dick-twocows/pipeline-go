package v1

import "testing"

func TestPipeline(t *testing.T) {
	pipeline := NewPipeline()
	pipeline.Start()
	pipeline.Stop()

	source := NewSource[int](pipeline)
	source.Start()
	source.Stop()

	slice := NewSlice[int](pipeline, []int{1, 2, 3, 4, 5})
	slice.Start()
	slice.Stop()

}
