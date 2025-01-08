package v3

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Return a new source T which will timeout if a T is not received within the given timeout.
// The given timeout effects the time waiting to receive T on the in channel, it does not included any other time taken.
func NewReceiveTimeout[T any](pipeline *pipeline, in Source[T], timeout time.Duration) *source[T] {
	source := NewSource[T](pipeline, 0)

	logger := pipeline.Logger().With(slog.String("ReceiveTimeout", uuid.NewString()))

	go func() {
		defer func() {
			logger.Debug("closing source")
			source.Close()
		}()

		count := 0

		for true {
			select {
			case t, ok := <-in.Out():
				if !ok {
					logger.Debug("in closed")
					return
				}
				logger.Debug("received from in", slog.Any("t", t))

				select {
				case source.out <- t:
					logger.Debug("sent to out")
				case <-source.Control():
					return
				case <-pipeline.Control():
					return
				}

				count++
			case <-time.After(timeout):
				logger.Debug("timeout waiting to receive t from in")
				return
			case <-source.Control():
				return
			case <-pipeline.Control():
				return
			}
		}
	}()

	logger.Debug("returning new receive timeout", slog.String("in", fmt.Sprintf("%T", in)), slog.Any("out", source))
	return source
}

// Return a new source T which will timeout if a T is not sent within the given timeout duration.
// The given timeout effects the time waiting to send T on the out channel, it does not included any other time taken.
func NewSendTimeout[T any](pipeline *pipeline, in Source[T], timeout time.Duration) *source[T] {
	source := NewSource[T](pipeline, 0)

	logger := pipeline.Logger().With(slog.String("ReceiveTimeout", uuid.NewString()))

	go func() {
		defer func() {
			logger.Debug("closing source")
			source.Close()
		}()

		count := 0

		for true {
			select {
			case t, ok := <-in.Out():
				if !ok {
					logger.Debug("in closed")
					return
				}
				logger.Debug("received T from in", slog.Any("t", t))

				select {
				case source.out <- t:
					logger.Debug("sent to out")
				case <-time.After(timeout):
					logger.Debug("timeout waiting to send T to out")
					return
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

	logger.Debug("returning new receive timeout", slog.String("in", fmt.Sprintf("%T", in)), slog.Any("out", source))
	return source
}

// Return a new source T which will run for the given duration.
func NewPeriodIntermediate[T any](pipeline *pipeline, in Source[T], timeout time.Duration) *source[T] {
	source := NewSource[T](pipeline, 0)

	logger := pipeline.Logger().With(slog.String("PeriodIntermediate", uuid.NewString()))

	go func() {
		period := time.After(timeout)
		count := 0

		defer func() {
			logger.Debug("closing source", slog.Int("count", count))
			source.Close()
		}()

		for true {
			select {
			case t, ok := <-in.Out():
				if !ok {
					logger.Debug("in closed")
					return
				}
				logger.Debug("received T from in", slog.Any("t", t))

				select {
				case source.out <- t:
					logger.Debug("sent T to out")
				case <-period:
					logger.Debug("timeout waiting to send T to out")
					return
				case <-source.Control():
					return
				case <-pipeline.Control():
					return
				}

				count++
			case <-period:
				logger.Debug("timeout waiting to receive T from in")
				return
			case <-source.Control():
				return
			case <-pipeline.Control():
				return
			}
		}
	}()

	logger.Debug("returning new period intermediate", slog.String("in", fmt.Sprintf("%T", in)), slog.Any("out", source))
	return source
}
