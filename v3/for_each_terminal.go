package v3

import (
	"fmt"
	"log/slog"
)

func TStdOutConsumer[T any](t T) error {
	fmt.Printf("%v\n", t)
	return nil
}

// Consume each in T using the given consumer function, returning a count >=0.
func NewForEachTerminal[T any](pipeline Pipeline, in Source[T], consumer func(T) error) Source[int] {
	// The returned out which returns a count.
	out := NewSource[int](pipeline, 0)

	logger := NewSourceLogger(out, "ForEachTerminal")

	logger.Debug("Created", slog.Any("Source", in), slog.Any("consumer", consumer))

	go func() {
		count := 0

		defer func() {
			logger.Debug("Sending count", slog.Int("count", count))

			select {
			// Send the count.
			case out.Out() <- count:
				logger.Debug("Sent count")
			case <-out.Control():
				logger.Debug("Out control closed")
			case <-out.Pipeline().Control():
				logger.Debug("Pipeline out closed")
			}

			out.Close()
		}()

		for {
			select {
			// Read a T.
			case t, ok := <-in.Out():
				if !ok {
					logger.Debug("In out closed")
					return
				}
				logger.Debug("Received t", slog.Any("t", t))

				// Call the consumer and check the returned error.
				if err := consumer(t); err != nil {
					logger.Warn("Error consuming t", slog.Any("error", err), slog.Any("t", t))
					return
				}

				// Update count.
				count++
			case <-out.Control():
				logger.Debug("Out control closed")
				return
			case <-out.Pipeline().Control():
				logger.Debug("Pipeline control closed")
				return
			}
		}
	}()

	logger.Debug("Returning")

	return out
}

// Consume the in whilst discarding the T's and return a count, AKA >/dev/null
func NewNullTerminal[T any](pipeline *pipeline, in Source[T]) Source[int] {
	return NewForEachTerminal[T](pipeline, in, func(_ T) error { return nil })
}

func NewCountTerminal[T any](pipeline *pipeline, in Source[T]) Source[int] {
	return NewNullTerminal[T](pipeline, in)
}

func NewExtrenumTerminal[T any](pipeline *pipeline, in Source[T], extrenum func(T, T) (T, bool, error)) Source[*optional[T]] {
	count := 0
	var r T
	source := NewSource[*optional[T]](pipeline, 0)

	go func() {
		defer func() {
			var result *optional[T]
			if count == 0 {
				result = EmptyOptional[T]()
			} else {
				result = NewOptional(&r)
			}
			select {
			case source.Out() <- result:
			case <-source.Control():
			case <-source.Pipeline().Control():
			}

			source.Close()
		}()

		for {
			select {
			// Read a T.
			case t, ok := <-in.Out():
				if !ok {
					// In source is closed so return.
					return
				}

				if count == 0 {
					r = t
					continue
				}

				e, changed, err := extrenum(r, t)
				if err != nil {
					return
				}

				if changed {
					r = e
				}

				// Update count.
				count++
			case <-source.Control():
				return
			case <-source.Pipeline().Control():
				return
			}
		}
	}()

	return source
}
