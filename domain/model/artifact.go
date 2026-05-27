package model

import "encoding/json"

type Artifact struct {
	Key      string          `json:"key"`
	MimeType string          `json:"mimeType,omitempty"`
	URI      string          `json:"uri,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}
