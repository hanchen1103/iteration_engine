package engine

import (
	"context"
	"time"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

const defaultMaxDepth = 50

type Option func(*Service)

func WithAutoContinue() Option {
	return func(s *Service) {
		s.allowAutoContinue = true
	}
}

type Service struct {
	store             ports.Store
	executor          ports.JobExecutor
	registry          *ports.SceneRegistry
	clock             func() time.Time
	allowAutoContinue bool
}

func NewService(store ports.Store, executor ports.JobExecutor, registry *ports.SceneRegistry, opts ...Option) *Service {
	service := &Service{
		store:    store,
		executor: executor,
		registry: registry,
		clock:    time.Now,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(service)
		}
	}
	return service
}

func (s *Service) adapter(sceneKey string) (ports.SceneAdapter, domain.SceneSpec, error) {
	if sceneKey == "" {
		return nil, domain.SceneSpec{}, invalidError("sceneKey is required")
	}
	adapter, ok := s.registry.Get(sceneKey)
	if !ok || adapter == nil {
		return nil, domain.SceneSpec{}, notFoundError("scene adapter not found")
	}
	spec := adapter.Spec()
	if spec.SceneKey == "" {
		return nil, domain.SceneSpec{}, invalidError("scene adapter has empty sceneKey")
	}
	return adapter, spec, nil
}

func (s *Service) ready() error {
	if s == nil {
		return invalidError("service is nil")
	}
	if s.store == nil {
		return invalidError("store is nil")
	}
	if s.executor == nil {
		return invalidError("executor is nil")
	}
	if s.registry == nil {
		return invalidError("scene registry is nil")
	}
	return nil
}

func (s *Service) now() time.Time {
	if s.clock != nil {
		return s.clock().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) getRun(ctx context.Context, runID string) (*domain.Run, error) {
	if runID == "" {
		return nil, invalidError("runID is required")
	}
	return s.store.GetRun(ctx, runID)
}
