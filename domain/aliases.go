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
	RunStatusMaxDepth      = model.RunStatusMaxDepth
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

	PlanSourceInitial    = model.PlanSourceInitial
	PlanSourceManual     = model.PlanSourceManual
	PlanSourceAutoReview = model.PlanSourceAutoReview
	PlanSourceManualEdit = model.PlanSourceManualEdit

	DecisionPass         = model.DecisionPass
	DecisionAutoContinue = model.DecisionAutoContinue
	DecisionWaitManual   = model.DecisionWaitManual
	DecisionMaxDepth     = model.DecisionMaxDepth
	DecisionError        = model.DecisionError

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
