package model

import "encoding/json"

type TargetRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type TargetSnapshot struct {
	Ref      TargetRef       `json:"ref"`
	Snapshot json.RawMessage `json:"snapshot,omitempty"`
}
