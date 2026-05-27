package model

import "encoding/json"

type ReviewResult struct {
	Pass       bool                       `json:"pass"`
	Score      *float64                   `json:"score,omitempty"`
	Summary    string                     `json:"summary,omitempty"`
	Feedback   string                     `json:"feedback,omitempty"`
	Issues     []ReviewIssue              `json:"issues,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	RawJSON    json.RawMessage            `json:"rawJson,omitempty"`
}

type ReviewIssue struct {
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Path       string `json:"path,omitempty"`
}
