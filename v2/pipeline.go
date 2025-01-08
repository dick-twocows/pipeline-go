package v2

import (
	"fmt"
	"sync"
)

type Optional[T any] interface {
	Get() *T
	OK() bool
}

type optional[T any] struct {
	t  *T
	ok bool
}

func (optional *optional[T]) Get() *T {
	if !optional.ok {
		panic("t not defined")
	}
	return optional.t
}

func (optional *optional[T]) OK() bool {
	return optional.ok
}

func NewOptional[T any](t *T) *optional[T] {
	return &optional[T]{t, true}
}

func EmptyOptional[T any]() *optional[T] {
	return &optional[T]{nil, false}
}

type Control interface {
	Close() error
}

type control struct {
	control      chan struct{}
	controlClose *sync.Once
}

func (control *control) Close() error {
	control.controlClose.Do(func() {
		close(control.control)
	})

	return nil
}

func NewControl() *control {
	return &control{make(chan struct{}), &sync.Once{}}
}

type Source[T any] interface {
	Control
	Out() <-chan T
}

type source[T any] struct {
	*control
	out      chan T
	outClose *sync.Once
}

func (source *source[T]) Out() <-chan T {
	return source.out
}

func (source *source[T]) Close() error {
	source.outClose.Do(func() {
		close(source.out)

		source.control.Close()
	})

	return nil
}

func newSource[T any](size int) *source[T] {
	return &source[T]{NewControl(), make(chan T, size), &sync.Once{}}
}

func NewSource[T any]() *source[T] {
	return newSource[T](0)
}

func NewBufferedSource[T any](size int) *source[T] {
	return newSource[T](size)
}

func NewSliceSource[T any](in []T) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			source.Close()
		}()

		for _, t := range in {
			select {
			case source.out <- t:
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewSupplierSource[T any](f func() T) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			t := f()
			select {
			case source.out <- t:
			case <-source.control.control:
				return

			}
		}
	}()

	return source
}

func NewTeeSource[T any](in Source[T], count int, skew int) []*source[T] {
	outs := make([]*source[T], count)

	for i := range count {
		outs[i] = NewBufferedSource[T](skew)
	}

	go func() {
		defer func() {
			for _, out := range outs {
				out.Close()
			}
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				wg := sync.WaitGroup{}

				for i := range count {
					wg.Add(1)
					go func() {
						defer func() {
							wg.Done()
						}()

						select {
						case outs[i].out <- t:
						}
					}()
				}

				wg.Wait()
			}
		}
	}()

	return outs
}

func NewLimit[T any](in Source[T], max int) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			source.Close()
		}()

		count := 0

		for count < max {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				select {
				case source.out <- t:
				case <-source.control.control:
					return
				}

				count++
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewFilter[T any](in Source[T], f func(T) bool) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				if !f(t) {
					break
				}

				select {
				case source.out <- t:
				case <-source.control.control:
					return
				}
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewPeek[T any](in Source[T], f func(T)) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				f(t)

				select {
				case source.out <- t:
				case <-source.control.control:
					return
				}
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewMap[T, R any](in Source[T], f func(T) R) *source[R] {
	source := NewSource[R]()

	go func() {
		defer func() {
			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}
				r := f(t)

				select {
				case source.out <- r:
				case <-source.control.control:
					return
				}
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewTee[T any](in Source[T], outputs []*source[T]) *source[T] {
	source := NewSource[T]()

	go func() {
		defer func() {
			for _, output := range outputs {
				output.Close()
			}

			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				for _, output := range outputs {

					select {
					case output.out <- t:
					case <-source.control.control:
						return
					}
				}
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

func NewSplit[T any](in Source[T], predicates []func(T) bool) (*source[T], []*source[T]) {

	control := NewSource[T]()
	sources := []*source[T]{}
	filters := []*source[T]{}

	for _, predicate := range predicates {
		source := NewSource[T]()
		sources = append(sources, source)
		filters = append(filters, NewFilter[T](source, predicate))
	}

	go func() {
		defer func() {
			for _, source := range sources {
				source.Close()
			}
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				for _, source := range sources {
					select {
					case source.out <- t:
					case <-source.control.control:
						return
					}
				}
			case <-control.control.control:
				return
			}
		}
	}()

	return control, filters
}

func NewForEach[T any](in Source[T], consumer func(T)) *source[int] {
	source := NewSource[int]()

	go func() {

		count := 0

		defer func() {
			select {
			case source.out <- count:
			case <-source.control.control:
			}
			source.Close()
		}()

		for {
			select {
			case t, ok := <-in.Out():
				if !ok {
					return
				}

				consumer(t)

				count++
			case <-source.control.control:
				return
			}
		}
	}()

	return source
}

// Wait for the given terminal and return an array of optional[T].
func WaitForTerminal[T any](source Source[T]) []*optional[T] {
	result := []*optional[T]{}

	for {
		select {
		case t, ok := <-source.Out():
			if !ok {
				fmt.Printf("return result [%v]\n", result)
				return result
			}
			result = append(result, NewOptional[T](&t))
		}
	}
}

// Wait for the given terminals and return an array of array of optionals.
func WaitForTerminals[T any](sources []Source[T]) [][]*optional[T] {

	result := make([][]*optional[T], len(sources))

	wg := sync.WaitGroup{}

	for index, source := range sources {
		wg.Add(1)

		go func(index int, source Source[T]) {
			fmt.Printf("index [%v] source [%v]\n", index, source)

			defer func() {
				fmt.Printf("defer index [%v] source [%v]\n", index, source)
				wg.Done()
			}()

			result[index] = WaitForTerminal[T](source)
		}(index, source)
	}

	wg.Wait()

	fmt.Printf("result [%v]\n", len(result))
	fmt.Printf("result [%v]\n", result)

	return result
}
