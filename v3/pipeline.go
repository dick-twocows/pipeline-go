package v3

import (
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type Pipeline interface {
	Control
	Logger() *slog.Logger
}

type pipeline struct {
	*control
	logger *slog.Logger
}

func (pipeline *pipeline) Logger() *slog.Logger {
	return pipeline.logger
}

func NewPipeline() *pipeline {
	return &pipeline{
		NewControl(),
		Logger().With(slog.String("Pipeline", uuid.NewString()))}
}

var logger *slog.Logger

func init() {
	logger = slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)
}

func Logger() *slog.Logger {
	return logger
}
