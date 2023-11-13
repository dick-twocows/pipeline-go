package pipeline

import "testing"

func TestSupplier(t *testing.T) {
	pipeline := Background()

	count := 0
	s := func() (int, error) {
		count++
		return count, nil
	}
	supplier := Supplier[int](pipeline, s)

	limit := Limit[int](pipeline, supplier, 5)

	StdOutV[int](pipeline, limit)
}
