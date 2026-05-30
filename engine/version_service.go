package engine

import (
	"context"
	"encoding/json"
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
	if run.Status == domain.RunStatusGenerating || run.Status == domain.RunStatusReviewing {
		return nil, conflictError("run is active")
	}
	if run.Status == domain.RunStatusAdopted {
		return nil, conflictError("adopted run cannot be edited")
	}

	now := s.now()
	state.ApplyManualEditToVersion(version, state.EditVersionInput{
		Content:   req.Content,
		Artifacts: req.Artifacts,
		Actor:     req.Actor,
	}, spec, now)
	version.ReviewConfig = nil
	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}
	state.ApplyManualEditToRun(run, version, req.Actor, now)
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, domain.EventManualEdit, run.ID, version.ID, req.Actor, "", nil)
	return version, nil
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
	if run.Status == domain.RunStatusGenerating || run.Status == domain.RunStatusReviewing {
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
	plan := req.Plan
	if plan.Source == "" {
		plan.Source = domain.PlanSourceReviewOnly
	}
	if plan.BaseVersionID == "" {
		plan.BaseVersionID = version.ID
	}
	reviewVersion, err := s.createVersionFromContent(ctx, run, version, version.EffectiveContent(), version.EffectiveArtifacts(), plan, version.GenerateConfig, req.ReviewConfig, policy, req.Actor)
	if err != nil {
		return nil, err
	}
	if err := s.dispatchReview(ctx, run, reviewVersion, policy, req.ReviewConfig, req.Actor); err != nil {
		return nil, err
	}
	return s.store.GetVersion(ctx, reviewVersion.ID)
}

func (s *Service) SubmitCandidateForReview(ctx context.Context, req SubmitCandidateForReviewRequest) (*domain.Version, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	if len(req.Content) == 0 {
		return nil, invalidError("content is required")
	}
	run, err := s.getRun(ctx, strings.TrimSpace(req.RunID))
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
	if run.Status == domain.RunStatusGenerating || run.Status == domain.RunStatusReviewing {
		return nil, conflictError("run is active")
	}
	base, err := s.resolveOptionalBaseVersion(ctx, run, req.BaseVersionID, req.BaseVersionNo)
	if err != nil {
		return nil, err
	}
	policy := domain.NormalizeReviewPolicy(req.OnFail)
	if policy == domain.ReviewPolicyAutoContinue && !s.allowAutoContinue {
		return nil, forbiddenError("auto continue is disabled")
	}
	if policy == domain.ReviewPolicyRunDefault {
		policy = domain.ReviewPolicyWaitManual
	}
	plan := req.Plan
	if plan.Source == "" {
		plan.Source = domain.PlanSourceSubmittedCandidate
	}
	if base != nil && plan.BaseVersionID == "" {
		plan.BaseVersionID = base.ID
	}
	version, err := s.createVersionFromContent(ctx, run, base, req.Content, req.Artifacts, plan, req.GenerateConfig, req.ReviewConfig, policy, req.Actor)
	if err != nil {
		return nil, err
	}
	if err := s.dispatchReview(ctx, run, version, policy, req.ReviewConfig, req.Actor); err != nil {
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
	if target == nil {
		return nil, invalidError("adapter returned nil target")
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

func (s *Service) createVersionFromContent(ctx context.Context, run *domain.Run, base *domain.Version, content []byte, artifacts []domain.Artifact, plan domain.IterationPlan, generateConfig any, reviewConfig any, reviewPolicy domain.ReviewPolicy, actor string) (*domain.Version, error) {
	adapter, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, err
	}
	generateFallback := json.RawMessage(nil)
	if base != nil {
		generateFallback = base.GenerateConfig
	}
	effectiveGenerateConfig, err := resolveConfigWithFallback("generateConfig", generateConfig, generateFallback)
	if err != nil {
		return nil, err
	}
	effectiveReviewConfig, err := resolveReviewConfig(run, &domain.Version{GenerateConfig: effectiveGenerateConfig}, reviewConfig)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, invalidError("content is required")
	}
	if err := validatePlan(plan); err != nil {
		return nil, err
	}
	if err := state.EnsureCanCreateGeneratedVersion(run); err != nil {
		return nil, err
	}
	targetSnapshot := domain.CloneRawMessage(nil)
	if base != nil {
		targetSnapshot = domain.CloneRawMessage(base.TargetSnapshot)
	}
	if len(targetSnapshot) == 0 {
		target, err := adapter.LoadTarget(ctx, run.Target)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, invalidError("adapter returned nil target")
		}
		targetSnapshot = domain.CloneRawMessage(target.Snapshot)
	}

	now := s.now()
	depth := 1
	if base != nil {
		depth = base.Depth + 1
	}
	version := &domain.Version{
		RunID:                run.ID,
		VersionNo:            run.VersionCount + 1,
		Depth:                depth,
		Status:               domain.VersionStatusGenerated,
		ReviewPolicy:         domain.NormalizeReviewPolicy(reviewPolicy),
		IterationPlan:        plan,
		TargetSnapshot:       targetSnapshot,
		GenerateRuleSnapshot: spec.GenerateRule,
		ReviewRuleSnapshot:   spec.ReviewRule,
		GenerateConfig:       effectiveGenerateConfig,
		ReviewConfig:         effectiveReviewConfig,
		GeneratedContent:     domain.CloneRawMessage(content),
		GeneratedArtifacts:   domain.CloneArtifacts(artifacts),
		CreatedBy:            strings.TrimSpace(actor),
		UpdatedBy:            strings.TrimSpace(actor),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if base != nil {
		version.BaseVersionID = base.ID
		if version.IterationPlan.BaseVersionID == "" {
			version.IterationPlan.BaseVersionID = base.ID
		}
	}
	if err := s.store.CreateVersion(ctx, version); err != nil {
		return nil, err
	}
	run.VersionCount = version.VersionNo
	return version, nil
}
