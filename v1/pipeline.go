package v1

import (
	"log/slog"
	"os"

	"github.com/google/uuid"
)

// In Go, there are six categories of value that are passed by reference rather than by value. These are pointers, slices, maps, channels, interfaces and functions.

// TODO: Custom attrributes for error
// See https://stackoverflow.com/questions/77304845/how-to-log-errors-with-log-slog
var logger *slog.Logger

func init() {
	handlerOptions := &slog.HandlerOptions{
		// AddSource: true,
		Level: slog.LevelDebug,
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions)).With(
		slog.Group(
			"Pipeline",
			slog.String("UUID", uuid.NewString()),
		),
	)
}

// Stream

type Pipeline interface {
	Control
	Logger() *slog.Logger
}

type pipeline struct {
	*control
	logger *slog.Logger
}

func (stream *pipeline) Start() error {
	stream.logger.Debug("Starting stream")
	return nil
}

func (stream *pipeline) Stop() error {
	stream.logger.Debug("Stopping stream")
	return nil
}

func (stream *pipeline) Logger() *slog.Logger {
	return stream.logger
}

func NewPipeline() *pipeline {
	control := NewControl()

	stream := &pipeline{
		control,
		control.logger.With(
			"Pipeline",
			slog.String("UUID", uuid.NewString()),
		),
	}

	stream.Logger().Debug("created pipeline")

	return stream
}
