package pipeline

import (
	"context"
	"log/slog"
	"os"
	"time"
)

var logger *slog.Logger

func init() {
	// buildInfo, _ := debug.ReadBuildInfo()

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	// x := slog.Group("program_info",
	// 	slog.Int("pid", os.Getpid()),
	// 	slog.String("go_version", buildInfo.GoVersion),
	// )

	logger = slog.New(
		slog.NewTextHandler(os.Stdout, opts),
	)
	// .With(x)
}

type pipeline struct {
	ctx context.Context
	// If this pipeline can be cancelled, otherwise nil.
	cancel context.CancelFunc
	// If the pipeline has an error, otherwise nil.
	// If a step within a pipeline calls CancelWithError(*error).
	err error
}

func (p *pipeline) CTX() context.Context {
	return p.ctx
}

func (p pipeline) Done() <-chan struct{} {
	return p.ctx.Done()
}

func (p pipeline) Error() error {
	return p.err
}

// Cancel the pipeline.
func (p pipeline) Cancel() {
	if p.cancel == nil {
		return
	}
	(p.cancel)()
}

// Cancel the pipeline with an error.
// We return the error passed in to allow the function to be used in a return.
//
//	if err != nil {
//		return pipeline.CancelWithError(err)
//	}
func (p *pipeline) CancelWithError(err error) error {
	p.err = err
	p.Cancel()
	return err
}

func Background() pipeline {
	p := new(pipeline)
	p.ctx = context.Background()
	return *p
}

func Using(parent context.Context) pipeline {
	p := new(pipeline)
	p.ctx = parent
	return *p
}

func WithCancel(parent context.Context) pipeline {
	p := new(pipeline)
	p.ctx, p.cancel = context.WithCancel(parent)
	return *p
}

func WithTimeout(parent context.Context, d time.Time) pipeline {
	p := new(pipeline)
	p.ctx, p.cancel = context.WithDeadline(parent, d)
	return *p
}

func Logger() *slog.Logger {
	return logger
}
