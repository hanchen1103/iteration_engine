package engine

import (
	"context"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Service) resolveBaseVersion(ctx context.Context, run *domain.Run, versionID string, versionNo int) (*domain.Version, error) {
	if strings.TrimSpace(versionID) != "" {
		version, err := s.store.GetVersion(ctx, strings.TrimSpace(versionID))
		if err != nil {
			return nil, err
		}
		if version.RunID != run.ID {
			return nil, invalidError("base version does not belong to run")
		}
		return version, nil
	}
	if versionNo > 0 {
		return s.store.GetVersionByRunVersionNo(ctx, run.ID, versionNo)
	}
	versions, err := s.store.ListVersions(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, notFoundError("run has no versions")
	}
	return versions[len(versions)-1], nil
}

func (s *Service) resolveOptionalBaseVersion(ctx context.Context, run *domain.Run, versionID string, versionNo int) (*domain.Version, error) {
	if strings.TrimSpace(versionID) == "" && versionNo <= 0 {
		return nil, nil
	}
	return s.resolveBaseVersion(ctx, run, versionID, versionNo)
}

func (s *Service) runVersionAdapter(ctx context.Context, runID string, versionID string) (*domain.Run, *domain.Version, ports.SceneAdapter, error) {
	run, err := s.getRun(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, nil, nil, err
	}
	version, err := s.store.GetVersion(ctx, strings.TrimSpace(versionID))
	if err != nil {
		return nil, nil, nil, err
	}
	if version.RunID != run.ID {
		return nil, nil, nil, invalidError("version does not belong to run")
	}
	adapter, _, err := s.adapter(run.SceneKey)
	if err != nil {
		return nil, nil, nil, err
	}
	return run, version, adapter, nil
}
