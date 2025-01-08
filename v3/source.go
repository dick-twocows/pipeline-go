package v3

import (
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
)

type Source[T any] interface {
	Control
	ID() string
	Pipeline() Pipeline
	Out() chan T
}

type source[T any] struct {
	*control
	id       string
	pipeline Pipeline
	out      chan T
	outClose *sync.Once
}

func (source *source[T]) ID() string {
	return source.id
}

func (source *source[T]) Pipeline() Pipeline {
	return source.pipeline
}

func (source *source[T]) Out() chan T {
	return source.out
}

// Safely close the out channel and then the control channel.
func (source *source[T]) Close() error {
	source.outClose.Do(func() {
		close(source.out)

		source.control.Close()
	})

	return nil
}

// Return a new source.
func NewSource[T any](pipeline Pipeline, size int) *source[T] {
	return &source[T]{NewControl(), NewSourceID(), pipeline, make(chan T, size), &sync.Once{}}
}

var sourceID = atomic.Int64{}

// Return a new Source ID.
// Currently this increments an atomic Int64.
func NewSourceID() string {
	return strconv.Itoa(int(sourceID.Add(1)))
}

func NewSourceLogger[T any](source Source[T], group string) *slog.Logger {
	return Logger().WithGroup(group).With(slog.String("ID", source.ID()))
}
