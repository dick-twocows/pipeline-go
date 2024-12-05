package stream

import (
	"log/slog"
	"os"

	"github.com/google/uuid"
)

var logger slog.Logger

func init() {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	logger = *slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
}

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
