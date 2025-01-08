package v1

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Apply a terminal action to the given Source, e.g. count.
type Terminal[T, R any] interface {
	Control
	Logger() *slog.Logger
	Pipeline() Pipeline
	Source() Source[T]
	Result() chan optional[R]
}

type TerminalFinally func() error

func NilTerminalFinally() error {
	return nil
}

type terminal[T, R any] struct {
	//
	*control
	// slog
	logger *slog.Logger
	// Pipeline this terminal belongs to.
	pipeline Pipeline
	// Source this terminal will consume.
	source Source[T]
	// The func called when the source has been consumed.
	// Only called when the source is consumed.
	finally TerminalFinally
	// The result channel.
	result chan optional[R]
	// Ensures the result channel is closed once.
	resultClose *sync.Once
}

func (terminal *terminal[T, R]) Stop() error {
	terminal.resultClose.Do(func() {
		close(terminal.result)
	})

	terminal.Stop()

	return nil
}

func (terminal *terminal[T, R]) Logger() *slog.Logger {
	return terminal.logger
}

func (terminal *terminal[T, R]) Pipeline() Pipeline {
	return terminal.pipeline
}

func (terminal *terminal[T, R]) Source() Source[T] {
	return terminal.source
}

func (terminal *terminal[T, R]) Result() <-chan optional[R] {
	return terminal.result
}

func (terminal *terminal[T, R]) sendResult(r optional[R]) error {
	select {
	case terminal.result <- r:
		terminal.logger.Debug("sent result", slog.Any("r", r))
		return nil
	case <-terminal.control.Control():
		return errors.New("terminal control closed")
	case <-terminal.pipeline.Control():
		return errors.New("terminal pipeline control closed")
	}
}

func newTerminal[T, R any](pipeline Pipeline, source Source[T], finally TerminalFinally) *terminal[T, R] {
	control := NewControl()

	logger := control.Logger().With(
		slog.Group(
			"Terminal",
			slog.String("UUID", uuid.New().String()),
		),
	)

	return &terminal[T, R]{
		control,
		logger,
		pipeline,
		source,
		finally,
		make(chan optional[R]),
		&sync.Once{},
	}
}
