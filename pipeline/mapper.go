package pipeline

import (
	"github.com/google/uuid"
)

type MapperOpts struct {
	workerMax *int
}

func Mapper[T, R any](pipeline pipeline, input chan T, f func(T) (R, error), opts groupOptions) chan R {
	logger := Logger().With("Mapper", uuid.New())

	output := make(chan R)

	c := func(t T) error {
		r, err := f(t)
		if err != nil {
			return err
		}
		select {
		case output <- r:
			return nil
		case <-pipeline.Done():
			return nil
		}
	}

	go func() {
		defer func() {
			logger.Debug("Close output")
			close(output)
			logger.Debug("OK")
		}()
		logger.Debug("Run worker")
		<-Group[T](pipeline, input, c, opts)
		logger.Debug("Worker done")
	}()

	return output
}
