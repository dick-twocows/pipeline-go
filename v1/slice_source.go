package v1

import (
	"log/slog"

	"github.com/google/uuid"
)

type slice[T any] struct {
	*source[T]
	logger *slog.Logger
	in     []T
}

func (slice *slice[T]) Start() error {

	go func() {
		slice.Logger().Debug("started")

		defer func() {
			slice.logger.Debug("calling slice stop")
			slice.Stop()
		}()

		for _, t := range slice.in {
			select {
			case slice.out <- t:
				slice.logger.Debug("sent t", slog.Any("t", t))
			case <-slice.Control():
				slice.logger.Debug("slice control channel closed")
				return
			case <-slice.pipeline.Control():
				slice.logger.Debug("slice pipeline control channel closed")
				return
			}
		}
	}()

	return nil
}

func NewSlice[T any](pipeline Pipeline, in []T) *slice[T] {
	source := NewSource[T](pipeline)

	logger := source.Logger().With(slog.Group("SliceSource", slog.String("UUID", uuid.NewString())))

	slice := &slice[T]{source, logger, in}

	slice.Logger().Debug("created slice source")

	return slice
}
