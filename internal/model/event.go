package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Event[T any] struct {
	Payload T
	tags    map[string]struct{}
	headers map[string]string
}

const (
	HeaderID        string = "ID"
	HeaderCreatedAt string = "CREATED_AT"
)

func NewEvent[T any](payload T) *Event[T] {
	return &Event[T]{
		Payload: payload,
		tags:    map[string]struct{}{},
		headers: map[string]string{
			HeaderID:        uuid.NewString(),
			HeaderCreatedAt: time.Now().Format(time.RFC3339),
		},
	}
}

// Add a tag to event.
// Tag will be canonicalized before adding.
func (e *Event[T]) Tag(t string) {
	e.tags[strings.ToUpper(t)] = struct{}{}
}

// Get all event's tags
func (e *Event[T]) Tags() []string {
	res := make([]string, 0, len(e.tags))
	for tag := range e.tags {
		res = append(res, tag)
	}
	return res
}

// Add header to event.
// Header's key will be canonicalized before adding.
func (e *Event[T]) Header(key, value string) {
	e.headers[strings.ToUpper(key)] = value
}

// Get all event's headers.
func (e *Event[T]) Headers() map[string]string {
	res := map[string]string{}
	for k, v := range e.headers {
		res[k] = v
	}
	return res
}
