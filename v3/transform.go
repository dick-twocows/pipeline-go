package v3

// Return a new intermediate which wraps T in a sequence.
func NewMapperIntermediate[T, R any](pipeline Pipeline, in Source[T], f func(T) (R, bool, error)) *source[R] {
	source := NewSource[R](pipeline, 0)

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				r, ok, err := f(t)
				if err != nil {
					return
				}
				if !ok {
					continue
				}

				select {
				case source.Out() <- r:
				case <-in.Control():
					return
				case <-pipeline.Control():
					return
				}
			case <-in.Control():
				return
			case <-pipeline.Control():
				return
			}
		}
	}()

	return source
}
