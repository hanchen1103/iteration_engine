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
	if version.VersionNo > run.VersionCount {
		run.VersionCount = version.VersionNo
	}
	run.FinalScore = nil
	run.FinalFeedback = ""
	run.ActiveJobID = ""
	run.ActiveRoleKey = ""
	run.ActiveVersionID = ""
	run.UpdatedBy = strings.TrimSpace(actor)
	run.UpdatedAt = now
}

func ApplyManualEditToVersion(version *domain.Version, input EditVersionInput, spec domain.SceneSpec, now time.Time) {
	if version.GenerateRuleSnapshot.Role == "" && version.GenerateRuleSnapshot.RuleKey == "" {
		version.GenerateRuleSnapshot = spec.GenerateRule
	}
	if version.ReviewRuleSnapshot.Role == "" && version.ReviewRuleSnapshot.RuleKey == "" {
		version.ReviewRuleSnapshot = spec.ReviewRule
	}
	version.Status = domain.VersionStatusGenerated
	version.ReviewPolicy = domain.ReviewPolicyWaitManual
	version.GeneratedContent = domain.CloneRawMessage(input.Content)
	version.GeneratedArtifacts = domain.CloneArtifacts(input.Artifacts)
	version.EditedContent = nil
	version.EditedArtifacts = nil
	version.EditedBy = ""
	version.EditedAt = nil
	version.ReviewInputJSON = nil
	version.ReviewJSON = nil
	version.ReviewPass = nil
	version.ReviewScore = nil
	version.ReviewSummary = ""
	version.ReviewFeedback = ""
	version.ReviewIssues = nil
	version.ReviewExtensions = nil
	version.ReviewJobID = ""
	version.ErrorMessage = ""
	version.UpdatedBy = strings.TrimSpace(input.Actor)
	version.UpdatedAt = now
}
