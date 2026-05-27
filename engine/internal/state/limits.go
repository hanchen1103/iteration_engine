package state

import "github.com/hanchen1103/iteration_engine/domain"

func ResolveMaxIterations(requested int, sceneDefault int, fallback int) int {
	if requested > 0 {
		return requested
	}
	if sceneDefault > 0 {
		return sceneDefault
	}
	return fallback
}

func EnsureCanCreateVersion(run *domain.Run) error {
	if run.MaxIterations > 0 && run.VersionCount >= run.MaxIterations {
		return domain.NewError(domain.ErrorCodeConflict, "run has reached max iterations")
	}
	return nil
}

func EnsureCanCreateGeneratedVersion(run *domain.Run) error {
	return EnsureCanCreateVersion(run)
}
