package v3

import (
	"log/slog"
)

type Terminal[T any] interface {
	Result() []Optional[T]
	Count() int
}

type terminal[T any] struct {
	result []Optional[T]
}

func (terminal *terminal[T]) Result() []Optional[T] {
	return terminal.result
}

func (terminal *terminal[T]) Count() int {
	return len(terminal.result)
}

// func (terminal *terminal[T]) String() string {
// 	r := strings.Builder{}
// 	for i, e := range terminal.result {
// 		r.WriteString(fmt.Sprintf("%d [%v]\n", i, e))
// 	}
// 	return r.String()
// }

func NewTerminal[T any]() *terminal[T] {
	return &terminal[T]{}
}

// Block for the given terminal and return a Terminal[T].
// This will receive on the Source[T] and add each T to the Terminal[T].
// It will return when the Source[T] is closed or the Pipeline Control is closed.
func WaitForTerminal[T any](source Source[T]) Terminal[T] {
	result := NewTerminal[T]()

	pipeline := source.Pipeline()
	logger := Logger().WithGroup("WaitForTerminal").With(slog.String("ID", NewSourceID()))

	defer func() {
		logger.Debug("Returning", slog.Int("count", result.Count()))
	}()

	for {
		select {
		case t, ok := <-source.Out():
			if !ok {
				logger.Debug("In closed")
				return result
			}

			result.result = append(result.result, NewOptional[T](&t))
		case <-pipeline.Control():
			logger.Debug("Pipeline control closed")
			return result
		}
	}
}
