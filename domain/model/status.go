package model

type RunStatus string

const (
	RunStatusPending       RunStatus = "PENDING"
	RunStatusGenerating    RunStatus = "GENERATING"
	RunStatusReviewing     RunStatus = "REVIEWING"
	RunStatusWaitingManual RunStatus = "WAITING_MANUAL"
	RunStatusSucceeded     RunStatus = "SUCCEEDED"
	RunStatusMaxDepth      RunStatus = "MAX_DEPTH"
	RunStatusFailed        RunStatus = "FAILED"
	RunStatusAdopted       RunStatus = "ADOPTED"
)

func (s RunStatus) IsActive() bool {
	switch s {
	case RunStatusPending, RunStatusGenerating, RunStatusReviewing:
		return true
	default:
		return false
	}
}

func (s RunStatus) IsClosed() bool {
	switch s {
	case RunStatusFailed, RunStatusAdopted:
		return true
	default:
		return false
	}
}

type VersionStatus string

const (
	VersionStatusGenerating VersionStatus = "GENERATING"
	VersionStatusGenerated  VersionStatus = "GENERATED"
	VersionStatusReviewing  VersionStatus = "REVIEWING"
	VersionStatusReviewed   VersionStatus = "REVIEWED"
	VersionStatusEdited     VersionStatus = "EDITED"
	VersionStatusFailed     VersionStatus = "FAILED"
	VersionStatusAdopted    VersionStatus = "ADOPTED"
)

type IterationMode string

const (
	IterationModeAuto   IterationMode = "AUTO"
	IterationModeManual IterationMode = "MANUAL"
)

func NormalizeIterationMode(value IterationMode) IterationMode {
	if value == IterationModeAuto {
		return IterationModeAuto
	}
	return IterationModeManual
}

type ReviewPolicy string

const (
	ReviewPolicyRunDefault   ReviewPolicy = "RUN_DEFAULT"
	ReviewPolicyWaitManual   ReviewPolicy = "WAIT_MANUAL"
	ReviewPolicyAutoContinue ReviewPolicy = "AUTO_CONTINUE"
)

func NormalizeReviewPolicy(value ReviewPolicy) ReviewPolicy {
	switch value {
	case ReviewPolicyWaitManual, ReviewPolicyAutoContinue:
		return value
	default:
		return ReviewPolicyRunDefault
	}
}

type DecisionType string

const (
	DecisionPass         DecisionType = "PASS"
	DecisionAutoContinue DecisionType = "FAIL_AUTO_CONTINUE"
	DecisionWaitManual   DecisionType = "FAIL_WAIT_MANUAL"
	DecisionMaxDepth     DecisionType = "FAIL_MAX_DEPTH"
	DecisionError        DecisionType = "ERROR"
)
