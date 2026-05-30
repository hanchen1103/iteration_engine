package memory

import "github.com/hanchen1103/iteration_engine/domain"

func cloneRun(in *domain.Run) *domain.Run {
	if in == nil {
		return nil
	}
	out := *in
	out.Config = domain.CloneRawMessage(in.Config)
	out.GenerateContext = domain.CloneGenerateContextOptions(in.GenerateContext)
	out.DefaultDirectives = domain.CloneDirectives(in.DefaultDirectives)
	out.RuleSetSnapshot = domain.CloneRawMessage(in.RuleSetSnapshot)
	out.FinalScore = cloneFloat64Ptr(in.FinalScore)
	return &out
}

func cloneVersion(in *domain.Version) *domain.Version {
	if in == nil {
		return nil
	}
	out := *in
	out.IterationPlan.Directives = domain.CloneDirectives(in.IterationPlan.Directives)
	out.TargetSnapshot = domain.CloneRawMessage(in.TargetSnapshot)
	out.GenerateConfig = domain.CloneRawMessage(in.GenerateConfig)
	out.GenerateInputJSON = domain.CloneRawMessage(in.GenerateInputJSON)
	out.GeneratedContent = domain.CloneRawMessage(in.GeneratedContent)
	out.GeneratedArtifacts = domain.CloneArtifacts(in.GeneratedArtifacts)
	out.ReviewConfig = domain.CloneRawMessage(in.ReviewConfig)
	out.ReviewInputJSON = domain.CloneRawMessage(in.ReviewInputJSON)
	out.ReviewJSON = domain.CloneRawMessage(in.ReviewJSON)
	out.ReviewPass = cloneBoolPtr(in.ReviewPass)
	out.ReviewScore = cloneFloat64Ptr(in.ReviewScore)
	out.ReviewIssues = cloneReviewIssues(in.ReviewIssues)
	out.ReviewExtensions = domain.CloneRawMessageMap(in.ReviewExtensions)
	out.EditedContent = domain.CloneRawMessage(in.EditedContent)
	out.EditedArtifacts = domain.CloneArtifacts(in.EditedArtifacts)
	if in.EditedAt != nil {
		editedAt := *in.EditedAt
		out.EditedAt = &editedAt
	}
	return &out
}

func cloneEvent(in *domain.Event) *domain.Event {
	if in == nil {
		return nil
	}
	out := *in
	out.Payload = domain.CloneRawMessage(in.Payload)
	return &out
}

func cloneBoolPtr(in *bool) *bool {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneFloat64Ptr(in *float64) *float64 {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneReviewIssues(in []domain.ReviewIssue) []domain.ReviewIssue {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.ReviewIssue, len(in))
	copy(out, in)
	return out
}
