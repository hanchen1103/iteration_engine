package memory

import (
	"context"
	"fmt"
	"sort"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

func (s *Store) CreateVersion(ctx context.Context, version *domain.Version) error {
	_ = ctx
	if version == nil {
		return invalidError("version is nil")
	}
	if version.RunID == "" {
		return invalidError("version run id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[version.RunID]; !ok {
		return notFoundError("run not found")
	}
	for _, existing := range s.versions {
		if existing.RunID == version.RunID && existing.VersionNo == version.VersionNo {
			return conflictError("version number already exists")
		}
	}
	if version.VersionNo <= 0 {
		return invalidError("version number is required")
	}
	if version.Depth <= 0 {
		return invalidError("version depth is required")
	}
	if version.BaseVersionID != "" {
		base, ok := s.versions[version.BaseVersionID]
		if !ok || base.RunID != version.RunID {
			return notFoundError("base version not found")
		}
		if version.Depth != base.Depth+1 {
			return invalidError("version depth must be base depth plus one")
		}
	}
	if version.ID == "" {
		s.nextVersionID++
		version.ID = fmt.Sprintf("ver_%d", s.nextVersionID)
	}
	if _, ok := s.versions[version.ID]; ok {
		return conflictError("version already exists")
	}
	s.versions[version.ID] = cloneVersion(version)
	return nil
}

func (s *Store) UpdateVersion(ctx context.Context, version *domain.Version) error {
	_ = ctx
	if version == nil || version.ID == "" {
		return invalidError("version id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.versions[version.ID]; !ok {
		return notFoundError("version not found")
	}
	s.versions[version.ID] = cloneVersion(version)
	return nil
}

func (s *Store) GetVersion(ctx context.Context, versionID string) (*domain.Version, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	version, ok := s.versions[versionID]
	if !ok {
		return nil, notFoundError("version not found")
	}
	return cloneVersion(version), nil
}

func (s *Store) GetVersionByRunVersionNo(ctx context.Context, runID string, versionNo int) (*domain.Version, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, version := range s.versions {
		if version.RunID == runID && version.VersionNo == versionNo {
			return cloneVersion(version), nil
		}
	}
	return nil, notFoundError("version not found")
}

func (s *Store) ListVersions(ctx context.Context, runID string) ([]*domain.Version, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []*domain.Version{}
	for _, version := range s.versions {
		if version.RunID == runID {
			items = append(items, cloneVersion(version))
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].VersionNo < items[j].VersionNo
	})
	return items, nil
}

func (s *Store) FindVersionByJobID(ctx context.Context, jobID string) (*domain.Version, ports.JobRole, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, version := range s.versions {
		if version.GenerateJobID == jobID {
			return cloneVersion(version), ports.JobRoleGenerate, nil
		}
		if version.ReviewJobID == jobID {
			return cloneVersion(version), ports.JobRoleReview, nil
		}
	}
	return nil, "", notFoundError("version job not found")
}
