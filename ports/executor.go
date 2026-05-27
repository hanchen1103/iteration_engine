package ports

import (
	"context"
	"encoding/json"
)

type JobExecutor interface {
	Submit(ctx context.Context, req *JobRequest) (*JobHandle, error)
}

type JobRequest struct {
	RunID     string            `json:"runID"`
	VersionID string            `json:"versionID"`
	SceneKey  string            `json:"sceneKey"`
	RoleKey   string            `json:"roleKey"`
	TaskName  string            `json:"taskName,omitempty"`
	Input     json.RawMessage   `json:"input,omitempty"`
	Schema    json.RawMessage   `json:"schema,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type JobHandle struct {
	JobID string `json:"jobID"`
}
