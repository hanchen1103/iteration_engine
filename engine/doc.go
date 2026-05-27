// Package engine provides the orchestration service for role-based iteration.
//
// The package owns state transitions, review decisions, manual edits, manual
// continuation, and adopt lifecycle tracking. Business systems provide
// ports.SceneAdapter implementations, while queues, workers, and optional LLM
// runtimes are hidden behind ports.JobExecutor.
package engine
