package model

import "encoding/json"

type directiveValue struct {
	Value any `json:"value"`
}

// BoolDirective creates a directive whose JSON value is {"value": <bool>}.
func BoolDirective(key string, value bool, description string) IterationDirective {
	return mustValueDirective(key, value, description)
}

// IntDirective creates a directive whose JSON value is {"value": <int>}.
func IntDirective(key string, value int, description string) IterationDirective {
	return mustValueDirective(key, value, description)
}

// StringDirective creates a directive whose JSON value is {"value": <string>}.
func StringDirective(key string, value string, description string) IterationDirective {
	return mustValueDirective(key, value, description)
}

// ValueDirective creates a directive by wrapping any JSON-marshalable value under the "value" field.
func ValueDirective(key string, value any, description string) (IterationDirective, error) {
	raw, err := json.Marshal(directiveValue{Value: value})
	if err != nil {
		return IterationDirective{}, err
	}
	return IterationDirective{
		Key:         key,
		Value:       raw,
		Description: description,
	}, nil
}

// ObjectDirective creates a directive from a caller-defined JSON object shape.
func ObjectDirective(key string, value any, description string) (IterationDirective, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return IterationDirective{}, err
	}
	if !isJSONObject(raw) {
		return IterationDirective{}, NewError(ErrorCodeInvalid, "directive value must be a JSON object")
	}
	return IterationDirective{
		Key:         key,
		Value:       raw,
		Description: description,
	}, nil
}

func mustValueDirective(key string, value any, description string) IterationDirective {
	directive, err := ValueDirective(key, value, description)
	if err != nil {
		panic("iteration_engine: failed to marshal directive value: " + err.Error())
	}
	return directive
}

func isJSONObject(raw json.RawMessage) bool {
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return false
	}
	return object != nil
}
