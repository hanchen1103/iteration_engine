package engine

import (
	"encoding/json"

	"github.com/hanchen1103/iteration_engine/domain"
)

type CreateRunRequest struct {
	SceneKey          string                      `json:"sceneKey"`
	Target            domain.TargetRef            `json:"target"`
	IterationMode     domain.IterationMode        `json:"iterationMode,omitempty"`
	MaxIterations     int                         `json:"maxIterations,omitempty"`
	Config            json.RawMessage             `json:"config,omitempty"`
	DefaultDirectives []domain.IterationDirective `json:"defaultDirectives,omitempty"`
	Actor             string                      `json:"actor,omitempty"`
}

type ContinueRunRequest struct {
	RunID         string               `json:"runID"`
	BaseVersionID string               `json:"baseVersionID,omitempty"`
	BaseVersionNo int                  `json:"baseVersionNo,omitempty"`
	MaxIterations int                  `json:"maxIterations,omitempty"`
	Plan          domain.IterationPlan `json:"plan"`
	Actor         string               `json:"actor,omitempty"`
}

type GenerateResultRequest struct {
	JobID        string `json:"jobID"`
	Raw          []byte `json:"raw,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type ReviewResultRequest struct {
	JobID        string `json:"jobID"`
	Raw          []byte `json:"raw,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type EditVersionRequest struct {
	RunID     string            `json:"runID"`
	VersionID string            `json:"versionID"`
	Content   json.RawMessage   `json:"content"`
	Artifacts []domain.Artifact `json:"artifacts,omitempty"`
	Actor     string            `json:"actor,omitempty"`
}

type ReviewVersionRequest struct {
	RunID     string               `json:"runID"`
	VersionID string               `json:"versionID"`
	OnFail    domain.ReviewPolicy  `json:"onFail,omitempty"`
	Plan      domain.IterationPlan `json:"plan"`
	Actor     string               `json:"actor,omitempty"`
}

type SubmitCandidateForReviewRequest struct {
	RunID         string               `json:"runID"`
	BaseVersionID string               `json:"baseVersionID,omitempty"`
	BaseVersionNo int                  `json:"baseVersionNo,omitempty"`
	Content       json.RawMessage      `json:"content"`
	Artifacts     []domain.Artifact    `json:"artifacts,omitempty"`
	Plan          domain.IterationPlan `json:"plan"`
	OnFail        domain.ReviewPolicy  `json:"onFail,omitempty"`
	Actor         string               `json:"actor,omitempty"`
}

type AdoptVersionRequest struct {
	RunID     string          `json:"runID"`
	VersionID string          `json:"versionID"`
	Actor     string          `json:"actor,omitempty"`
	Options   json.RawMessage `json:"options,omitempty"`
}

type RunDetail struct {
	Run         *domain.Run           `json:"run"`
	Versions    []*domain.Version     `json:"versions"`
	VersionTree []*domain.VersionNode `json:"versionTree,omitempty"`
	Events      []*domain.Event       `json:"events,omitempty"`
}
