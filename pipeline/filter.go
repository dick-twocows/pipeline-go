package pipeline

import "github.com/google/uuid"

func Filter[T any](pipeline pipeline, input chan T, predicate func(t T) (bool, error)) chan T {
	logger := Logger().With("Filter", uuid.New())
	logger.Debug("Begin")

	output := make(chan T)

	go func() {
		tCount := 0
		tPermit := 0

		defer func() {
			close(output)
			logger.Debug("End", "Count", tCount, "Permit", tPermit)
		}()

		for {
			select {
			case t, ok := <-input:
				if !ok {
					return
				}
				tCount++
				permit, err := predicate(t)
				if err != nil {
					pipeline.CancelWithError(err)
					return
				}
				if !permit {
					break
				}
				tPermit++
				select {
				case output <- t:
				case <-pipeline.ctx.Done():
					return
				}
			case <-pipeline.ctx.Done():
				return
			}
		}
	}()

	return output
}
