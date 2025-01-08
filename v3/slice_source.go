package v3

import (
	"log/slog"
)

func NewSliceSource[T any](pipeline Pipeline, in []T) Source[T] {
	out := NewSource[T](pipeline, 0)

	logger := NewSourceLogger(out, "SliceSource")

	logger.Debug("Created", slog.Int("count", len(in)))

	go func() {
		defer func() {
			logger.Debug("Closing out")
			out.Close()
		}()

		for i, t := range in {
			logger.Debug("Iterator", slog.Int("i", i), slog.Any("t", t))

			select {
			case out.Out() <- t:
			case <-out.Control():
				logger.Debug("Out control closed")
				return
			case <-pipeline.Control():
				logger.Debug("Pipeline control closed")
				return
			}
		}
	}()

	logger.Debug("Returning", slog.Any("out", out))
	return out
}
