package testkit

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

type Adapter struct {
	CanAuto         bool
	DefaultMaxDepth int
	Adopted         []json.RawMessage
}

func (a *Adapter) Spec() domain.SceneSpec {
	return domain.SceneSpec{
		SceneKey:   "fake",
		TargetType: "document",
		GenerateRule: domain.RuleSpec{
			Role:        "author",
			RuleKey:     "fake_author",
			RuleVersion: "v1",
		},
		ReviewRule: domain.RuleSpec{
			Role:        "review",
			RuleKey:     "fake_review",
			RuleVersion: "v1",
		},
		Capability: domain.SceneCapability{
			CanAutoContinue: a.CanAuto,
			CanManualEdit:   true,
			CanReviewOnly:   true,
			CanAdopt:        true,
			DefaultMaxDepth: a.DefaultMaxDepth,
		},
	}
}

func (a *Adapter) LoadTarget(ctx context.Context, target domain.TargetRef) (*domain.TargetSnapshot, error) {
	_ = ctx
	return &domain.TargetSnapshot{Ref: target, Snapshot: RawJSON(`{"loaded":true}`)}, nil
}

func (a *Adapter) BuildGenerateJob(ctx context.Context, req ports.GenerateRequest) (*ports.JobRequest, error) {
	_ = ctx
	input, _ := json.Marshal(map[string]any{
		"plan":             req.Plan,
		"has_base_version": req.BaseVersion != nil,
		"previous_review":  req.PreviousReview,
	})
	return &ports.JobRequest{TaskName: "fake_generate", Input: input}, nil
}

func (a *Adapter) ParseGenerateResult(ctx context.Context, raw []byte) (*domain.VersionContent, error) {
	_ = ctx
	if !json.Valid(raw) {
		return nil, errors.New("generate result must be json")
	}
	return &domain.VersionContent{Content: domain.CloneRawMessage(raw)}, nil
}

func (a *Adapter) BuildReviewJob(ctx context.Context, req ports.ReviewRequest) (*ports.JobRequest, error) {
	_ = ctx
	input, _ := json.Marshal(map[string]any{
		"content": json.RawMessage(req.Content),
	})
	return &ports.JobRequest{TaskName: "fake_review", Input: input}, nil
}

func (a *Adapter) ParseReviewResult(ctx context.Context, raw []byte) (*domain.ReviewResult, error) {
	_ = ctx
	var parsed struct {
		Pass     bool                 `json:"pass"`
		Score    *float64             `json:"score"`
		Summary  string               `json:"summary"`
		Feedback string               `json:"feedback"`
		Issues   []domain.ReviewIssue `json:"issues"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	var extensions map[string]json.RawMessage
	if err := json.Unmarshal(raw, &extensions); err != nil {
		return nil, err
	}
	for _, key := range []string{"pass", "score", "summary", "feedback", "issues"} {
		delete(extensions, key)
	}
	return &domain.ReviewResult{
		Pass:       parsed.Pass,
		Score:      parsed.Score,
		Summary:    parsed.Summary,
		Feedback:   parsed.Feedback,
		Issues:     parsed.Issues,
		Extensions: domain.CloneRawMessageMap(extensions),
		RawJSON:    domain.CloneRawMessage(raw),
	}, nil
}

func (a *Adapter) Adopt(ctx context.Context, req ports.AdoptRequest) (*ports.AdoptResult, error) {
	_ = ctx
	a.Adopted = append(a.Adopted, domain.CloneRawMessage(req.Content))
	return &ports.AdoptResult{Message: "adopted"}, nil
}
