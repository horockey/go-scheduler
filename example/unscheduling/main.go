package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/horockey/go-scheduler"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

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
	if err != nil {
		log.Fatal().
			Err(fmt.Errorf("creating scheduler: %w", err)).
			Send()
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Start(ctx)
	}()

	const (
		tag         string = "AWESOME_TAG"
		headerKey   string = "AWESOME_HEADER"
		headerValue string = "AWESOME_VALUE"
	)
	s.Schedule(
		"event_with_tag",
		scheduler.After[string](time.Second*3),
		scheduler.Every[string](time.Second),
		scheduler.Tag[string](tag),
	)
	s.Schedule(
		"event_with_tag_and_header",
		scheduler.After[string](time.Second*3),
		scheduler.Every[string](time.Second),
		scheduler.Tag[string](tag),
		scheduler.Header[string](headerKey, headerValue),
	)
	s.Schedule(
		"event_with_header",
		scheduler.After[string](time.Second*3),
		scheduler.Every[string](time.Second),
		scheduler.Header[string](headerKey, headerValue),
	)
	s.Schedule(
		"simple_event",
		scheduler.After[string](time.Second*3),
		scheduler.Every[string](time.Second),
	)

	go func() {
		time.Sleep(time.Second * 5)
		s.UnscheduleByTag(tag)
		log.Info().Msg("events with tag unscheduled")
		time.Sleep(time.Second * 3)
		s.UnscheduleByHeader(headerKey, headerValue)
		log.Info().Msg("events with header unscheduled")
	}()

	go func() {
		log.Info().Msg("start listening to scheduler")
		for e := range s.EmitChan() {
			log.Info().
				Str("event_id", e.Headers()[scheduler.EventHeaderID()]).
				Msg(e.Payload)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	sig := <-sigChan

	cancel()
	log.Warn().
		Str("signal", sig.String()).
		Msg("terminating process...")
	wg.Wait()
}
