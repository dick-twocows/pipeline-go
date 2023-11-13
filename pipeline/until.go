package pipeline

import "github.com/google/uuid"

func Until[T any](pipeline pipeline, input chan T, p func(T) (bool, error)) chan T {
	logger := Logger().With("Until", uuid.New())
	logger.Debug("Begin")

	output := make(chan T)
	tCount := 0

	go func() {
		defer func() {
			close(output)
			logger.Debug("End", "TCount", tCount)
		}()

		for {
			select {
			case t, ok := <-input:
				if !ok {
					return
				}
				tCount++
				b, err := p(t)
				if err != nil {
					pipeline.CancelWithError(err)
					return
				}
				if b {
					return
				}
				select {
				case output <- t:
				case <-pipeline.Done():
					return
				}
			case <-pipeline.Done():
				return
			}
		}
	}()

	return output
}

func Limit[T any](pipeline pipeline, input chan T, max int) chan T {
	count := 0
	p := func(t T) (bool, error) {
		count++
		return !(count <= max), nil
	}
	return Until[T](pipeline, input, p)
}
