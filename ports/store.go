package ports

import (
	"context"

	"github.com/hanchen1103/iteration_engine/domain"
)

type Store interface {
	CreateRun(ctx context.Context, run *domain.Run) error
	UpdateRun(ctx context.Context, run *domain.Run) error
	GetRun(ctx context.Context, runID string) (*domain.Run, error)
	ListRuns(ctx context.Context, filter ListRunsFilter) ([]*domain.Run, error)

	CreateVersion(ctx context.Context, version *domain.Version) error
	UpdateVersion(ctx context.Context, version *domain.Version) error
	GetVersion(ctx context.Context, versionID string) (*domain.Version, error)
	GetVersionByRunVersionNo(ctx context.Context, runID string, versionNo int) (*domain.Version, error)
	ListVersions(ctx context.Context, runID string) ([]*domain.Version, error)
	FindVersionByJobID(ctx context.Context, jobID string) (*domain.Version, JobRole, error)

	RecordEvent(ctx context.Context, event *domain.Event) error
	ListEvents(ctx context.Context, runID string) ([]*domain.Event, error)
}

type ListRunsFilter struct {
	SceneKey string
	Target   domain.TargetRef
	Status   domain.RunStatus
	Limit    int
	Offset   int
}

type JobRole string

const (
	JobRoleGenerate JobRole = "generate"
	JobRoleReview   JobRole = "review"
)
