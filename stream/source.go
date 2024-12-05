package stream

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Source, is a supplier of T values via a channel.
// Usually used to construct a more specific source.
type Source[T any] interface {
	Control
	Stream() Stream
	Out() chan T
}

type source[T any] struct {
	stream Stream
	control
	logger   *slog.Logger
	out      chan T
	closeOut *sync.Once
}

func (source *source[T]) Stream() Stream {
	return source.stream
}

// Stop the source.
// - safely close the out channel
// - stop the control
// If you override this function make sure it is safe!
func (source *source[T]) Stop() error {
	source.logger.Debug("stopping")

	source.closeOut.Do(func() {
		source.logger.Debug("closing source out")
		close(source.out)
	})

	source.logger.Debug("calling source control stop")
	source.control.Stop()

	return nil
}

func (source *source[T]) Out() chan T {
	return source.out
}

func newSource[T any](stream Stream) source[T] {
	return source[T]{
		stream,
		newControl(),
		stream.Logger().With(
			slog.Group(
				"Source",
				slog.String("UUID", uuid.New().String()),
			),
		),
		make(chan T),
		&sync.Once{},
	}
}
