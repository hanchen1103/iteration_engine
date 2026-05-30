package ports

import (
	"context"
	"encoding/json"

	"github.com/hanchen1103/iteration_engine/domain"
)

type SceneAdapter interface {
	Spec() domain.SceneSpec
	LoadTarget(ctx context.Context, target domain.TargetRef) (*domain.TargetSnapshot, error)
	BuildGenerateJob(ctx context.Context, req GenerateRequest) (*JobRequest, error)
	ParseGenerateResult(ctx context.Context, raw []byte) (*domain.VersionContent, error)
	BuildReviewJob(ctx context.Context, req ReviewRequest) (*JobRequest, error)
	ParseReviewResult(ctx context.Context, raw []byte) (*domain.ReviewResult, error)
	Adopt(ctx context.Context, req AdoptRequest) (*AdoptResult, error)
}

type GenerateRequest struct {
	Run            *domain.Run
	Version        *domain.Version
	Target         *domain.TargetSnapshot
	Config         json.RawMessage
	Context        domain.GenerateContext
	ContextOptions domain.GenerateContextOptions
	Plan           domain.IterationPlan
}

type ReviewRequest struct {
	Run       *domain.Run
	Target    *domain.TargetSnapshot
	Version   *domain.Version
	Config    json.RawMessage
	Content   json.RawMessage
	Artifacts []domain.Artifact
}

type AdoptRequest struct {
	Run       *domain.Run
	Version   *domain.Version
	Target    *domain.TargetSnapshot
	Content   json.RawMessage
	Artifacts []domain.Artifact
	Actor     string
	Options   json.RawMessage
}

type AdoptResult struct {
	RunID     string           `json:"runID"`
	VersionID string           `json:"versionID"`
	Target    domain.TargetRef `json:"target"`
	Message   string           `json:"message,omitempty"`
	Payload   json.RawMessage  `json:"payload,omitempty"`
}
