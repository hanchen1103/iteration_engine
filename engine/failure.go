package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
)

func (s *Service) failRunVersion(ctx context.Context, run *domain.Run, version *domain.Version, message string) error {
	now := s.now()
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = "iteration engine operation failed"
	}
	if version != nil {
		version.Status = domain.VersionStatusFailed
		version.ErrorMessage = msg
		version.UpdatedAt = now
		if err := s.store.UpdateVersion(ctx, version); err != nil {
			return err
		}
	}
	if run != nil {
		run.Status = domain.RunStatusFailed
		run.ErrorMessage = msg
		run.ActiveJobID = ""
		run.ActiveRoleKey = ""
		run.ActiveVersionID = ""
		run.UpdatedAt = now
		if err := s.store.UpdateRun(ctx, run); err != nil {
			return err
		}
		s.recordEvent(ctx, domain.EventRunFailed, run.ID, "", "", msg, nil)
	}
	return failedError(msg)
}
