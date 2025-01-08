package v3

func NewLimit[T any](pipeline *pipeline, in Source[T], max int) *source[T] {
	source := NewSource[T](pipeline, 0)

	go func() {
		defer func() {
			source.Close()
		}()

		count := 0

		for count < max {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				select {
				case source.out <- t:
				case <-source.Control():
					return
				case <-pipeline.Control():
					return
				}

				count++
			case <-source.Control():
				return
			case <-pipeline.Control():
				return
			}
		}
	}()

	return source
}
