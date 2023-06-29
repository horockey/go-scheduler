package model

import "time"

type Node[T any] struct {
	Event *Event[T]
	At    time.Time
	Every time.Duration
}
