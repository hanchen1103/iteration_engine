package model

import (
	"encoding/json"
	"math"
	"testing"
)

func TestScalarDirectiveHelpersWrapValue(t *testing.T) {
	intDirective := IntDirective("difficulty", 1, "1 means harder, -1 means easier.")
	assertDirectiveHeader(t, intDirective, "difficulty", "1 means harder, -1 means easier.")
	var intValue struct {
		Value int `json:"value"`
	}
	if err := json.Unmarshal(intDirective.Value, &intValue); err != nil {
		t.Fatalf("unmarshal int directive value: %v", err)
	}
	if intValue.Value != 1 {
		t.Fatalf("unexpected int directive value: %#v", intValue)
	}

	boolDirective := BoolDirective("preserve_topic", true, "Keep the same topic.")
	var boolValue struct {
		Value bool `json:"value"`
	}
	if err := json.Unmarshal(boolDirective.Value, &boolValue); err != nil {
		t.Fatalf("unmarshal bool directive value: %v", err)
	}
	if !boolValue.Value {
		t.Fatalf("unexpected bool directive value: %#v", boolValue)
	}

	stringDirective := StringDirective("tone", "formal", "Use a formal tone.")
	var stringValue struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(stringDirective.Value, &stringValue); err != nil {
		t.Fatalf("unmarshal string directive value: %v", err)
	}
	if stringValue.Value != "formal" {
		t.Fatalf("unexpected string directive value: %#v", stringValue)
	}
}

func TestValueDirectiveReturnsMarshalError(t *testing.T) {
	_, err := ValueDirective("score", math.NaN(), "Invalid numeric values should be rejected.")
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestObjectDirectiveUsesProvidedObject(t *testing.T) {
	directive, err := ObjectDirective("difficulty", map[string]any{
		"direction": "harder",
		"step":      1,
	}, "Make the next version harder.")
	if err != nil {
		t.Fatalf("ObjectDirective returned error: %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(directive.Value, &object); err != nil {
		t.Fatalf("unmarshal object directive value: %v", err)
	}
	if object["direction"] != "harder" || object["step"] != float64(1) {
		t.Fatalf("unexpected object directive value: %#v", object)
	}

	_, err = ObjectDirective("length", []string{"longer"}, "Make the next version longer.")
	if err == nil {
		t.Fatal("expected non-object value error")
	}
}

func assertDirectiveHeader(t *testing.T, directive IterationDirective, key string, description string) {
	t.Helper()
	if directive.Key != key || directive.Description != description {
		t.Fatalf("unexpected directive header: %#v", directive)
	}
}
