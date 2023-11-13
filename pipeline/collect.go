package pipeline

import (
	"cmp"
	"fmt"
	"sync"
)

func Count[T any](pipeline pipeline, input chan T) (optional[int], error) {
	count := 0
	err := ForEach[T](pipeline, input, func(t T) error { count++; return nil })
	return Optional[int](count), err
}

func Min[T cmp.Ordered](pipeline pipeline, input chan T) (optional[T], error) {
	var min T
	defined := false
	consumer := func(t T) error {
		if !defined {
			defined = true
			min = t
			return nil
		}
		if t < min {
			min = t
		}
		return nil
	}

	return optional[T]{min, defined}, ForEach[T](pipeline, input, consumer)
}

func Max[T cmp.Ordered](pipeline pipeline, input chan T) (optional[T], error) {
	var max T
	defined := false
	consumer := func(t T) error {
		if !defined {
			defined = true
			max = t
			return nil
		}
		if t > max {
			max = t
		}
		return nil
	}
	return Raw[T](max, defined), ForEach[T](pipeline, input, consumer)
}

func ToSlice[T any](pipeline pipeline, input chan T) (optional[[]T], error) {
	result := []T{}
	defined := false
	counter := func(t T) error {
		defined = true
		result = append(result, t)
		return nil
	}
	return Raw[[]T](result, defined), ForEach[T](pipeline, input, counter)
}

// Collect T in a map[K][]T which is created by applying mapper(T)(K,error) to produce a map key and adding T to the value []T.
// If the mapper returns an error the collection is stopped.
func GroupBy[T any, K cmp.Ordered](pipeline pipeline, input chan T, mapper func(t T) (K, error)) (optional[map[K][]T], error) {
	result := make(map[K][]T)
	defined := false

	adder := func(t T) error {
		k, err := mapper(t)
		if err != nil {
			return err
		}
		defined = true
		if v, ok := result[k]; ok {
			result[k] = append(v, t)
		} else {
			result[k] = []T{t}
		}
		return nil
	}

	return Raw[map[K][]T](result, defined), ForEach[T](pipeline, input, adder)
}

type to[T, A any] interface {
	accumulate(T) error
	result() (A, error)
}

func To[T, A, R any](
	pipeline pipeline,
	input chan T,
	supplier func() (A, error),
	accumulator func(A, T) (A, error),
	combiner func(A, A) (A, error),
	finisher func(A) (R, error),
) (optional[R], error) {

	result, err := supplier()
	if err != nil {
		return Empty[R](), err
	}

	defined := false

	workerCount := 0
	workerMax := 4

	workerInput := make(chan T)

	resultInput := make(chan A)

	rwg := sync.WaitGroup{}
	rwg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("closing result combiner\n")
			rwg.Done()
		}()

		for {
			select {
			case a, ok := <-resultInput:
				fmt.Printf("receive result combiner a[%v] ok[%v]\n", a, ok)
				if !ok {
					return
				}
				defined = true
				r, err := combiner(result, a)
				if err != nil {
					return
				}
				result = r
			case <-pipeline.Done():
				return
			}
		}
	}()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer func() {
			fmt.Printf("close shim\n")
			close(workerInput)

			wg.Done()
		}()

		for {
			select {
			case t, ok := <-input:
				fmt.Printf("receive [%v] [%v]\n", t, ok)
				if !ok {
					return
				}

				select {
				case workerInput <- t: // There is a free worker.
				case <-pipeline.Done():
					return
				default: // Try and create a new worker.
					if workerCount == workerMax {
						// We are at max workers so block for free worker.
						select {
						case workerInput <- t:
						case <-pipeline.Done():
							return
						}
					} else {
						// Create a new worker.
						workerCount++
						wg.Add(1)
						go func() {
							fmt.Printf("worker [%v]\n", workerCount)
							a, err := supplier()
							fmt.Printf("supplied a[%v]\n", a)
							if err != nil {
								pipeline.Cancel()
								return
							}

							defer func() {
								fmt.Printf("closing worker returning a[%v]\n", a)
								resultInput <- a
								fmt.Printf("done worker\n")
								wg.Done()
							}()

							for {
								select {
								case t, ok := <-workerInput:
									fmt.Printf("worker receive from shim t[%v] ok[%v]\n", t, ok)
									if !ok {
										return
									}
									r, err := accumulator(a, t)
									fmt.Printf("accumulated a[%v] t[%v] r[%v]\n", a, t, r)
									if err != nil {
										pipeline.Cancel()
										return
									}
									a = r
								case <-pipeline.Done():
									return
								}
							}
						}()
						// This will be received by the worker we just created or a free worker.
						select {
						case workerInput <- t:
						case <-pipeline.Done():
							return
						}
					}
				}

			case <-pipeline.Done():
				return
			}
		}
	}()

	// Wait for the core and workers to finish
	wg.Wait()
	fmt.Printf("workers complete\n")

	close(resultInput)

	rwg.Wait()
	fmt.Printf("result combiner complete\n")

	r, err := finisher(result)

	return Raw[R](r, defined), err
}
