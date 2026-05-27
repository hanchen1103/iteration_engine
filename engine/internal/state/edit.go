package state

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/hanchen1103/iteration_engine/domain"
)

type EditVersionInput struct {
	Content   json.RawMessage
	Artifacts []domain.Artifact
	Actor     string
}

func ApplyManualEditToRun(run *domain.Run, version *domain.Version, actor string, now time.Time) {
	run.Status = domain.RunStatusWaitingManual
	run.VersionCount = version.VersionNo
	run.FinalScore = nil
	run.FinalFeedback = ""
	run.ActiveJobID = ""
	run.ActiveRoleKey = ""
	run.ActiveVersionID = ""
	run.UpdatedBy = strings.TrimSpace(actor)
	run.UpdatedAt = now
}

func NewEditedVersion(run *domain.Run, base *domain.Version, input EditVersionInput, spec domain.SceneSpec, now time.Time) *domain.Version {
	generateRule := base.GenerateRuleSnapshot
	if generateRule.Role == "" && generateRule.RuleKey == "" {
		generateRule = spec.GenerateRule
	}
	reviewRule := base.ReviewRuleSnapshot
	if reviewRule.Role == "" && reviewRule.RuleKey == "" {
		reviewRule = spec.ReviewRule
	}
	return &domain.Version{
		RunID:                run.ID,
		VersionNo:            run.VersionCount + 1,
		BaseVersionID:        base.ID,
		Depth:                base.Depth + 1,
		Status:               domain.VersionStatusEdited,
		ReviewPolicy:         domain.ReviewPolicyWaitManual,
		IterationPlan:        manualEditPlan(base),
		TargetSnapshot:       domain.CloneRawMessage(base.TargetSnapshot),
		GenerateRuleSnapshot: generateRule,
		ReviewRuleSnapshot:   reviewRule,
		EditedContent:        domain.CloneRawMessage(input.Content),
		EditedArtifacts:      domain.CloneArtifacts(input.Artifacts),
		EditedBy:             strings.TrimSpace(input.Actor),
		EditedAt:             &now,
		CreatedBy:            strings.TrimSpace(input.Actor),
		UpdatedBy:            strings.TrimSpace(input.Actor),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func manualEditPlan(base *domain.Version) domain.IterationPlan {
	return domain.IterationPlan{
		BaseVersionID: base.ID,
		Source:        domain.PlanSourceManualEdit,
		Instruction:   "Use the manually edited content as the next candidate version.",
	}
}
