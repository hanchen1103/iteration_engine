package model

import "testing"

func TestGenerateContextOptionsBuilder(t *testing.T) {
	options := NewGenerateContextOptions(
		WithoutBaseVersion(),
		WithBaseContent(),
		WithoutReview(),
		WithReviewFeedback(),
		WithReviewExtensions(" difficulty_profile ", "", "manual_controls"),
	)

	if options.BaseVersion.ShouldIncludeMetadata() {
		t.Fatal("base metadata should be disabled")
	}
	if !options.BaseVersion.ShouldIncludeContent() {
		t.Fatal("base content should be enabled")
	}
	if options.BaseVersion.ShouldIncludeArtifacts() {
		t.Fatal("base artifacts should be disabled")
	}
	if options.Review.ShouldIncludePass() || options.Review.ShouldIncludeScore() || options.Review.ShouldIncludeSummary() {
		t.Fatalf("non-feedback review fields should be disabled: %#v", options.Review)
	}
	if !options.Review.ShouldIncludeFeedback() || !options.Review.ShouldIncludeExtensions() {
		t.Fatalf("feedback and extensions should be enabled: %#v", options.Review)
	}
	if options.Review.ExtensionKeys == nil {
		t.Fatal("extension keys should be set")
	}
	keys := *options.Review.ExtensionKeys
	if len(keys) != 2 || keys[0] != "difficulty_profile" || keys[1] != "manual_controls" {
		t.Fatalf("unexpected extension keys: %#v", keys)
	}
}

func TestGenerateContextPresets(t *testing.T) {
	none := GenerateContextNone()
	if none.BaseVersion.IncludesAny() || none.Review.IncludesAny() {
		t.Fatalf("none preset should disable all context: %#v", none)
	}

	baseOnly := GenerateContextBaseContentOnly()
	if !baseOnly.BaseVersion.ShouldIncludeContent() || baseOnly.BaseVersion.ShouldIncludeMetadata() || baseOnly.BaseVersion.ShouldIncludeArtifacts() {
		t.Fatalf("base content preset is wrong: %#v", baseOnly.BaseVersion)
	}
	if baseOnly.Review.IncludesAny() {
		t.Fatalf("base content preset should disable review: %#v", baseOnly.Review)
	}

	feedbackOnly := GenerateContextReviewFeedbackOnly()
	if feedbackOnly.BaseVersion.IncludesAny() || !feedbackOnly.Review.ShouldIncludeFeedback() {
		t.Fatalf("review feedback preset is wrong: %#v", feedbackOnly)
	}
	if feedbackOnly.Review.ShouldIncludePass() || feedbackOnly.Review.ShouldIncludeExtensions() {
		t.Fatalf("review feedback preset should not include other review fields: %#v", feedbackOnly.Review)
	}
}
