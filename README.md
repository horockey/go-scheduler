# go-scheduler

![CI](https://github.com/horockey/go-scheduler/actions/workflows/main.yaml/badge.svg)

[![Go Reference](https://pkg.go.dev/badge/github.com/horockey/go-scheduler.svg)](https://pkg.go.dev/github.com/horockey/go-scheduler)
![License](https://img.shields.io/github/license/horockey/go-scheduler)


Library for delayed sending message to channel with ability to cancel sending.

## Installation

`go get github.com/horockey/go-scheduler@latest`

## Examples

See [/example](./example/) directory

## Events

Library is working with entities called *events*.
Event is simply payload, wrapped with tags and headers:

```go
type Event[T any] struct {
	Payload T
	tags    map[string]struct{}
	headers map[string]string
}
```

Creating *event* via `NewEvent()` method sets 2 default headers to it:
* `CREATED_AT` header is set to `time.Now()`
* `ID` header is set to `uuid.NewString()`

You are free to add any tags and headers to your events.

### Note
Adding new tags and headers to *event* keep in mind, that they will be canonicalized (`strings.ToUpper()`) before actual addition to event
