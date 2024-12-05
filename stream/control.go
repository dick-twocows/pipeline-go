package stream

import "sync"

type Control interface {
	Control() <-chan struct{}
	Start() error
	Stop() error
}

type control struct {
	control      chan struct{}
	controlClose *sync.Once
}

// Return the control channel as receive only.
func (control *control) Control() <-chan struct{} {
	return control.control
}

func (control *control) Start() error {
	return nil
}

// Stop, which safely closes the control channel.
func (control *control) Stop() error {
	control.controlClose.Do(func() {
		close(control.control)
	})
	return nil
}

func newControl() control {
	return control{make(chan struct{}), &sync.Once{}}
}
