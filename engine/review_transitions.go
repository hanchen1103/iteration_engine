package engine

import (
	"context"

	"github.com/hanchen1103/iteration_engine/domain"
	decisionpkg "github.com/hanchen1103/iteration_engine/domain/decision"
	"github.com/hanchen1103/iteration_engine/engine/internal/state"
)

func (s *Service) applyReviewDecision(ctx context.Context, run *domain.Run, version *domain.Version, review *domain.ReviewResult, spec domain.SceneSpec) error {
	decision := decisionpkg.Decide(run, version, review)
	now := s.now()

	switch decision.Type {
	case domain.DecisionPass:
		state.ApplyReviewedRunState(run, domain.RunStatusSucceeded, review, now)
		return s.store.UpdateRun(ctx, run)
	case domain.DecisionMaxIterations:
		state.ApplyReviewedRunState(run, domain.RunStatusMaxIterations, review, now)
		return s.store.UpdateRun(ctx, run)
	case domain.DecisionWaitManual:
		state.ApplyReviewedRunState(run, domain.RunStatusWaitingManual, review, now)
		return s.store.UpdateRun(ctx, run)
	case domain.DecisionAutoContinue:
		if !s.allowAutoContinue || !spec.Capability.CanAutoContinue {
			state.ApplyReviewedRunState(run, domain.RunStatusWaitingManual, review, now)
			return s.store.UpdateRun(ctx, run)
		}
		contextOptions := resolveGenerateContextOptions(run, nil)
		plan := autoReviewPlan(version, review, contextOptions.Review)
		_, err := s.startGenerate(ctx, run, version, review, plan, contextOptions, domain.ReviewPolicyRunDefault, "")
		return err
	default:
		return s.failRunVersion(ctx, run, version, decision.Message)
	}
}
