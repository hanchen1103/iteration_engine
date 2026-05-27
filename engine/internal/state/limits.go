package state

import "github.com/hanchen1103/iteration_engine/domain"

func ResolveMaxDepth(requested int, sceneDefault int, fallback int) int {
	if requested > 0 {
		return requested
	}
	if sceneDefault > 0 {
		return sceneDefault
	}
	return fallback
}

func EnsureCanCreateVersion(run *domain.Run) error {
	if run.MaxVersions > 0 && run.VersionCount >= run.MaxVersions {
		return domain.NewError(domain.ErrorCodeConflict, "run has reached max versions")
	}
	return nil
}

func EnsureCanCreateGeneratedVersion(run *domain.Run, base *domain.Version) error {
	if err := EnsureCanCreateVersion(run); err != nil {
		return err
	}
	if base != nil && run.MaxDepth > 0 && base.Depth >= run.MaxDepth {
		return domain.NewError(domain.ErrorCodeConflict, "base version has reached max depth")
	}
	return nil
}
