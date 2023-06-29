package options

import "fmt"

type Option[T any] func(target *T) error

func ApplyOpts[T any](target *T, opts ...Option[T]) error {
	for idx, opt := range opts {
		if opt == nil {
			return fmt.Errorf("got nil opt at position %d", idx)
		}
		if err := opt(target); err != nil {
			return fmt.Errorf("applying opt at position %d: %w", idx, err)
		}
	}
	return nil
}
