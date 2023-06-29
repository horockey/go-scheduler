package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/horockey/go-scheduler"
	"github.com/horockey/go-scheduler/internal/model"
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
	err = s.Unschedule(e.Headers()[model.HeaderID])
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 1200)
	require.Equal(t, requiredPayload, v)
}