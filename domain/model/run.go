package model

import (
	"encoding/json"
	"time"
)

type Run struct {
	ID                string                  `json:"id"`
	SceneKey          string                  `json:"sceneKey"`
	Target            TargetRef               `json:"target"`
	Status            RunStatus               `json:"status"`
	IterationMode     IterationMode           `json:"iterationMode"`
	MaxIterations     int                     `json:"maxIterations"`
	VersionCount      int                     `json:"versionCount"`
	AdoptedVersionID  string                  `json:"adoptedVersionID,omitempty"`
	Config            json.RawMessage         `json:"config,omitempty"`
	GenerateContext   *GenerateContextOptions `json:"generateContext,omitempty"`
	DefaultDirectives []IterationDirective    `json:"defaultDirectives,omitempty"`
	RuleSetSnapshot   json.RawMessage         `json:"ruleSetSnapshot,omitempty"`
	FinalScore        *float64                `json:"finalScore,omitempty"`
	FinalFeedback     string                  `json:"finalFeedback,omitempty"`
	ErrorMessage      string                  `json:"errorMessage,omitempty"`
	ActiveVersionID   string                  `json:"activeVersionID,omitempty"`
	ActiveJobID       string                  `json:"activeJobID,omitempty"`
	ActiveRoleKey     string                  `json:"activeRoleKey,omitempty"`
	CreatedBy         string                  `json:"createdBy,omitempty"`
	UpdatedBy         string                  `json:"updatedBy,omitempty"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}
