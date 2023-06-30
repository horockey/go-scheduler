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

	s.Schedule(
		"message to be shown",
		scheduler.After[string](time.Second*5),
		scheduler.Every[string](time.Second),
	)
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
