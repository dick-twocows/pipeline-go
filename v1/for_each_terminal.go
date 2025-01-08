package v1

import (
	"log/slog"

	"github.com/google/uuid"
)

type forEach[T, R any] struct {
	*terminal[T, R]
	logger   *slog.Logger
	consumer func(T) error
}

func (forEach *forEach[T, R]) Start() error {
	go func() {
		logger := forEach.logger.With(
			slog.Group(
				"ForEach",
				slog.String("UUID", uuid.New().String()),
			),
		)

		defer func() {
			logger.Debug("calling Stop()")
			forEach.Stop()
		}()

		for {
			select {
			case t, ok := <-forEach.source.Out():
				if !ok {
					logger.Debug("calling finally")
					forEach.finally()
					return
				}
				logger.Debug("consuming", slog.Any("t", t))
				if err := forEach.consumer(t); err != nil {
					return
				}
			case <-forEach.Control():
				return
			case <-forEach.stream.Control():
				return
			}
		}
	}()

	return nil
}

// Return a new for each terminal for the given stream and source[T] applying the given consumer func[T] to each source T and yielding a result R.
func NewForEach[T, R any](stream Stream, in Source[T], consumer func(T) error, finally TerminalFinally) *forEach[T, R] {
	return &forEach[T, R]{*newTerminal[T, R](stream, in, finally), consumer}
}
