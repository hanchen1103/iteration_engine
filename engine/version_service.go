package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/engine/internal/state"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) EditVersion(ctx context.Context, req EditVersionRequest) (*domain.Version, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	if len(req.Content) == 0 {
		return nil, invalidError("content is required")
	}
	run, version, _, err := s.runVersionAdapter(ctx, req.RunID, req.VersionID)
	if err != nil {
		return nil, err
	}
	_, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, err
	}
	if !spec.Capability.CanManualEdit {
		return nil, forbiddenError("scene does not support manual edit")
	}
	if run.Status.IsActive() {
		return nil, conflictError("run is active")
	}
	if run.Status == domain.RunStatusAdopted {
		return nil, conflictError("adopted run cannot be edited")
	}
	if err := state.EnsureCanCreateVersion(run); err != nil {
		return nil, err
	}

	now := s.now()
	edited := state.NewEditedVersion(run, version, state.EditVersionInput{
		Content:   req.Content,
		Artifacts: req.Artifacts,
		Actor:     req.Actor,
	}, spec, now)
	if err := s.store.CreateVersion(ctx, edited); err != nil {
		return nil, err
	}
	state.ApplyManualEditToRun(run, edited, req.Actor, now)
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, domain.EventManualEdit, run.ID, edited.ID, req.Actor, "", mustMarshal(map[string]string{
		"base_version_id": version.ID,
	}))
	return edited, nil
}

func (s *Service) ReviewVersion(ctx context.Context, req ReviewVersionRequest) (*domain.Version, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	run, version, _, err := s.runVersionAdapter(ctx, req.RunID, req.VersionID)
	if err != nil {
		return nil, err
	}
	_, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, err
	}
	if !spec.Capability.CanReviewOnly {
		return nil, forbiddenError("scene does not support review-only")
	}
	if run.Status.IsActive() {
		return nil, conflictError("run is active")
	}
	if len(version.EffectiveContent()) == 0 {
		return nil, invalidError("version has no effective content")
	}

	policy := domain.NormalizeReviewPolicy(req.OnFail)
	if policy == domain.ReviewPolicyAutoContinue && !s.allowAutoContinue {
		return nil, forbiddenError("auto continue is disabled")
	}
	if policy == domain.ReviewPolicyRunDefault {
		policy = domain.ReviewPolicyWaitManual
	}
	if err := s.dispatchReview(ctx, run, version, policy, req.Actor); err != nil {
		return nil, err
	}
	return s.store.GetVersion(ctx, version.ID)
}

func (s *Service) AdoptVersion(ctx context.Context, req AdoptVersionRequest) (*ports.AdoptResult, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	run, version, adapter, err := s.runVersionAdapter(ctx, req.RunID, req.VersionID)
	if err != nil {
		return nil, err
	}
	_, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, err
	}
	if !spec.Capability.CanAdopt {
		return nil, forbiddenError("scene does not support adopt")
	}
	if run.Status.IsActive() {
		return nil, conflictError("run is active")
	}
	if run.Status == domain.RunStatusAdopted {
		return nil, conflictError("run is already adopted")
	}

	content := version.EffectiveContent()
	if len(content) == 0 {
		return nil, invalidError("version has no effective content")
	}
	target, err := adapter.LoadTarget(ctx, run.Target)
	if err != nil {
		return nil, err
	}
	result, err := adapter.Adopt(ctx, ports.AdoptRequest{
		Run:       run,
		Version:   version,
		Target:    target,
		Content:   content,
		Artifacts: version.EffectiveArtifacts(),
		Actor:     strings.TrimSpace(req.Actor),
		Options:   domain.CloneRawMessage(req.Options),
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = &ports.AdoptResult{}
	}
	result.RunID = run.ID
	result.VersionID = version.ID
	if result.Target.Type == "" && result.Target.ID == "" {
		result.Target = run.Target
	}

	now := s.now()
	state.ApplyAdoptToVersion(version, req.Actor, now)
	state.ApplyAdoptToRun(run, version, req.Actor, now)
	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, domain.EventVersionAdopted, run.ID, version.ID, req.Actor, result.Message, result.Payload)
	return result, nil
}
