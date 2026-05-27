package model

import "encoding/json"

type PlanSource string

const (
	PlanSourceInitial            PlanSource = "initial"
	PlanSourceManual             PlanSource = "manual"
	PlanSourceAutoReview         PlanSource = "auto_review"
	PlanSourceManualEdit         PlanSource = "manual_edit"
	PlanSourceSubmittedCandidate PlanSource = "submitted_candidate"
	PlanSourceReviewOnly         PlanSource = "review_only"
)

type IterationPlan struct {
	BaseVersionID string               `json:"baseVersionID,omitempty"`
	Source        PlanSource           `json:"source"`
	Explanation   string               `json:"explanation,omitempty"`
	Instruction   string               `json:"instruction,omitempty"`
	Directives    []IterationDirective `json:"directives,omitempty"`
}

type IterationDirective struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value"`
	Description string          `json:"description"`
}
