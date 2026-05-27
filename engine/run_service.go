package engine

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/engine/internal/state"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) CreateRun(ctx context.Context, req CreateRunRequest) (*domain.Run, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	_, spec, err := s.adapter(strings.TrimSpace(req.SceneKey))
	if err != nil {
		return nil, err
	}
	if err := validateTarget(spec, req.Target); err != nil {
		return nil, err
	}
	mode := domain.NormalizeIterationMode(req.IterationMode)
	if mode == domain.IterationModeAuto {
		if !s.allowAutoContinue {
			return nil, forbiddenError("auto continue is disabled")
		}
		if !spec.Capability.CanAutoContinue {
			return nil, forbiddenError("scene does not support auto continue")
		}
	}
	if err := validateDirectives(req.DefaultDirectives); err != nil {
		return nil, err
	}

	ruleSetSnapshot, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	now := s.now()
	run := &domain.Run{
		SceneKey:          spec.SceneKey,
		Target:            req.Target,
		Status:            domain.RunStatusPending,
		IterationMode:     mode,
		MaxDepth:          state.ResolveMaxDepth(req.MaxDepth, spec.Capability.DefaultMaxDepth, defaultMaxDepth),
		MaxVersions:       req.MaxVersions,
		Config:            domain.CloneRawMessage(req.Config),
		DefaultDirectives: domain.CloneDirectives(req.DefaultDirectives),
		RuleSetSnapshot:   ruleSetSnapshot,
		CreatedBy:         strings.TrimSpace(req.Actor),
		UpdatedBy:         strings.TrimSpace(req.Actor),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, domain.EventRunCreated, run.ID, "", req.Actor, "", nil)
	return run, nil
}

func (s *Service) StartRun(ctx context.Context, runID string) (*domain.Version, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	run, err := s.getRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	if run.Status != domain.RunStatusPending {
		return nil, conflictError("run is not pending")
	}
	plan := domain.IterationPlan{
		Source:      domain.PlanSourceInitial,
		Instruction: "Generate the first candidate version.",
		Directives:  domain.CloneDirectives(run.DefaultDirectives),
	}
	return s.startGenerate(ctx, run, nil, nil, plan, domain.ReviewPolicyRunDefault, run.CreatedBy)
}

func (s *Service) ContinueRun(ctx context.Context, req ContinueRunRequest) (*domain.Version, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	run, err := s.getRun(ctx, strings.TrimSpace(req.RunID))
	if err != nil {
		return nil, err
	}
	if run.Status.IsActive() {
		return nil, conflictError("run is active")
	}
	if run.Status == domain.RunStatusAdopted || run.Status == domain.RunStatusFailed {
		return nil, conflictError("run cannot be continued")
	}
	if req.MaxDepth > run.MaxDepth {
		run.MaxDepth = req.MaxDepth
		run.UpdatedBy = strings.TrimSpace(req.Actor)
		run.UpdatedAt = s.now()
		if err := s.store.UpdateRun(ctx, run); err != nil {
			return nil, err
		}
	}
	if req.MaxVersions > run.MaxVersions {
		run.MaxVersions = req.MaxVersions
		run.UpdatedBy = strings.TrimSpace(req.Actor)
		run.UpdatedAt = s.now()
		if err := s.store.UpdateRun(ctx, run); err != nil {
			return nil, err
		}
	}

	base, err := s.resolveBaseVersion(ctx, run, req.BaseVersionID, req.BaseVersionNo)
	if err != nil {
		return nil, err
	}
	if len(base.EffectiveContent()) == 0 {
		return nil, invalidError("base version has no effective content")
	}
	if err := state.EnsureCanCreateGeneratedVersion(run, base); err != nil {
		return nil, err
	}

	plan := req.Plan
	if plan.Source == "" {
		plan.Source = domain.PlanSourceManual
	}
	if plan.BaseVersionID == "" {
		plan.BaseVersionID = base.ID
	}
	if err := validatePlan(plan); err != nil {
		return nil, err
	}

	previousReview := reviewResultFromVersion(base)
	s.recordEvent(ctx, domain.EventManualContinue, run.ID, base.ID, req.Actor, "", mustMarshal(plan))
	return s.startGenerate(ctx, run, base, previousReview, plan, domain.ReviewPolicyRunDefault, req.Actor)
}

func (s *Service) GetRunDetail(ctx context.Context, runID string) (*RunDetail, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	run, err := s.getRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	versions, err := s.store.ListVersions(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	versionTree, err := domain.BuildVersionTree(versions)
	if err != nil {
		return nil, err
	}
	events, _ := s.store.ListEvents(ctx, run.ID)
	return &RunDetail{Run: run, Versions: versions, VersionTree: versionTree, Events: events}, nil
}

func (s *Service) ListRuns(ctx context.Context, filter ports.ListRunsFilter) ([]*domain.Run, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	return s.store.ListRuns(ctx, filter)
}
