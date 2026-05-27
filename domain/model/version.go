package model

import (
	"encoding/json"
	"time"
)

type Version struct {
	ID                   string                     `json:"id"`
	RunID                string                     `json:"runID"`
	VersionNo            int                        `json:"versionNo"`
	BaseVersionID        string                     `json:"baseVersionID,omitempty"`
	Depth                int                        `json:"depth"`
	Status               VersionStatus              `json:"status"`
	ReviewPolicy         ReviewPolicy               `json:"reviewPolicy,omitempty"`
	IterationPlan        IterationPlan              `json:"iterationPlan"`
	TargetSnapshot       json.RawMessage            `json:"targetSnapshot,omitempty"`
	GenerateRuleSnapshot RuleSpec                   `json:"generateRuleSnapshot"`
	ReviewRuleSnapshot   RuleSpec                   `json:"reviewRuleSnapshot"`
	GenerateAttemptNo    int                        `json:"generateAttemptNo"`
	ReviewAttemptNo      int                        `json:"reviewAttemptNo"`
	GenerateJobID        string                     `json:"generateJobID,omitempty"`
	ReviewJobID          string                     `json:"reviewJobID,omitempty"`
	GenerateInputJSON    json.RawMessage            `json:"generateInputJson,omitempty"`
	GeneratedContent     json.RawMessage            `json:"generatedContent,omitempty"`
	GeneratedArtifacts   []Artifact                 `json:"generatedArtifacts,omitempty"`
	ReviewInputJSON      json.RawMessage            `json:"reviewInputJson,omitempty"`
	ReviewJSON           json.RawMessage            `json:"reviewJson,omitempty"`
	ReviewPass           *bool                      `json:"reviewPass,omitempty"`
	ReviewScore          *float64                   `json:"reviewScore,omitempty"`
	ReviewSummary        string                     `json:"reviewSummary,omitempty"`
	ReviewFeedback       string                     `json:"reviewFeedback,omitempty"`
	ReviewIssues         []ReviewIssue              `json:"reviewIssues,omitempty"`
	ReviewExtensions     map[string]json.RawMessage `json:"reviewExtensions,omitempty"`
	EditedContent        json.RawMessage            `json:"editedContent,omitempty"`
	EditedArtifacts      []Artifact                 `json:"editedArtifacts,omitempty"`
	EditedBy             string                     `json:"editedBy,omitempty"`
	EditedAt             *time.Time                 `json:"editedAt,omitempty"`
	ErrorMessage         string                     `json:"errorMessage,omitempty"`
	CreatedBy            string                     `json:"createdBy,omitempty"`
	UpdatedBy            string                     `json:"updatedBy,omitempty"`
	CreatedAt            time.Time                  `json:"createdAt"`
	UpdatedAt            time.Time                  `json:"updatedAt"`
}

func (v *Version) EffectiveContent() json.RawMessage {
	if v == nil {
		return nil
	}
	if len(v.EditedContent) > 0 {
		return CloneRawMessage(v.EditedContent)
	}
	return CloneRawMessage(v.GeneratedContent)
}

func (v *Version) EffectiveArtifacts() []Artifact {
	if v == nil {
		return nil
	}
	if len(v.EditedContent) > 0 {
		return CloneArtifacts(v.EditedArtifacts)
	}
	return CloneArtifacts(v.GeneratedArtifacts)
}

type VersionContent struct {
	Content   json.RawMessage `json:"content"`
	Artifacts []Artifact      `json:"artifacts,omitempty"`
}

type VersionNode struct {
	Version  *Version       `json:"version"`
	Children []*VersionNode `json:"children,omitempty"`
}
