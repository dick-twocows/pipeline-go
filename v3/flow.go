package v3

import "log/slog"

type Sequence[S, T any] interface {
	S() S
	T() T
}

type sequence[S, T any] struct {
	s S
	t T
}

func (sequence *sequence[S, T]) S() S {
	return sequence.s
}

func (sequence sequence[S, T]) T() T {
	return sequence.t
}

// Return a new intermediate which wraps T in a sequence.
func NewSequenceIntermediate[S, T any](pipeline Pipeline, in Source[T], s func(t T) (sequence[S, T], bool, error)) *source[sequence[S, T]] {
	return NewMapperIntermediate[T, sequence[S, T]](pipeline, in, s)
}

func NewIntSequenceAddIntermediate[T any](pipeline *pipeline, in Source[T]) *source[sequence[int, T]] {
	i := 0

	f := func(t T) (sequence[int, T], bool, error) {
		i++
		return sequence[int, T]{i, t}, true, nil
	}

	return NewSequenceIntermediate(pipeline, in, f)
}

type node[S, T any] struct {
	s        Sequence[S, T]
	previous *node[S, T]
	next     *node[S, T]
}

const (
	LESS = iota
	EQUAL
	GREATER
)

func NewSequenceRemove[S, T any](pipeline *pipeline, in Source[Sequence[S, T]], from S, c func(S, S) (bool, error)) *source[T] {
	out := NewSource[T](pipeline, 0)

	var head *node[S, T]

	// h	nil
	// h	node->nil
	// h	node->node->nil
	go func() {
		select {
		case t, ok := <-in.Out():
			if !ok {
				return
			}

			if head == nil {
				// No
				head = &node[S, T]{t, nil, nil}
			} else {
				cN := head

				for {
					// func c will return true if <, false if >, and error is equal.
					b, err := c(cN.s.S(), t.S())
					if err != nil {
						return
					}

					if b {

					} else {
						cN = cN.next
					}
				}
			}

		case <-in.Control():
			return
		case <-pipeline.Control():
			return
		}
	}()

	return out
}

// Convenience function to return a nested source.
func NewNestedSource[T any](pipeline Pipeline, in ...Source[T]) Source[Source[T]] {
	return NewSliceSource(pipeline, in)
}

func NewNestedIntermediate[T any](pipeline Pipeline, in Source[Source[T]]) Source[T] {
	out := NewSource[T](pipeline, 0)

	logger := NewSourceLogger(out, "NestedIntermediate")
	logger.Debug("Created", slog.Any("out", out))

	f := func(t Source[T]) error {
		logger.Debug("Consuming nested source", slog.Any("Source[T]", t))

		forEach := NewForEachTerminal(pipeline, t, func(t T) error {
			select {
			case out.Out() <- t:
			case <-out.Control():
			case <-pipeline.Control():
			}
			return nil
		})

		terminal := WaitForTerminal[int](forEach)

		logger.Debug("Consumed nested source", slog.Any("Source[T]", t), slog.Any("Terminal", terminal))
		return nil
	}

	go func() {
		defer func() {
			out.Close()
		}()

		logger.Debug("Completed nested intermediate", slog.Any("Terminal", WaitForTerminal(NewForEachTerminal(pipeline, in, f))))
	}()

	return out
}
