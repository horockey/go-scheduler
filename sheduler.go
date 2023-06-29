package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/horockey/go-scheduler/internal/model"
	"github.com/horockey/go-scheduler/pkg/options"
)

type Scheduler[T any] struct {
	sync.RWMutex
	nodes       []*model.Node[T]
	headChanged chan struct{}
	emitEvent   chan *model.Event[T]
	timeCh      chan time.Time
}

func NewScheduler[T any](opts ...options.Option[Scheduler[T]]) (*Scheduler[T], error) {
	s := &Scheduler[T]{
		nodes:       []*model.Node[T]{},
		headChanged: make(chan struct{}, 1),
		emitEvent:   make(chan *model.Event[T], 100),
		timeCh:      make(chan time.Time),
	}
	if err := options.ApplyOpts(s, opts...); err != nil {
		return nil, fmt.Errorf("applying opts: %w", err)
	}
	return s, nil
}

func (s *Scheduler[T]) Schedule(payload T, opts ...options.Option[model.Node[T]]) error {
	n := &model.Node[T]{
		Event: model.NewEvent[T](payload),
		At:    time.Now(),
	}
	if err := options.ApplyOpts(n, opts...); err != nil {
		return fmt.Errorf("applying opts: %w", err)
	}

	s.Lock()
	defer s.Unlock()

	s.nodes = append(s.nodes, n)
	oldHead := s.nodes[0]
	sort.Slice(s.nodes, func(i, j int) bool {
		return s.nodes[i].At.Before(s.nodes[j].At)
	})
	newHead := s.nodes[0]
	if oldHead != newHead {
		s.headChanged <- struct{}{}
	}
	return nil
}
