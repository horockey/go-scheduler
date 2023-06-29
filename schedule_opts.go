package scheduler

import (
	"fmt"
	"time"

	"github.com/horockey/go-scheduler/internal/model"
	"github.com/horockey/go-scheduler/pkg/options"
)

// Add tag to scheduled
func Tag[T any](t string) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.Event.Tag(t)
		return nil
	}
}

func Header[T any](k, v string) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.Event.Header(k, v)
		return nil
	}
}

func After[T any](dur time.Duration) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		if dur <= 0 {
			return fmt.Errorf("duration must be positive: %d", dur)
		}
		target.At = time.Now().Add(dur)
		return nil
	}
}

func At[T any](t time.Time) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		target.At = t
		return nil
	}
}

func Every[T any](dur time.Duration) options.Option[model.Node[T]] {
	return func(target *model.Node[T]) error {
		if dur <= 0 {
			return fmt.Errorf("duration must be positive: %d", dur)
		}
		target.Every = dur
		return nil
	}
}
