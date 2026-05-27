package model

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID        string          `json:"id"`
	RunID     string          `json:"runID"`
	VersionID string          `json:"versionID,omitempty"`
	Type      string          `json:"type"`
	Actor     string          `json:"actor,omitempty"`
	Message   string          `json:"message,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}

const (
	EventRunCreated        = "run_created"
	EventGenerateSubmitted = "generate_submitted"
	EventGenerateReceived  = "generate_received"
	EventReviewSubmitted   = "review_submitted"
	EventReviewReceived    = "review_received"
	EventManualContinue    = "manual_continue"
	EventManualEdit        = "manual_edit"
	EventVersionAdopted    = "version_adopted"
	EventRunFailed         = "run_failed"
)
