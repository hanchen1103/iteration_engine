package memory

import (
	"context"
	"fmt"
	"sort"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Store) CreateRun(ctx context.Context, run *domain.Run) error {
	_ = ctx
	if run == nil {
		return invalidError("run is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if run.ID == "" {
		s.nextRunID++
		run.ID = fmt.Sprintf("run_%d", s.nextRunID)
	}
	if _, ok := s.runs[run.ID]; ok {
		return conflictError("run already exists")
	}
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) UpdateRun(ctx context.Context, run *domain.Run) error {
	_ = ctx
	if run == nil || run.ID == "" {
		return invalidError("run id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[run.ID]; !ok {
		return notFoundError("run not found")
	}
	s.runs[run.ID] = cloneRun(run)
	return nil
}

func (s *Store) GetRun(ctx context.Context, runID string) (*domain.Run, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[runID]
	if !ok {
		return nil, notFoundError("run not found")
	}
	return cloneRun(run), nil
}

func (s *Store) ListRuns(ctx context.Context, filter ports.ListRunsFilter) ([]*domain.Run, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]*domain.Run, 0, len(s.runs))
	for _, run := range s.runs {
		if filter.SceneKey != "" && run.SceneKey != filter.SceneKey {
			continue
		}
		if filter.Status != "" && run.Status != filter.Status {
			continue
		}
		if filter.Target.Type != "" && run.Target.Type != filter.Target.Type {
			continue
		}
		if filter.Target.ID != "" && run.Target.ID != filter.Target.ID {
			continue
		}
		items = append(items, cloneRun(run))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []*domain.Run{}, nil
	}
	limit := filter.Limit
	if limit <= 0 || offset+limit > len(items) {
		limit = len(items) - offset
	}
	return items[offset : offset+limit], nil
}
