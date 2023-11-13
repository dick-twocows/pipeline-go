package pipeline

func Slice[T any](p pipeline, input []T) chan T {
	output := make(chan T)

	go func() {
		defer close(output)

		for _, t := range input {
			select {
			case output <- t:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	return output
}
