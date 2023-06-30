package scheduler

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/horockey/go-scheduler/internal/model"
	"github.com/horockey/go-scheduler/pkg/options"
)

var (
	ErrNotRunning          = fmt.Errorf("scheduller is not running")
	ErrEventNotFound       = fmt.Errorf("event with given id not found")
	ErrUnexpectedEmptyList = fmt.Errorf("unexpected empty shedule event list")
	ErrEventWithNoIDHeader = fmt.Errorf("got event with no ID header. It will be generated")
)

type Scheduler[T any] struct {
	mu sync.RWMutex

	isRunning bool

	nodes       []*model.Node[T]
	headChanged chan struct{}
	emitEvent   chan *model.Event[T]
	timeCh      <-chan time.Time

	errorCB func(error)
}

// Create new scheduler with given opts.
func NewScheduler[T any](opts ...options.Option[Scheduler[T]]) (*Scheduler[T], error) {
	s := &Scheduler[T]{
		nodes:       []*model.Node[T]{},
		headChanged: make(chan struct{}, 2),
		emitEvent:   make(chan *model.Event[T], 100),
		timeCh:      make(<-chan time.Time),
		errorCB:     func(err error) {},
	}
	if err := options.ApplyOpts(s, opts...); err != nil {
		return nil, fmt.Errorf("applying opts: %w", err)
	}
	return s, nil
}

// Start scheduler.
// To stop it, given context should be canceled.
func (s *Scheduler[T]) Start(ctx context.Context) error {
	updTimeChan := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if len(s.nodes) == 0 {
			s.timeCh = make(<-chan time.Time)
			return
		}
		dur := time.Until(s.nodes[0].At)
		if dur < 0 {
			dur = 0
		}
		s.timeCh = time.After(dur)
	}

	emitEvent := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if len(s.nodes) == 0 {
			s.errorCB(ErrUnexpectedEmptyList)
			return
		}
		n := s.nodes[0]
		s.emitEvent <- n.Event

		s.removeNode(0)

		if n.Every > 0 {
			s.approveIdHeader(n)
			n.At = time.Now().Add(n.Every)
			s.addNode(n)
		}
	}

	s.setRunning(true)

	for {
		select {
		case <-s.timeCh:
			emitEvent()
		case <-s.headChanged:
			updTimeChan()
		case <-ctx.Done():
			s.setRunning(false)
			close(s.headChanged)
			close(s.emitEvent)
			return fmt.Errorf("context err: %w", ctx.Err())
		}
	}
}

// Schedule new event.
// Scheduler must be started to call this method properly.
func (s *Scheduler[T]) Schedule(payload T, opts ...options.Option[model.Node[T]]) (*model.Event[T], error) {
	e := model.NewEvent[T](payload)
	n := &model.Node[T]{
		Event: e,
		At:    time.Now(),
	}
	if err := options.ApplyOpts(n, opts...); err != nil {
		return nil, fmt.Errorf("applying opts: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isRunning {
		return nil, ErrNotRunning
	}

	s.addNode(n)

	return e, nil
}

// Unschedule scheduled event by id.
// Scheduler must be started to call this method properly.
func (s *Scheduler[T]) Unschedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for idx, node := range s.nodes {
		eventId := s.approveIdHeader(node)
		if eventId != id {
			continue
		}

		s.removeNode(idx)
		return nil
	}
	return ErrEventNotFound
}

// Get channel, that emits scheduled events.
func (s *Scheduler[T]) EmitChan() <-chan *model.Event[T] {
	return s.emitEvent
}

func (s *Scheduler[T]) addNode(node *model.Node[T]) {
	s.nodes = append(s.nodes, node)
	oldHead := s.nodes[0]
	sort.Slice(s.nodes, func(i, j int) bool {
		return s.nodes[i].At.Before(s.nodes[j].At)
	})
	newHead := s.nodes[0]
	if oldHead != newHead || len(s.nodes) == 1 {
		s.headChanged <- struct{}{}
	}
}

func (s *Scheduler[T]) removeNode(idx int) {
	switch len(s.nodes) {
	case 1:
		s.nodes = []*model.Node[T]{}
	default:
		s.nodes = append(s.nodes[:idx], s.nodes[idx+1:]...)
	}

	if idx == 0 {
		s.headChanged <- struct{}{}
	}
}

func (s *Scheduler[T]) approveIdHeader(node *model.Node[T]) string {
	id, ok := node.Event.Headers()[model.HeaderID]
	if !ok {
		s.errorCB(ErrEventWithNoIDHeader)
		id = uuid.NewString()
		node.Event.Header(model.HeaderID, id)
	}
	return id
}

func (s *Scheduler[T]) setRunning(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = v
}
