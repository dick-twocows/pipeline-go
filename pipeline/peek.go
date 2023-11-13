package pipeline

import (
	"time"

	"github.com/google/uuid"
)

// Peek the input channel using the given consumer.
// If the consumer returns an error return early.
func Peek[T any](pipeline pipeline, input chan T, consumer func(t T) error) chan T {
	logger := Logger().With("Peek", uuid.New())

	logger.Debug("Begin")

	output := make(chan T)

	go func() {
		defer func() {
			close(output)
			logger.Debug("End")
		}()

		for {
			select {
			case t, ok := <-input:
				if !ok {
					return
				}
				if err := consumer(t); err != nil {
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

// Throttle each T for the given time duration.
func Throttle[T any](pipeline pipeline, input chan T, d time.Duration) chan T {
	return Peek[T](pipeline, input, func(t T) error { time.Sleep(d); return nil })
}
