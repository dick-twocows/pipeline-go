package v1

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Controls

type Control interface {
	Control() <-chan struct{}
	Start() error
	Stop() error
	Logger() *slog.Logger
}

var ControlStartNoOverride = errors.New("Control Start() no override")

// Implement the Control interface.
type control struct {
	control      chan struct{}
	controlClose sync.Once
	logger       *slog.Logger
}

// Return the control channel as receive only.
// Currently this is used in a select{...} to check for closing.
func (control *control) Control() <-chan struct{} {
	return control.control
}

// Returns an error as Start() needs to be overridden.
func (control *control) Start() error {
	control.logger.Debug("starting control")
	control.logger.Error("oops", slog.Any("Error", ControlStartNoOverride))
	return ControlStartNoOverride
}

// Safely closes (idempotent) the control channel.
func (control *control) Stop() error {
	control.logger.Debug("stopping control")

	control.controlClose.Do(func() {
		control.logger.Debug("closing control channel")
		close(control.control)
	})

	return nil
}

// Return the logger.
func (control *control) Logger() *slog.Logger {
	return control.logger
}

// Return a pointer to a new control
func NewControl() *control {
	control := &control{
		make(chan struct{}),
		sync.Once{},
		logger.With(
			"Control",
			slog.String("UUID", uuid.NewString()),
		),
	}

	control.Logger().Debug("created control")

	return control
}
