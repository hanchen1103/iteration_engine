package model

import (
	"encoding/json"
	"errors"
)

type Config map[string]any

func NormalizeConfig(value any) (json.RawMessage, error) {
	switch config := value.(type) {
	case nil:
		return nil, nil
	case json.RawMessage:
		return normalizeRawConfig(config)
	case *json.RawMessage:
		if config == nil {
			return nil, nil
		}
		return normalizeRawConfig(*config)
	case []byte:
		return normalizeRawConfig(config)
	case Config:
		return marshalConfig(config)
	case map[string]any:
		return marshalConfig(config)
	default:
		return marshalConfig(config)
	}
}

func normalizeRawConfig(raw []byte) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if !json.Valid(raw) {
		return nil, errors.New("config must be valid JSON")
	}
	return CloneRawMessage(json.RawMessage(raw)), nil
}

func marshalConfig(value any) (json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	return json.RawMessage(raw), nil
}
