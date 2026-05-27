package model

import (
	"encoding/json"
	"strings"
)

type GenerateContextOptions struct {
	BaseVersion BaseVersionContextOptions `json:"baseVersion,omitempty"`
	Review      ReviewContextOptions      `json:"review,omitempty"`
}

type BaseVersionContextOptions struct {
	IncludeMetadata  *bool `json:"includeMetadata,omitempty"`
	IncludeContent   *bool `json:"includeContent,omitempty"`
	IncludeArtifacts *bool `json:"includeArtifacts,omitempty"`
}

type ReviewContextOptions struct {
	IncludePass       *bool     `json:"includePass,omitempty"`
	IncludeScore      *bool     `json:"includeScore,omitempty"`
	IncludeSummary    *bool     `json:"includeSummary,omitempty"`
	IncludeFeedback   *bool     `json:"includeFeedback,omitempty"`
	IncludeIssues     *bool     `json:"includeIssues,omitempty"`
	IncludeExtensions *bool     `json:"includeExtensions,omitempty"`
	ExtensionKeys     *[]string `json:"extensionKeys,omitempty"`
	IncludeRawJSON    *bool     `json:"includeRawJson,omitempty"`
}

type GenerateContext struct {
	BaseVersion    *BaseVersionContext `json:"baseVersion,omitempty"`
	PreviousReview *ReviewContext      `json:"previousReview,omitempty"`
}

type BaseVersionContext struct {
	Metadata  *BaseVersionMetadata `json:"metadata,omitempty"`
	Content   json.RawMessage      `json:"content,omitempty"`
	Artifacts []Artifact           `json:"artifacts,omitempty"`
}

type BaseVersionMetadata struct {
	ID            string        `json:"id"`
	RunID         string        `json:"runID"`
	VersionNo     int           `json:"versionNo"`
	BaseVersionID string        `json:"baseVersionID,omitempty"`
	Depth         int           `json:"depth"`
	Status        VersionStatus `json:"status"`
}

type ReviewContext struct {
	Pass       *bool                      `json:"pass,omitempty"`
	Score      *float64                   `json:"score,omitempty"`
	Summary    string                     `json:"summary,omitempty"`
	Feedback   string                     `json:"feedback,omitempty"`
	Issues     []ReviewIssue              `json:"issues,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
	RawJSON    json.RawMessage            `json:"rawJson,omitempty"`
}

type GenerateContextOption func(*GenerateContextOptions)

func NewGenerateContextOptions(options ...GenerateContextOption) *GenerateContextOptions {
	out := &GenerateContextOptions{}
	for _, option := range options {
		if option != nil {
			option(out)
		}
	}
	return out
}

func GenerateContextFull() *GenerateContextOptions {
	return NewGenerateContextOptions()
}

func GenerateContextNone() *GenerateContextOptions {
	return NewGenerateContextOptions(WithoutBaseVersion(), WithoutReview())
}

func GenerateContextBaseContentOnly() *GenerateContextOptions {
	return NewGenerateContextOptions(WithoutBaseVersion(), WithBaseContent(), WithoutReview())
}

func GenerateContextReviewFeedbackOnly() *GenerateContextOptions {
	return NewGenerateContextOptions(WithoutBaseVersion(), WithoutReview(), WithReviewFeedback())
}

func WithBaseVersion() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeMetadata = boolPtr(true)
		options.BaseVersion.IncludeContent = boolPtr(true)
		options.BaseVersion.IncludeArtifacts = boolPtr(true)
	}
}

func WithoutBaseVersion() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeMetadata = boolPtr(false)
		options.BaseVersion.IncludeContent = boolPtr(false)
		options.BaseVersion.IncludeArtifacts = boolPtr(false)
	}
}

func WithBaseMetadata() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeMetadata = boolPtr(true)
	}
}

func WithoutBaseMetadata() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeMetadata = boolPtr(false)
	}
}

func WithBaseContent() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeContent = boolPtr(true)
	}
}

func WithoutBaseContent() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeContent = boolPtr(false)
	}
}

func WithBaseArtifacts() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeArtifacts = boolPtr(true)
	}
}

func WithoutBaseArtifacts() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.BaseVersion.IncludeArtifacts = boolPtr(false)
	}
}

func WithReview() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludePass = boolPtr(true)
		options.Review.IncludeScore = boolPtr(true)
		options.Review.IncludeSummary = boolPtr(true)
		options.Review.IncludeFeedback = boolPtr(true)
		options.Review.IncludeIssues = boolPtr(true)
		options.Review.IncludeExtensions = boolPtr(true)
		options.Review.ExtensionKeys = nil
		options.Review.IncludeRawJSON = boolPtr(true)
	}
}

func WithoutReview() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludePass = boolPtr(false)
		options.Review.IncludeScore = boolPtr(false)
		options.Review.IncludeSummary = boolPtr(false)
		options.Review.IncludeFeedback = boolPtr(false)
		options.Review.IncludeIssues = boolPtr(false)
		options.Review.IncludeExtensions = boolPtr(false)
		options.Review.ExtensionKeys = nil
		options.Review.IncludeRawJSON = boolPtr(false)
	}
}

func WithReviewPass() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludePass = boolPtr(true)
	}
}

func WithoutReviewPass() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludePass = boolPtr(false)
	}
}

func WithReviewScore() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeScore = boolPtr(true)
	}
}

func WithoutReviewScore() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeScore = boolPtr(false)
	}
}

func WithReviewSummary() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeSummary = boolPtr(true)
	}
}

func WithoutReviewSummary() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeSummary = boolPtr(false)
	}
}

func WithReviewFeedback() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeFeedback = boolPtr(true)
	}
}

func WithoutReviewFeedback() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeFeedback = boolPtr(false)
	}
}

func WithReviewIssues() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeIssues = boolPtr(true)
	}
}

func WithoutReviewIssues() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeIssues = boolPtr(false)
	}
}

func WithReviewExtensions(keys ...string) GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeExtensions = boolPtr(true)
		options.Review.ExtensionKeys = stringSlicePtr(normalizeStringKeys(keys))
	}
}

func WithoutReviewExtensions() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeExtensions = boolPtr(false)
		options.Review.ExtensionKeys = nil
	}
}

func WithReviewRawJSON() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeRawJSON = boolPtr(true)
	}
}

func WithoutReviewRawJSON() GenerateContextOption {
	return func(options *GenerateContextOptions) {
		options.Review.IncludeRawJSON = boolPtr(false)
	}
}

func (o BaseVersionContextOptions) ShouldIncludeMetadata() bool {
	return boolDefaultTrue(o.IncludeMetadata)
}

func (o BaseVersionContextOptions) ShouldIncludeContent() bool {
	return boolDefaultTrue(o.IncludeContent)
}

func (o BaseVersionContextOptions) ShouldIncludeArtifacts() bool {
	return boolDefaultTrue(o.IncludeArtifacts)
}

func (o BaseVersionContextOptions) IncludesAny() bool {
	return o.ShouldIncludeMetadata() || o.ShouldIncludeContent() || o.ShouldIncludeArtifacts()
}

func (o ReviewContextOptions) ShouldIncludePass() bool {
	return boolDefaultTrue(o.IncludePass)
}

func (o ReviewContextOptions) ShouldIncludeScore() bool {
	return boolDefaultTrue(o.IncludeScore)
}

func (o ReviewContextOptions) ShouldIncludeSummary() bool {
	return boolDefaultTrue(o.IncludeSummary)
}

func (o ReviewContextOptions) ShouldIncludeFeedback() bool {
	return boolDefaultTrue(o.IncludeFeedback)
}

func (o ReviewContextOptions) ShouldIncludeIssues() bool {
	return boolDefaultTrue(o.IncludeIssues)
}

func (o ReviewContextOptions) ShouldIncludeExtensions() bool {
	return boolDefaultTrue(o.IncludeExtensions)
}

func (o ReviewContextOptions) ShouldIncludeRawJSON() bool {
	return boolDefaultTrue(o.IncludeRawJSON)
}

func (o ReviewContextOptions) IncludesAny() bool {
	return o.ShouldIncludePass() ||
		o.ShouldIncludeScore() ||
		o.ShouldIncludeSummary() ||
		o.ShouldIncludeFeedback() ||
		o.ShouldIncludeIssues() ||
		o.ShouldIncludeExtensions() ||
		o.ShouldIncludeRawJSON()
}

func (o ReviewContextOptions) IncludesReviewGuidance() bool {
	return o.ShouldIncludeSummary() ||
		o.ShouldIncludeFeedback() ||
		o.ShouldIncludeIssues() ||
		o.ShouldIncludeExtensions() ||
		o.ShouldIncludeRawJSON()
}

func CloneGenerateContextOptions(in *GenerateContextOptions) *GenerateContextOptions {
	if in == nil {
		return nil
	}
	out := &GenerateContextOptions{}
	out.BaseVersion = cloneBaseVersionContextOptions(in.BaseVersion)
	out.Review = cloneReviewContextOptions(in.Review)
	return out
}

func cloneBaseVersionContextOptions(in BaseVersionContextOptions) BaseVersionContextOptions {
	return BaseVersionContextOptions{
		IncludeMetadata:  cloneBoolPtr(in.IncludeMetadata),
		IncludeContent:   cloneBoolPtr(in.IncludeContent),
		IncludeArtifacts: cloneBoolPtr(in.IncludeArtifacts),
	}
}

func cloneReviewContextOptions(in ReviewContextOptions) ReviewContextOptions {
	return ReviewContextOptions{
		IncludePass:       cloneBoolPtr(in.IncludePass),
		IncludeScore:      cloneBoolPtr(in.IncludeScore),
		IncludeSummary:    cloneBoolPtr(in.IncludeSummary),
		IncludeFeedback:   cloneBoolPtr(in.IncludeFeedback),
		IncludeIssues:     cloneBoolPtr(in.IncludeIssues),
		IncludeExtensions: cloneBoolPtr(in.IncludeExtensions),
		ExtensionKeys:     cloneStringSlicePtr(in.ExtensionKeys),
		IncludeRawJSON:    cloneBoolPtr(in.IncludeRawJSON),
	}
}

func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneStringSlicePtr(value *[]string) *[]string {
	if value == nil {
		return nil
	}
	cloned := make([]string, len(*value))
	copy(cloned, *value)
	return &cloned
}

func boolPtr(value bool) *bool {
	return &value
}

func stringSlicePtr(value []string) *[]string {
	if len(value) == 0 {
		return nil
	}
	cloned := make([]string, len(value))
	copy(cloned, value)
	return &cloned
}

func normalizeStringKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out = append(out, key)
	}
	return out
}

func boolDefaultTrue(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}
