package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/engine/internal/state"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) ReceiveGenerateResult(ctx context.Context, req GenerateResultRequest) error {
	if err := s.ready(); err != nil {
		return err
	}
	jobID := strings.TrimSpace(req.JobID)
	if jobID == "" {
		return invalidError("jobID is required")
	}
	version, role, err := s.store.FindVersionByJobID(ctx, jobID)
	if err != nil {
		return err
	}
	if role != ports.JobRoleGenerate {
		return invalidError("jobID is not a generate job")
	}
	run, err := s.store.GetRun(ctx, version.RunID)
	if err != nil {
		return err
	}
	if run.Status.IsClosed() || run.ActiveJobID != jobID {
		return nil
	}
	if strings.TrimSpace(req.ErrorMessage) != "" {
		return s.failRunVersion(ctx, run, version, strings.TrimSpace(req.ErrorMessage))
	}

	adapter, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return err
	}
	content, err := adapter.ParseGenerateResult(ctx, req.Raw)
	if err != nil {
		return s.failRunVersion(ctx, run, version, err.Error())
	}
	if content == nil {
		return s.failRunVersion(ctx, run, version, "adapter returned nil generate result")
	}

	version.GeneratedContent = domain.CloneRawMessage(content.Content)
	version.GeneratedArtifacts = domain.CloneArtifacts(content.Artifacts)
	version.Status = domain.VersionStatusGenerated
	version.ErrorMessage = ""
	now := s.now()
	version.UpdatedAt = now
	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return err
	}
	s.recordEvent(ctx, domain.EventGenerateReceived, run.ID, version.ID, "", "", nil)
	if domain.NormalizeIterationMode(run.IterationMode) == domain.IterationModeAuto && s.allowAutoContinue && spec.Capability.CanAutoContinue {
		return s.dispatchReview(ctx, run, version, domain.ReviewPolicyRunDefault, nil, "")
	}
	state.ApplyGeneratedRunState(run, version, now)
	return s.store.UpdateRun(ctx, run)
}

func (s *Service) ReceiveReviewResult(ctx context.Context, req ReviewResultRequest) error {
	if err := s.ready(); err != nil {
		return err
	}
	jobID := strings.TrimSpace(req.JobID)
	if jobID == "" {
		return invalidError("jobID is required")
	}
	version, role, err := s.store.FindVersionByJobID(ctx, jobID)
	if err != nil {
		return err
	}
	if role != ports.JobRoleReview {
		return invalidError("jobID is not a review job")
	}
	run, err := s.store.GetRun(ctx, version.RunID)
	if err != nil {
		return err
	}
	if run.Status.IsClosed() || run.ActiveJobID != jobID {
		return nil
	}
	if strings.TrimSpace(req.ErrorMessage) != "" {
		return s.failRunVersion(ctx, run, version, strings.TrimSpace(req.ErrorMessage))
	}

	adapter, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return err
	}
	review, err := adapter.ParseReviewResult(ctx, req.Raw)
	if err != nil {
		return s.failRunVersion(ctx, run, version, err.Error())
	}
	if review == nil {
		return s.failRunVersion(ctx, run, version, "adapter returned nil review result")
	}
	if len(review.RawJSON) == 0 {
		review.RawJSON = domain.CloneRawMessage(req.Raw)
	}

	now := s.now()
	state.ApplyReviewToVersion(version, review, now)
	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return err
	}
	s.recordEvent(ctx, domain.EventReviewReceived, run.ID, version.ID, "", "", version.ReviewJSON)

	return s.applyReviewDecision(ctx, run, version, review, spec)
}
