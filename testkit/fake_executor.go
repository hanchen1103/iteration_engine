package testkit

import (
	"context"
	"fmt"

	"github.com/hanchen1103/iteration_engine/ports"
)

type Executor struct {
	next int
	Jobs []ports.JobRequest
}

func (e *Executor) Submit(ctx context.Context, req *ports.JobRequest) (*ports.JobHandle, error) {
	_ = ctx
	e.next++
	jobID := fmt.Sprintf("job_%d", e.next)
	e.Jobs = append(e.Jobs, *req)
	return &ports.JobHandle{JobID: jobID}, nil
}
