package pipeline

import "fmt"

func ForEach[T any](pipeline pipeline, input chan T, consumer func(t T) error) error {
	for {
		select {
		case t, ok := <-input:
			if !ok {
				return nil
			}
			if err := consumer(t); err != nil {
				return err
			}
		case <-pipeline.Done():
			return nil
		}
	}
}

func StdOutV[T any](pipeline pipeline, input chan T) error {
	return ForEach[T](pipeline, input, func(t T) error { fmt.Printf("%v\n", t); return nil })
}

// Drop everything from the input, AKA /dev/null, blackhole, etc...
func Drop[T any](pipeline pipeline, input chan T) error {
	return ForEach[T](pipeline, input, func(t T) error { return nil })
}
