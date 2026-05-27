package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

func fillJobRequest(req *ports.JobRequest, run *domain.Run, version *domain.Version, role string, fallbackRole string) {
	if req == nil {
		return
	}
	req.RunID = run.ID
	req.VersionID = version.ID
	req.SceneKey = run.SceneKey
	if strings.TrimSpace(req.RoleKey) == "" {
		req.RoleKey = strings.TrimSpace(role)
	}
	if strings.TrimSpace(req.RoleKey) == "" {
		req.RoleKey = fallbackRole
	}
	if req.Metadata == nil {
		req.Metadata = map[string]string{}
	}
	req.Metadata["run_id"] = run.ID
	req.Metadata["version_id"] = version.ID
	req.Metadata["version_no"] = fmt.Sprintf("%d", version.VersionNo)
	req.Metadata["depth"] = fmt.Sprintf("%d", version.Depth)
}

func autoReviewPlan(version *domain.Version, review *domain.ReviewResult) domain.IterationPlan {
	explanation := ""
	if review != nil {
		explanation = strings.TrimSpace(review.Feedback)
		if explanation == "" {
			explanation = strings.TrimSpace(review.Summary)
		}
	}
	return domain.IterationPlan{
		BaseVersionID: version.ID,
		Source:        domain.PlanSourceAutoReview,
		Explanation:   explanation,
		Instruction:   "Revise the previous version using the review feedback. Fix blocking issues first.",
	}
}

func reviewResultFromVersion(version *domain.Version) *domain.ReviewResult {
	if version == nil || version.ReviewPass == nil {
		return nil
	}
	return &domain.ReviewResult{
		Pass:       *version.ReviewPass,
		Score:      cloneFloat64Ptr(version.ReviewScore),
		Summary:    version.ReviewSummary,
		Feedback:   version.ReviewFeedback,
		Issues:     cloneReviewIssues(version.ReviewIssues),
		Extensions: domain.CloneRawMessageMap(version.ReviewExtensions),
		RawJSON:    domain.CloneRawMessage(version.ReviewJSON),
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneReviewIssues(in []domain.ReviewIssue) []domain.ReviewIssue {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.ReviewIssue, len(in))
	copy(out, in)
	return out
}

func mustMarshal(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return raw
}
