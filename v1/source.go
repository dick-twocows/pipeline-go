package v1

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

//  Sources

type Source[T any] interface {
	Control
	Logger() *slog.Logger
	Pipeline() Pipeline
	Out() chan T
}

type source[T any] struct {
	*control
	pipeline Pipeline
	logger   *slog.Logger
	out      chan T
	closeOut *sync.Once
}

func (source *source[T]) Pipeline() Pipeline {
	return source.pipeline
}

func (source *source[T]) Logger() *slog.Logger {
	return source.logger
}

func NewSource[T any](pipeline Pipeline) *source[T] {
	control := NewControl()
	logger := control.Logger().With("Source", slog.String("UUID", uuid.NewString()))

	source := &source[T]{
		control,
		pipeline,
		logger,
		make(chan T),
		&sync.Once{},
	}

	source.Logger().Debug("created source")

	return source
}
