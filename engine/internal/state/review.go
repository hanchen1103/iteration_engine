package state

import (
	"strings"
	"time"

	"github.com/hanchen1103/iteration_engine/domain"
)

func ApplyReviewToVersion(version *domain.Version, review *domain.ReviewResult, now time.Time) {
	version.Status = domain.VersionStatusReviewed
	version.ReviewJSON = domain.CloneRawMessage(review.RawJSON)
	version.ReviewPass = boolPtr(review.Pass)
	version.ReviewScore = cloneFloat64Ptr(review.Score)
	version.ReviewSummary = strings.TrimSpace(review.Summary)
	version.ReviewFeedback = strings.TrimSpace(review.Feedback)
	version.ReviewIssues = cloneReviewIssues(review.Issues)
	version.ReviewExtensions = domain.CloneRawMessageMap(review.Extensions)
	version.ErrorMessage = ""
	version.UpdatedAt = now
}

func ApplyReviewedRunState(run *domain.Run, status domain.RunStatus, review *domain.ReviewResult, now time.Time) {
	run.Status = status
	run.FinalScore = cloneFloat64Ptr(review.Score)
	run.FinalFeedback = strings.TrimSpace(review.Feedback)
	run.ActiveJobID = ""
	run.ActiveRoleKey = ""
	run.ActiveVersionID = ""
	run.ErrorMessage = ""
	run.UpdatedAt = now
}

func ApplyGeneratedRunState(run *domain.Run, version *domain.Version, now time.Time) {
	run.Status = domain.RunStatusWaitingManual
	run.ActiveJobID = ""
	run.ActiveRoleKey = ""
	run.ActiveVersionID = ""
	run.FinalScore = nil
	run.FinalFeedback = ""
	run.ErrorMessage = ""
	run.VersionCount = version.VersionNo
	run.UpdatedAt = now
}

func boolPtr(value bool) *bool {
	return &value
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneReviewIssues(in []domain.ReviewIssue) []domain.ReviewIssue {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.ReviewIssue, len(in))
	copy(out, in)
	return out
}
