package memory

import (
	"context"
	"fmt"
	"sort"

	"github.com/hanchen1103/iteration_engine/domain"
)

func (s *Store) RecordEvent(ctx context.Context, event *domain.Event) error {
	_ = ctx
	if event == nil {
		return invalidError("event is nil")
	}
	if event.RunID == "" {
		return invalidError("event run id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[event.RunID]; !ok {
		return notFoundError("run not found")
	}
	if event.ID == "" {
		s.nextEventID++
		event.ID = fmt.Sprintf("evt_%d", s.nextEventID)
	}
	s.events[event.RunID] = append(s.events[event.RunID], cloneEvent(event))
	return nil
}

func (s *Store) ListEvents(ctx context.Context, runID string) ([]*domain.Event, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := s.events[runID]
	items := make([]*domain.Event, 0, len(events))
	for _, event := range events {
		items = append(items, cloneEvent(event))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}
