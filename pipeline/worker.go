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

If WithProgress is true each worker will output each element processed (the output channel needs to be received from).
Effectively a passthru.
*/
func WorkerGroup[T any](pipeline Pipeline, input Source[T], c func(Pipeline, T) error, opts groupOptions) *source[groupProgress[T]] {
	// groupUUID := uuid.New()

	// logger := Logger().With("group", groupUUID)

	// logger.Debug("Start", "Options", opts)

	workersCount := atomic.Int64{}
	workersRunning := atomic.Int64{}

	sliceInput := make(chan T)
	sliceInputCount := 0

	workerWaitGroup := sync.WaitGroup{}

	workerGroup := NewSource[groupProgress[T]](pipeline, "WorkerGroup")

	slice := func() {
		sliceUUID := uuid.New()

		sliceInputCount := 0

		logger := workerGroup.Logger().With("Slice", sliceUUID)

		logger.Debug("Start")

		defer func() {
			workersRunning.Add(-1)
			workerWaitGroup.Done()
			logger.Debug("Metrics", "SliceInputCount", sliceInputCount)
		}()

		idleTimer := time.NewTimer(opts.IdleWorkerDuration)
		for {
			select {
			case t, ok := <-sliceInput:
				sliceInputCount++
				logger.Debug("Received from input", "t", t, "ok", ok)
				if !ok {
					return
				}
				if err := c(pipeline, t); err != nil {
					return
				}
				if opts.Progress {
					select {
					case workerGroup.Output() <- groupProgress[T]{workerGroup.ID(), sliceUUID, t}:
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
			close(sliceInput)

			workerWaitGroup.Wait()

			close(workerGroup.Output())
			workerGroup.Metrics("SliceCount", workersCount.Load(), "SliceInputCount", sliceInputCount)
			pipeline.FlowDone(workerGroup)
		}()

		for {
			select {
			case t, ok := <-input.Output():
				if !ok {
					workerGroup.Logger().Debug("Input closed")
					return
				}

				select {
				case sliceInput <- t:
					sliceInputCount++
				case <-pipeline.Done():
					workerGroup.Logger().Debug("Pipeline done")
					return
				default:
					workerGroup.Logger().Debug("No slice available")
					if int(workersRunning.Load()) < opts.MaxWorkers {
						logger.Debug("New slice")
						workersCount.Add(1)
						workersRunning.Add(1)
						workerWaitGroup.Add(1)
						go slice()
					}
					logger.Debug("Waiting on slice")
					select {
					case sliceInput <- t:
						sliceInputCount++
					case <-pipeline.Done():
						return
					}
				}
			case <-pipeline.Done():
				return
			}
		}
	}()

	return workerGroup
}
