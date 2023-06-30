package scheduler

import (
	"fmt"

	"github.com/horockey/go-scheduler/internal/model"
	"github.com/horockey/go-scheduler/pkg/options"
)

// Set custom out chan.
// Strongly recommended to be left default!
// Change it only if you know what you are doing!
func OutChan[T any](ch chan *model.Event[T]) options.Option[Scheduler[T]] {
	return func(target *Scheduler[T]) error {
		if ch == nil {
			return fmt.Errorf("got nil channel")
		}
		close(target.emitEvent)
		target.emitEvent = ch
		return nil
	}
}

// Add custom error handler.
// By default it does noothing.
func ErrorCB[T any](cb func(error)) options.Option[Scheduler[T]] {
	return func(target *Scheduler[T]) error {
		if cb == nil {
			return fmt.Errorf("got nil callback")
		}
		target.errorCB = cb
		return nil
	}
}
