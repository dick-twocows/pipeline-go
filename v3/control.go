package v3

import "sync"

type Control interface {
	Control() <-chan struct{}
	Close() error
}

type control struct {
	control      chan struct{}
	controlClose *sync.Once
}

func (control *control) Control() <-chan struct{} {
	return control.control
}

// Safely close the control channel.
func (control *control) Close() error {
	control.controlClose.Do(func() {
		close(control.control)
	})

	return nil
}

func NewControl() *control {
	return &control{make(chan struct{}), &sync.Once{}}
}
