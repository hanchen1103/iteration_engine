package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) dispatchReview(ctx context.Context, run *domain.Run, version *domain.Version, policy domain.ReviewPolicy, actor string) error {
	adapter, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return err
	}
	content := version.EffectiveContent()
	if len(content) == 0 {
		return invalidError("version has no effective content")
	}

	target := &domain.TargetSnapshot{
		Ref:      run.Target,
		Snapshot: domain.CloneRawMessage(version.TargetSnapshot),
	}
	if len(target.Snapshot) == 0 {
		target, err = adapter.LoadTarget(ctx, run.Target)
		if err != nil {
			return err
		}
	}

	jobReq, err := adapter.BuildReviewJob(ctx, ports.ReviewRequest{
		Run:       run,
		Target:    target,
		Version:   version,
		Content:   content,
		Artifacts: version.EffectiveArtifacts(),
	})
	if err != nil {
		return s.failRunVersion(ctx, run, version, err.Error())
	}
	fillJobRequest(jobReq, run, version, spec.ReviewRule.Role, "review")

	now := s.now()
	version.ReviewPolicy = domain.NormalizeReviewPolicy(policy)
	version.ReviewAttemptNo++
	version.ReviewInputJSON = domain.CloneRawMessage(jobReq.Input)
	version.ReviewJSON = nil
	version.ReviewPass = nil
	version.ReviewScore = nil
	version.ReviewSummary = ""
	version.ReviewFeedback = ""
	version.ReviewIssues = nil
	version.Status = domain.VersionStatusReviewing
	version.UpdatedBy = strings.TrimSpace(actor)
	version.UpdatedAt = now

	handle, err := s.executor.Submit(ctx, jobReq)
	if err != nil {
		return s.failRunVersion(ctx, run, version, err.Error())
	}
	if handle == nil || strings.TrimSpace(handle.JobID) == "" {
		return s.failRunVersion(ctx, run, version, "executor returned empty review job id")
	}

	version.ReviewJobID = strings.TrimSpace(handle.JobID)
	run.Status = domain.RunStatusReviewing
	run.ActiveVersionID = version.ID
	run.ActiveJobID = version.ReviewJobID
	run.ActiveRoleKey = jobReq.RoleKey
	run.FinalScore = nil
	run.FinalFeedback = ""
	run.ErrorMessage = ""
	run.UpdatedBy = strings.TrimSpace(actor)
	run.UpdatedAt = now

	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return err
	}
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return err
	}
	s.recordEvent(ctx, domain.EventReviewSubmitted, run.ID, version.ID, actor, "", mustMarshal(jobReq))
	return nil
}
