package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

// We use slog for logging, https://betterstack.com/community/guides/logging/logging-in-go/#getting-started-with-slog

var logger slog.Logger

func init() {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	logger = *slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
}

// Optional

type Optional[T any] interface {
	Get() (T, error)
	OK() bool
}

type optional[T any] struct {
	value T
	ok    bool
}

func (optional *optional[T]) Get() (T, error) {
	if optional.ok {
		return optional.value, nil
	}

	return *new(T), errors.New("Optional value not ok")
}

func (optional *optional[T]) OK() bool {
	return optional.ok
}

func NewOptional[T any](t T) optional[T] {
	return optional[T]{t, true}
}

func EmptyOptional[T any]() optional[T] {
	return optional[T]{*new(T), false}
}

func OptionalOr[T any](optional Optional[T], or func() Optional[T]) Optional[T] {
	if optional.OK() {
		return optional
	}
	return or()
}

// Control

type Control interface {
	Control() <-chan struct{}
	Start() error
	Stop() error
}

type control struct {
	control      chan struct{}
	controlClose *sync.Once
}

// Return the control channel as receive only.
func (control *control) Control() <-chan struct{} {
	return control.control
}

func (control *control) Start() error {
	return nil
}

// Stop, which safely closes the control channel.
func (control *control) Stop() error {
	control.controlClose.Do(func() {
		close(control.control)
	})
	return nil
}

func newControl() control {
	return control{make(chan struct{}), &sync.Once{}}
}

// Stream

type Stream interface {
	Control
	Logger() *slog.Logger
}

type stream struct {
	control
	logger *slog.Logger
}

func (stream *stream) Logger() *slog.Logger {
	return stream.logger
}

func NewStream() *stream {
	return &stream{
		newControl(),
		logger.With(
			slog.Group(
				"Stream",
				slog.String("UUID", uuid.New().String()),
			),
		),
	}
}

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

// Sources

// Slice Source, which outputs the given []T slice.
type slice[T any] struct {
	source[T]
	in []T
}

func (slice *slice[T]) Start() error {
	go func() {
		defer slice.Stop()

		for _, t := range slice.in {
			select {
			case slice.out <- t:
			case <-slice.Control():
				return
			case <-slice.stream.Control():
				return
			}
		}
	}()

	return nil
}

func NewSlice[T any](stream Stream, in []T) *slice[T] {
	return &slice[T]{newSource[T](stream), in}
}

// Intermediate

type Intermediate[T, R any] interface {
	Control
	Source() Source[T]
	Out() chan R
}

// Terminal
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

// Block until the given terminal yields a result.
// We do not check the source control because it will be closed.
func WaitForTerminalResult[T, R any](terminal Terminal[T, R]) []optional[R] {
	result := []optional[R]{}
	for {
		select {
		case t, ok := <-terminal.Result():
			if !ok {
				return result
			}
			result = append(result, t)
		case <-terminal.Control():
			return result
		case <-terminal.Source().Stream().Control():
			return result
		}
	}

}

// Terminals

type forEach[T, R any] struct {
	terminal[T, R]
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

// Return a new for each terminal which yields no result.
func NewIgnoreResultTerminal[T any](stream Stream, in Source[T], consumer func(T) error, finally TerminalFinally) *forEach[T, struct{}] {
	return &forEach[T, struct{}]{*newTerminal[T, struct{}](stream, in, finally), consumer}
}

func NewConsumeToStdOut[T any](stream Stream, in Source[T]) *forEach[T, struct{}] {
	consumer := func(t T) error {
		fmt.Printf("%v\n", t)
		return nil
	}

	return NewIgnoreResultTerminal(stream, in, consumer, NilTerminalFinally)
}

func NewCount[T any](stream Stream, in Source[T]) *forEach[T, int] {
	count := 0
	consumer := func(_ T) error {
		count++
		fmt.Printf("consumer count %v\n", count)
		return nil
	}

	var forEach *forEach[T, int]
	finally := func() error {
		fmt.Printf("Count finally %v\n", count)
		return forEach.SendResult(NewOptional(count))
	}
	forEach = NewForEach[T, int](stream, in, consumer, finally)

	return forEach
}

// Consume the given source T returning the extrenum T using the given comnparator.
// The extrenum is determined by the given comparator.
// Returns a single optional T result.
func NewExtrenum[T any](stream Stream, in Source[T], comparator func(extrenum T, t T) (bool, error)) *forEach[T, T] {
	var (
		count    int
		extrenum T
	)

	consumer := func(t T) error {
		count++

		if count == 1 {
			extrenum = t

			return nil
		}

		update, err := comparator(extrenum, t)

		if err != nil {
			return err
		}

		if update {
			extrenum = t
		}

		return nil
	}

	var forEach *forEach[T, T]

	finally := func() error {
		if count == 0 {
			return forEach.SendResult(EmptyOptional[T]())
		}
		return forEach.SendResult(NewOptional(extrenum))
	}

	forEach = NewForEach[T, T](stream, in, consumer, finally)

	return forEach
}

func NewMin[T constraints.Ordered](stream Stream, in Source[T]) *forEach[T, T] {
	return NewExtrenum(
		stream,
		in,
		func(extrenum T, t T) (bool, error) {
			return (t < extrenum), nil
		},
	)
}

func NewMax[T constraints.Ordered](stream Stream, in Source[T]) *forEach[T, T] {
	return NewExtrenum(
		stream,
		in,
		func(extrenum T, t T) (bool, error) {
			return (t > extrenum), nil
		},
	)
}
