package pipeline

import "github.com/google/uuid"

func Supplier[T any](p pipeline, s func() (T, error)) chan T {
	logger := Logger().With("Supplier", uuid.New())
	logger.Debug("Begin")

	output := make(chan T)

	count := 0

	go func() {
		defer func() {
			close(output)
			logger.Debug("End", "Count", count)
		}()

		for {
			t, err := s()
			if err != nil {
				p.CancelWithError(err)
				return
			}
			count++
			select {
			case output <- t:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	return output
}
