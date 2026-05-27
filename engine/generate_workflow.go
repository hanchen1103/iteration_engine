package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/engine/internal/state"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) startGenerate(ctx context.Context, run *domain.Run, base *domain.Version, previousReview *domain.ReviewResult, plan domain.IterationPlan, reviewPolicy domain.ReviewPolicy, actor string) (*domain.Version, error) {
	adapter, spec, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, err
	}
	if err := validatePlan(plan); err != nil {
		return nil, err
	}
	if err := state.EnsureCanCreateGeneratedVersion(run, base); err != nil {
		return nil, err
	}
	target, err := adapter.LoadTarget(ctx, run.Target)
	if err != nil {
		return nil, err
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
		Status:               domain.VersionStatusGenerating,
		ReviewPolicy:         domain.NormalizeReviewPolicy(reviewPolicy),
		IterationPlan:        plan,
		TargetSnapshot:       domain.CloneRawMessage(target.Snapshot),
		GenerateRuleSnapshot: spec.GenerateRule,
		ReviewRuleSnapshot:   spec.ReviewRule,
		GenerateAttemptNo:    1,
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

	jobReq, err := adapter.BuildGenerateJob(ctx, ports.GenerateRequest{
		Run:            run,
		Target:         target,
		BaseVersion:    base,
		PreviousReview: previousReview,
		Plan:           version.IterationPlan,
	})
	if err != nil {
		_ = s.failRunVersion(ctx, run, version, err.Error())
		return nil, err
	}
	fillJobRequest(jobReq, run, version, spec.GenerateRule.Role, "generate")
	version.GenerateInputJSON = domain.CloneRawMessage(jobReq.Input)

	handle, err := s.executor.Submit(ctx, jobReq)
	if err != nil {
		_ = s.failRunVersion(ctx, run, version, err.Error())
		return nil, err
	}
	if handle == nil || strings.TrimSpace(handle.JobID) == "" {
		err := failedError("executor returned empty generate job id")
		_ = s.failRunVersion(ctx, run, version, err.Error())
		return nil, err
	}

	version.GenerateJobID = strings.TrimSpace(handle.JobID)
	version.Status = domain.VersionStatusGenerating
	version.UpdatedAt = now
	run.Status = domain.RunStatusGenerating
	run.VersionCount = version.VersionNo
	run.ActiveVersionID = version.ID
	run.ActiveJobID = version.GenerateJobID
	run.ActiveRoleKey = jobReq.RoleKey
	run.FinalScore = nil
	run.FinalFeedback = ""
	run.ErrorMessage = ""
	run.UpdatedBy = strings.TrimSpace(actor)
	run.UpdatedAt = now

	if err := s.store.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return nil, err
	}
	s.recordEvent(ctx, domain.EventGenerateSubmitted, run.ID, version.ID, actor, "", mustMarshal(jobReq))
	return version, nil
}
