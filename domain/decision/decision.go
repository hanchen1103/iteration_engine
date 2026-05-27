package decision

import "github.com/hanchen1103/iteration_engine/domain/model"

type Decision struct {
	Type    model.DecisionType `json:"type"`
	Message string             `json:"message,omitempty"`
}

func Decide(run *model.Run, version *model.Version, review *model.ReviewResult) Decision {
	if run == nil || version == nil || review == nil {
		return Decision{Type: model.DecisionError, Message: "missing run, version, or review result"}
	}
	if review.Pass {
		return Decision{Type: model.DecisionPass}
	}
	switch model.NormalizeReviewPolicy(version.ReviewPolicy) {
	case model.ReviewPolicyWaitManual:
		return Decision{Type: model.DecisionWaitManual}
	case model.ReviewPolicyAutoContinue:
		if run.MaxDepth > 0 && version.Depth >= run.MaxDepth {
			return Decision{Type: model.DecisionMaxDepth}
		}
		return Decision{Type: model.DecisionAutoContinue}
	default:
		if model.NormalizeIterationMode(run.IterationMode) == model.IterationModeAuto {
			if run.MaxDepth > 0 && version.Depth >= run.MaxDepth {
				return Decision{Type: model.DecisionMaxDepth}
			}
			return Decision{Type: model.DecisionAutoContinue}
		}
		return Decision{Type: model.DecisionWaitManual}
	}
}
