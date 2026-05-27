package model

import "encoding/json"

func CloneRawMessage(in json.RawMessage) json.RawMessage {
	if len(in) == 0 {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func CloneRawMessageMap(in map[string]json.RawMessage) map[string]json.RawMessage {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]json.RawMessage, len(in))
	for key, value := range in {
		out[key] = CloneRawMessage(value)
	}
	return out
}

func CloneArtifacts(in []Artifact) []Artifact {
	if len(in) == 0 {
		return nil
	}
	out := make([]Artifact, len(in))
	for i, item := range in {
		out[i] = item
		out[i].Data = CloneRawMessage(item.Data)
	}
	return out
}

func CloneDirectives(in []IterationDirective) []IterationDirective {
	if len(in) == 0 {
		return nil
	}
	out := make([]IterationDirective, len(in))
	for i, item := range in {
		out[i] = item
		out[i].Value = CloneRawMessage(item.Value)
	}
	return out
}
