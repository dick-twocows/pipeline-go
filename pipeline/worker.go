/*
Worker provides a way to process a channel.

<-group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().SequentialWorker())

<-group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().ParallelWorkers())

progress := group[int](pipeline, Slice[int](pipeline, data), f, *GroupOptions().SequentialWorker().WithProgress())
*/
package pipeline

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	maxWorkersEnv        = "PIPELINE_GROUP_MAX_WORKERS"
	sequentialMaxWorkers = 1
	defaultMaxWorkers    = sequentialMaxWorkers
)

type groupOptions struct {
	MaxWorkers         int
	Progress           bool
	IdleWorkerDuration time.Duration
	OnWorkerFCallError func(error) error
}

func (o *groupOptions) SequentialWorker() *groupOptions {
	o.MaxWorkers = 1
	return o
}

func (o *groupOptions) ParallelWorkers() *groupOptions {
	o.MaxWorkers = (2 * runtime.NumCPU())
	return o
}

// Send the group/worker/t to the output channel.
func (o *groupOptions) WithProgress() *groupOptions {
	o.Progress = true
	return o
}

func (o *groupOptions) WithIdleWorkerDuration(t time.Duration) *groupOptions {
	o.IdleWorkerDuration = t
	return o
}

func GroupOptions() *groupOptions {
	return (&groupOptions{}).SequentialWorker().WithIdleWorkerDuration(1 * time.Minute)
}

// Sanitise the given GroupOptions.
func sanitiseGroupOptions(opts *groupOptions) error {
	logger := Logger().With("f", "sanitiseGroupOptions")

	logger.Debug("Before", "opts", opts)

	logger.Debug("After", "opts", opts)

	return nil
}

type groupProgress[T any] struct {
	group  uuid.UUID
	worker uuid.UUID
	t      T
}

/*
Define a group of workers.
Read the input channel calling consumer for each element.
The group will end when the input is closed or the consumer returns an error.

If WithProgress is true each worker will output each element processed and the output channel needs to be received from.
Otherwise the output channel will be closed.
*/
func group[T any](pipeline pipeline, input chan T, c func(T) error, opts groupOptions) chan groupProgress[T] {
	groupUUID := uuid.New()

	logger := Logger().With("group", groupUUID)

	logger.Debug("Start", "Options", opts)

	workersCount := atomic.Int64{}
	workersRunning := atomic.Int64{}
	workerInput := make(chan T)
	workerWaitGroup := sync.WaitGroup{}

	output := make(chan groupProgress[T])

	worker := func() {
		workerUUID := uuid.New()

		logger := logger.With("worker", workerUUID)

		logger.Debug("Start")

		defer func() {
			workersRunning.Add(-1)
			workerWaitGroup.Done()
			logger.Debug("End", "WorkersCount", workersCount.Load())
		}()

		workersCount.Add(1)
		workersRunning.Add(1)
		workerWaitGroup.Add(1)
		idleTimer := time.NewTimer(opts.IdleWorkerDuration)
		for {
			select {
			case t, ok := <-workerInput:
				logger.Debug("Received from input", "t", t, "ok", ok)
				if !ok {
					return
				}
				if err := c(t); err != nil {
					return
				}
				if opts.Progress {
					select {
					case output <- groupProgress[T]{groupUUID, workerUUID, t}:
					case <-pipeline.Done():
						return
					}
				}
			case <-idleTimer.C:
				logger.Debug("Idle duration reached")
				return
			case <-pipeline.Done():
				return
			}
			//
			if !idleTimer.Stop() {
				<-idleTimer.C
			}
			idleTimer.Reset(opts.IdleWorkerDuration)
		}
	}

	go func() {
		defer func() {
			logger.Debug("Close worker input")
			close(workerInput)
			logger.Debug("Wait for workers to finish", "workerCount", workersRunning.Load())
			workerWaitGroup.Wait()
			logger.Debug("Close output")
			close(output)
			logger.Debug("OK")
		}()

		for {
			select {
			case t, ok := <-input:
				if !ok {
					logger.Debug("Input closed")
					return
				}
				select {
				case workerInput <- t:
				case <-pipeline.Done():
					return
				default:
					logger.Debug("No worker available")
					if int(workersRunning.Load()) < opts.MaxWorkers {
						logger.Debug("New worker")
						go worker()
					}
					logger.Debug("Waiting on worker")
					select {
					case workerInput <- t:
					case <-pipeline.Done():
						return
					}
				}
			case <-pipeline.Done():
				return
			}
		}
	}()

	return output
}

func Group[T any](pipeline pipeline, input chan T, f func(T) error, opts groupOptions) chan groupProgress[T] {
	return group[T](pipeline, input, f, opts)
}
