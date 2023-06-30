package scheduler

import (
	"fmt"
	"time"

	"github.com/horockey/go-scheduler/internal/model"
	"github.com/horockey/go-scheduler/pkg/options"
)

// Add a tag to event.
// Tag will be canonicalized before adding.
func Tag[T any](t string) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.Event.Tag(t)
		return nil
	}
}

// Add header to event.
// Header's key will be canonicalized before adding.
func Header[T any](k, v string) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.Event.Header(k, v)
		return nil
	}
}

// Emit event after given duration.
// Duration must be positive.
func After[T any](dur time.Duration) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		if dur <= 0 {
			return fmt.Errorf("duration must be positive: %d", dur)
		}
		target.At = time.Now().Add(dur)
		return nil
	}
}

// Emit event at given time.
func At[T any](t time.Time) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.At = t
		return nil
	}
}

// Continue eminitg event every given duration.
func Every[T any](dur time.Duration) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		if dur <= 0 {
			return fmt.Errorf("duration must be positive: %d", dur)
		}
		target.Every = dur
		return nil
	}
}
