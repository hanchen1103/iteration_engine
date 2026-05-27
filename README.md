# Iteration Engine

`iteration_engine` is a small Go component for role-based content iteration.

## Layout

```text
iteration_engine/
  domain/
    aliases.go        Stable facade for callers.
    model/            Run, Version, status, scene, review, plan, event.
    decision/         Review decision policy.
    versiontree/      Builds tree views from flat version rows.
  ports/
    scene_adapter.go, executor.go, scene_registry.go, store.go
  engine/
    service.go
    run_service.go, version_service.go, result_callbacks.go
    generate_workflow.go, review_workflow.go
    review_transitions.go, lookup.go, validation.go, failure.go
    internal/state/   Pure state mutation helpers used by the service.
  store/memory/
    store.go, run_store.go, version_store.go, event_store.go
  store/sql/mysql/
    schema.sql       Recommended production table schema.
  docs/
    mysql_store.md   Production Store mapping and migration notes.
  testkit/
    fake_adapter.go, fake_executor.go, json.go
```

It owns the generic lifecycle:

- run tracking
- version tree tracking
- generate/review job submission
- review result normalization
- decision after review
- manual continue
- manual edit through effective content
- adopt lifecycle
- event recording

It does not own business content, prompt text, model routing, LLM calls, or business writes.
Business systems integrate through `ports.SceneAdapter`; execution backends integrate through
`ports.JobExecutor`.

## Core Boundary

The `domain` package knows:

- `Run`
- `Version`
- `VersionNode`
- `IterationPlan`
- `IterationDirective`
- `ReviewResult`
- `Decision`
- `Event`

The `engine` package knows:

- how to create/start/continue runs
- how to move versions through generate/review/edit/adopt states
- how to turn review results into the next lifecycle decision
- when to submit generate/review jobs through `ports.JobExecutor`

The adapter knows:

- how to load a business target
- how to build generate/review jobs
- how to parse raw role output
- how to adopt a version into the business system

The executor knows:

- how to submit a `JobRequest`
- how to return a `JobHandle`

Callbacks are intentionally explicit:

- `ReceiveGenerateResult`
- `ReceiveReviewResult`

This keeps v0 independent from any specific queue, worker, LLM platform, or HTTP callback shape.

`MANUAL` runs stop after generate and enter `WAITING_MANUAL`. `AUTO` runs dispatch review after
generate and may continue until pass or max iterations.

`GetRunDetail` returns both a flat `Versions` list sorted by `VersionNo` and a `VersionTree`
view for callers that need branch-aware UI or selection.

## Stable v0 Decisions

- IDs are strings, so storage can use database IDs, UUIDs, or external IDs.
- Versions form a tree. `Version.BaseVersionID` is the parent pointer, `Version.VersionNo` is the run-local creation order, and `Version.Depth` is the node depth.
- `Run.MaxIterations` limits the total candidate versions a run can create. If neither request nor scene sets it, the engine defaults to 50.
- Auto continue is disabled by default. Construct the service with `engine.WithAutoContinue()` before accepting `IterationModeAuto` or `ReviewPolicyAutoContinue`.
- `IterationModeManual` means generate a candidate and wait. It does not auto-review.
- `IterationModeAuto` means generate, review, and continue on failed review when auto is enabled.
- Version content is `json.RawMessage`; the engine does not inspect business payloads.
- `IterationPlan.Source` is a typed string with stable built-ins: `PlanSourceInitial`, `PlanSourceManual`, `PlanSourceAutoReview`, `PlanSourceManualEdit`, `PlanSourceSubmittedCandidate`, and `PlanSourceReviewOnly`.
- `SubmitCandidateForReview` creates a version from caller-provided content and dispatches review without a generate job.
- `ReviewVersion` creates a new review-only child version from the selected version content, so re-review history is preserved.
- Manual edits overwrite the selected version's generated content in place and clear stale review fields.
- `SUCCEEDED` means a version passed review. `ADOPTED` means a business adapter committed it.
- Manual `ReviewVersion` defaults failed reviews to `WAITING_MANUAL`, even for auto runs.
- `ReviewResult.Extensions` carries structured business-specific review fields. `ReviewResult.RawJSON` keeps the full raw review output for audit/debug.
- `IterationDirective.Value` must be a JSON object and directive keys cannot repeat in one plan.
- `domain.IntDirective`, `domain.BoolDirective`, and `domain.StringDirective` wrap scalar directive values as `{"value": ...}`. Use `domain.ObjectDirective` when a directive needs a custom object shape.

## Minimal Use

```go
import (
    iterengine "github.com/hanchen1103/iteration_engine/engine"
    "github.com/hanchen1103/iteration_engine/domain"
    "github.com/hanchen1103/iteration_engine/ports"
    "github.com/hanchen1103/iteration_engine/store/memory"
)

store := memory.NewStore()
executor := myExecutor{}
registry := ports.NewSceneRegistry(myAdapter{})
service := iterengine.NewService(store, executor, registry)

// Auto continue is deliberately opt-in:
// service := iterengine.NewService(store, executor, registry, iterengine.WithAutoContinue())

run, _ := service.CreateRun(ctx, iterengine.CreateRunRequest{
    SceneKey: "my_scene",
    Target: domain.TargetRef{Type: "article", ID: "123"},
    IterationMode: domain.IterationModeManual,
    MaxIterations: 3,
    DefaultDirectives: []domain.IterationDirective{
        domain.IntDirective("difficulty", 1, "1 means harder, -1 means easier."),
    },
})

version, _ := service.StartRun(ctx, run.ID)

// Later, from the job backend callback:
_ = service.ReceiveGenerateResult(ctx, iterengine.GenerateResultRequest{
    JobID: version.GenerateJobID,
    Raw: rawGenerateOutput,
})
```

The in-memory store is for tests and local wiring. Production services should implement
`ports.Store` with their own database and transaction policy. A recommended MySQL
baseline is provided in `store/sql/mysql/schema.sql`; see `docs/mysql_store.md`.
