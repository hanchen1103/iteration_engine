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

func autoReviewPlan(version *domain.Version, review *domain.ReviewResult, reviewOptions domain.ReviewContextOptions) domain.IterationPlan {
	explanation := ""
	instruction := "Revise the previous version using the review context. Fix blocking issues first."
	if review != nil && reviewOptions.ShouldIncludeFeedback() {
		explanation = strings.TrimSpace(review.Feedback)
	}
	if review != nil && explanation == "" && reviewOptions.ShouldIncludeSummary() {
		explanation = strings.TrimSpace(review.Summary)
	}
	if !reviewOptions.IncludesReviewGuidance() {
		instruction = "Revise the previous version."
	}
	return domain.IterationPlan{
		BaseVersionID: version.ID,
		Source:        domain.PlanSourceAutoReview,
		Explanation:   explanation,
		Instruction:   instruction,
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

func resolveGenerateContextOptions(run *domain.Run, override *domain.GenerateContextOptions) domain.GenerateContextOptions {
	if override != nil {
		return *domain.CloneGenerateContextOptions(override)
	}
	if run != nil && run.GenerateContext != nil {
		return *domain.CloneGenerateContextOptions(run.GenerateContext)
	}
	return domain.GenerateContextOptions{}
}

func buildGenerateContext(base *domain.Version, previousReview *domain.ReviewResult, options domain.GenerateContextOptions) domain.GenerateContext {
	return domain.GenerateContext{
		BaseVersion:    buildBaseVersionContext(base, options.BaseVersion),
		PreviousReview: buildReviewContext(previousReview, options.Review),
	}
}

func buildBaseVersionContext(version *domain.Version, options domain.BaseVersionContextOptions) *domain.BaseVersionContext {
	if version == nil || !options.IncludesAny() {
		return nil
	}
	out := &domain.BaseVersionContext{}
	if options.ShouldIncludeMetadata() {
		out.Metadata = &domain.BaseVersionMetadata{
			ID:            version.ID,
			RunID:         version.RunID,
			VersionNo:     version.VersionNo,
			BaseVersionID: version.BaseVersionID,
			Depth:         version.Depth,
			Status:        version.Status,
		}
	}
	if options.ShouldIncludeContent() {
		out.Content = version.EffectiveContent()
	}
	if options.ShouldIncludeArtifacts() {
		out.Artifacts = version.EffectiveArtifacts()
	}
	if out.Metadata == nil && len(out.Content) == 0 && len(out.Artifacts) == 0 {
		return nil
	}
	return out
}

func buildReviewContext(review *domain.ReviewResult, options domain.ReviewContextOptions) *domain.ReviewContext {
	if review == nil || !options.IncludesAny() {
		return nil
	}
	out := &domain.ReviewContext{}
	if options.ShouldIncludePass() {
		out.Pass = boolPtr(review.Pass)
	}
	if options.ShouldIncludeScore() {
		out.Score = cloneFloat64Ptr(review.Score)
	}
	if options.ShouldIncludeSummary() {
		out.Summary = review.Summary
	}
	if options.ShouldIncludeFeedback() {
		out.Feedback = review.Feedback
	}
	if options.ShouldIncludeIssues() {
		out.Issues = cloneReviewIssues(review.Issues)
	}
	if options.ShouldIncludeExtensions() {
		out.Extensions = filterReviewExtensions(review.Extensions, options.ExtensionKeys)
	}
	if options.ShouldIncludeRawJSON() {
		out.RawJSON = domain.CloneRawMessage(review.RawJSON)
	}
	if reviewContextEmpty(out) {
		return nil
	}
	return out
}

func filterReviewExtensions(in map[string]json.RawMessage, keys *[]string) map[string]json.RawMessage {
	if keys == nil {
		return domain.CloneRawMessageMap(in)
	}
	out := map[string]json.RawMessage{}
	for _, key := range *keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if raw, ok := in[key]; ok {
			out[key] = domain.CloneRawMessage(raw)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func reviewContextEmpty(review *domain.ReviewContext) bool {
	return review == nil ||
		(review.Pass == nil &&
			review.Score == nil &&
			review.Summary == "" &&
			review.Feedback == "" &&
			len(review.Issues) == 0 &&
			len(review.Extensions) == 0 &&
			len(review.RawJSON) == 0)
}

func boolPtr(value bool) *bool {
	return &value
}

func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
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
