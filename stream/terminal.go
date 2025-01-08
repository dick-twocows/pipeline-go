package stream

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Apply a terminal action to the given Source, e.g. count.
type Terminal[T, R any] interface {
	Control
	Source() Source[T]
	Result() chan optional[R]
}

type TerminalFinally func() error

func NilTerminalFinally() error {
	return nil
}

type terminal[T, R any] struct {
	// Stream this terminal belongs to.
	stream Stream
	//
	control
	// slog
	logger *slog.Logger
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

	return nil
}

func (terminal *terminal[T, R]) Source() Source[T] {
	return terminal.source
}

func (terminal *terminal[T, R]) Result() chan optional[R] {
	return terminal.result
}

func (terminal *terminal[T, R]) SendResult(result optional[R]) error {
	select {
	case terminal.result <- result:
		fmt.Printf("terminal send result [%v]\n", result)
	case <-terminal.control.Control():
		return errors.New("Terminal control closed")
	case <-terminal.stream.Control():
		return errors.New("Terminal stream control closed")
	}
	return nil
}

func newTerminal[T, R any](stream Stream, source Source[T], finally TerminalFinally) *terminal[T, R] {
	return &terminal[T, R]{
		stream,
		newControl(),
		stream.Logger().With(
			slog.Group(
				"Terminal",
				slog.String("UUID", uuid.New().String()),
			),
		),
		source,
		finally,
		make(chan optional[R]),
		&sync.Once{},
	}
}
