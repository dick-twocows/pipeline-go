package v3

func NewSupplierSource[T any](pipeline Pipeline, f func() (T, error)) *source[T] {
	source := NewSource[T](pipeline, 0)

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			t, err := f()
			if err != nil {
				return
			}

			select {
			case source.out <- t:
			case <-source.Control():
				return
			case <-pipeline.Control():
				return
			}
		}
	}()

	return source
}
