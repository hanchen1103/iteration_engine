package domain

import (
	"encoding/json"

	decisionpkg "github.com/hanchen1103/iteration_engine/domain/decision"
	"github.com/hanchen1103/iteration_engine/domain/model"
	"github.com/hanchen1103/iteration_engine/domain/versiontree"
)

type TargetRef = model.TargetRef
type TargetSnapshot = model.TargetSnapshot
type RuleSpec = model.RuleSpec
type SceneSpec = model.SceneSpec
type SceneCapability = model.SceneCapability

type Run = model.Run
type Version = model.Version
type VersionContent = model.VersionContent
type VersionNode = model.VersionNode
type Artifact = model.Artifact

type ReviewResult = model.ReviewResult
type ReviewIssue = model.ReviewIssue
type GenerateContextOptions = model.GenerateContextOptions
type BaseVersionContextOptions = model.BaseVersionContextOptions
type ReviewContextOptions = model.ReviewContextOptions
type GenerateContext = model.GenerateContext
type BaseVersionContext = model.BaseVersionContext
type BaseVersionMetadata = model.BaseVersionMetadata
type ReviewContext = model.ReviewContext
type GenerateContextOption = model.GenerateContextOption
type IterationPlan = model.IterationPlan
type IterationDirective = model.IterationDirective
type PlanSource = model.PlanSource
type Event = model.Event

type RunStatus = model.RunStatus
type VersionStatus = model.VersionStatus
type IterationMode = model.IterationMode
type ReviewPolicy = model.ReviewPolicy
type DecisionType = model.DecisionType
type Decision = decisionpkg.Decision

type ErrorCode = model.ErrorCode
type Error = model.Error

const (
	RunStatusPending       = model.RunStatusPending
	RunStatusGenerating    = model.RunStatusGenerating
	RunStatusReviewing     = model.RunStatusReviewing
	RunStatusWaitingManual = model.RunStatusWaitingManual
	RunStatusSucceeded     = model.RunStatusSucceeded
	RunStatusMaxIterations = model.RunStatusMaxIterations
	RunStatusFailed        = model.RunStatusFailed
	RunStatusAdopted       = model.RunStatusAdopted

	VersionStatusGenerating = model.VersionStatusGenerating
	VersionStatusGenerated  = model.VersionStatusGenerated
	VersionStatusReviewing  = model.VersionStatusReviewing
	VersionStatusReviewed   = model.VersionStatusReviewed
	VersionStatusEdited     = model.VersionStatusEdited
	VersionStatusFailed     = model.VersionStatusFailed
	VersionStatusAdopted    = model.VersionStatusAdopted

	IterationModeAuto   = model.IterationModeAuto
	IterationModeManual = model.IterationModeManual

	ReviewPolicyRunDefault   = model.ReviewPolicyRunDefault
	ReviewPolicyWaitManual   = model.ReviewPolicyWaitManual
	ReviewPolicyAutoContinue = model.ReviewPolicyAutoContinue

	PlanSourceInitial            = model.PlanSourceInitial
	PlanSourceManual             = model.PlanSourceManual
	PlanSourceAutoReview         = model.PlanSourceAutoReview
	PlanSourceManualEdit         = model.PlanSourceManualEdit
	PlanSourceSubmittedCandidate = model.PlanSourceSubmittedCandidate
	PlanSourceReviewOnly         = model.PlanSourceReviewOnly

	DecisionPass          = model.DecisionPass
	DecisionAutoContinue  = model.DecisionAutoContinue
	DecisionWaitManual    = model.DecisionWaitManual
	DecisionMaxIterations = model.DecisionMaxIterations
	DecisionError         = model.DecisionError

	ErrorCodeInvalid   = model.ErrorCodeInvalid
	ErrorCodeNotFound  = model.ErrorCodeNotFound
	ErrorCodeConflict  = model.ErrorCodeConflict
	ErrorCodeFailed    = model.ErrorCodeFailed
	ErrorCodeForbidden = model.ErrorCodeForbidden

	EventRunCreated        = model.EventRunCreated
	EventGenerateSubmitted = model.EventGenerateSubmitted
	EventGenerateReceived  = model.EventGenerateReceived
	EventReviewSubmitted   = model.EventReviewSubmitted
	EventReviewReceived    = model.EventReviewReceived
	EventManualContinue    = model.EventManualContinue
	EventManualEdit        = model.EventManualEdit
	EventVersionAdopted    = model.EventVersionAdopted
	EventRunFailed         = model.EventRunFailed
)

func NormalizeIterationMode(value IterationMode) IterationMode {
	return model.NormalizeIterationMode(value)
}

func NormalizeReviewPolicy(value ReviewPolicy) ReviewPolicy {
	return model.NormalizeReviewPolicy(value)
}

func NewError(code ErrorCode, message string) error {
	return model.NewError(code, message)
}

func CloneRawMessage(in json.RawMessage) json.RawMessage {
	return model.CloneRawMessage(in)
}

func CloneRawMessageMap(in map[string]json.RawMessage) map[string]json.RawMessage {
	return model.CloneRawMessageMap(in)
}

func CloneArtifacts(in []Artifact) []Artifact {
	return model.CloneArtifacts(in)
}

func CloneDirectives(in []IterationDirective) []IterationDirective {
	return model.CloneDirectives(in)
}

func CloneGenerateContextOptions(in *GenerateContextOptions) *GenerateContextOptions {
	return model.CloneGenerateContextOptions(in)
}

func NewGenerateContextOptions(options ...GenerateContextOption) *GenerateContextOptions {
	return model.NewGenerateContextOptions(options...)
}

func GenerateContextFull() *GenerateContextOptions {
	return model.GenerateContextFull()
}

func GenerateContextNone() *GenerateContextOptions {
	return model.GenerateContextNone()
}

func GenerateContextBaseContentOnly() *GenerateContextOptions {
	return model.GenerateContextBaseContentOnly()
}

func GenerateContextReviewFeedbackOnly() *GenerateContextOptions {
	return model.GenerateContextReviewFeedbackOnly()
}

func WithBaseVersion() GenerateContextOption {
	return model.WithBaseVersion()
}

func WithoutBaseVersion() GenerateContextOption {
	return model.WithoutBaseVersion()
}

func WithBaseMetadata() GenerateContextOption {
	return model.WithBaseMetadata()
}

func WithoutBaseMetadata() GenerateContextOption {
	return model.WithoutBaseMetadata()
}

func WithBaseContent() GenerateContextOption {
	return model.WithBaseContent()
}

func WithoutBaseContent() GenerateContextOption {
	return model.WithoutBaseContent()
}

func WithBaseArtifacts() GenerateContextOption {
	return model.WithBaseArtifacts()
}

func WithoutBaseArtifacts() GenerateContextOption {
	return model.WithoutBaseArtifacts()
}

func WithReview() GenerateContextOption {
	return model.WithReview()
}

func WithoutReview() GenerateContextOption {
	return model.WithoutReview()
}

func WithReviewPass() GenerateContextOption {
	return model.WithReviewPass()
}

func WithoutReviewPass() GenerateContextOption {
	return model.WithoutReviewPass()
}

func WithReviewScore() GenerateContextOption {
	return model.WithReviewScore()
}

func WithoutReviewScore() GenerateContextOption {
	return model.WithoutReviewScore()
}

func WithReviewSummary() GenerateContextOption {
	return model.WithReviewSummary()
}

func WithoutReviewSummary() GenerateContextOption {
	return model.WithoutReviewSummary()
}

func WithReviewFeedback() GenerateContextOption {
	return model.WithReviewFeedback()
}

func WithoutReviewFeedback() GenerateContextOption {
	return model.WithoutReviewFeedback()
}

func WithReviewIssues() GenerateContextOption {
	return model.WithReviewIssues()
}

func WithoutReviewIssues() GenerateContextOption {
	return model.WithoutReviewIssues()
}

func WithReviewExtensions(keys ...string) GenerateContextOption {
	return model.WithReviewExtensions(keys...)
}

func WithoutReviewExtensions() GenerateContextOption {
	return model.WithoutReviewExtensions()
}

func WithReviewRawJSON() GenerateContextOption {
	return model.WithReviewRawJSON()
}

func WithoutReviewRawJSON() GenerateContextOption {
	return model.WithoutReviewRawJSON()
}

func BoolDirective(key string, value bool, description string) IterationDirective {
	return model.BoolDirective(key, value, description)
}

func IntDirective(key string, value int, description string) IterationDirective {
	return model.IntDirective(key, value, description)
}

func StringDirective(key string, value string, description string) IterationDirective {
	return model.StringDirective(key, value, description)
}

func ValueDirective(key string, value any, description string) (IterationDirective, error) {
	return model.ValueDirective(key, value, description)
}

func ObjectDirective(key string, value any, description string) (IterationDirective, error) {
	return model.ObjectDirective(key, value, description)
}

func Decide(run *Run, version *Version, review *ReviewResult) Decision {
	return decisionpkg.Decide(run, version, review)
}

func BuildVersionTree(versions []*Version) ([]*VersionNode, error) {
	return versiontree.Build(versions)
}
