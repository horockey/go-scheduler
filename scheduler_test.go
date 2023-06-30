package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/horockey/go-scheduler"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var log = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: time.RFC3339,
}).With().Timestamp().Logger()

func prepareScheduler(t *testing.T) *scheduler.Scheduler[string] {
	s, err := scheduler.NewScheduler[string](
		scheduler.ErrorCB[string](func(err error) {
			if errors.Is(err, scheduler.ErrEventNotFound) {
				return
			}
			log.Error().
				Err(fmt.Errorf("scheduler: %w", err)).
				Send()
		}),
	)
	require.Nil(t, err)
	return s
}

func TestSchedule_Once_UsingAfter(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	requiredPayload := "foo"
	_, err := s.Schedule(requiredPayload, scheduler.After[string](time.Second))
	require.Nil(t, err)

	var v string
	go func() {
		e := <-s.EmitChan()
		v = e.Payload
	}()

	time.Sleep(time.Millisecond * 1200)
	require.Equal(t, requiredPayload, v)
}

func TestSchedule_Once_UsingAt(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	requiredPayload := "foo"
	_, err := s.Schedule(requiredPayload, scheduler.At[string](time.Now().Add(time.Second)))
	require.Nil(t, err)

	var v string
	go func() {
		e := <-s.EmitChan()
		v = e.Payload
	}()

	time.Sleep(time.Millisecond * 1200)
	require.Equal(t, requiredPayload, v)
}

func TestSchedule_Multiple(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	payload := "foo"
	requiredPayload := "foo|foo|foo|"
	_, err := s.Schedule(
		payload,
		scheduler.After[string](time.Second),
		scheduler.Every[string](time.Second),
	)
	require.Nil(t, err)

	var v string
	go func() {
		for {
			e := <-s.EmitChan()
			v += e.Payload + "|"
		}
	}()

	time.Sleep(time.Millisecond * 3200)
	require.Equal(t, requiredPayload, v)
}

func TestUnschedule(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	payload := "foo"
	requiredPayload := ""
	e, err := s.Schedule(
		payload,
		scheduler.After[string](time.Second*2),
	)
	require.Nil(t, err)

	var v string
	go func() {
		for {
			e := <-s.EmitChan()
			v += e.Payload + "|"
		}
	}()
	time.Sleep(time.Second)
	err = s.Unschedule(e.Headers()[scheduler.EventHeaderID()])
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 1200)
	require.Equal(t, requiredPayload, v)
}

func TestUnscheduleByTag(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	requiredOutput := "foo|"
	tag := "CUSTOM_TAG"

	_, err := s.Schedule(
		"foo",
		scheduler.After[string](time.Second*2),
	)
	require.Nil(t, err)
	_, err = s.Schedule(
		"bar",
		scheduler.After[string](time.Second*2),
		scheduler.Tag[string](tag),
	)
	require.Nil(t, err)

	var v string
	go func() {
		for {
			e := <-s.EmitChan()
			v += e.Payload + "|"
		}
	}()
	time.Sleep(time.Second)
	err = s.UnscheduleByTag(tag)
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 2200)
	require.Equal(t, requiredOutput, v)
}

func TestUnscheduleByHeader(t *testing.T) {
	s := prepareScheduler(t)

	ctx := context.TODO()
	go s.Start(ctx)
	time.Sleep(time.Millisecond * 200)

	requiredOutput := "foo|"
	headerKey := "CUSTOM_HEADER"
	headerVal := "CUSTOM_VALUE"

	_, err := s.Schedule(
		"foo",
		scheduler.After[string](time.Second*2),
	)
	require.Nil(t, err)
	_, err = s.Schedule(
		"bar",
		scheduler.After[string](time.Second*2),
		scheduler.Header[string](headerKey, headerVal),
	)
	require.Nil(t, err)

	var v string
	go func() {
		for {
			e := <-s.EmitChan()
			v += e.Payload + "|"
		}
	}()
	time.Sleep(time.Second)
	err = s.UnscheduleByHeader(headerKey, headerVal)
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 2200)
	require.Equal(t, requiredOutput, v)
}
