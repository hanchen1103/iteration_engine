# MySQL Store

This repository defines the `ports.Store` contract but does not create database
tables at runtime. Production services should apply the schema through their own
migration system, then implement `ports.Store` against those tables.

The recommended baseline schema is:

```text
store/sql/mysql/schema.sql
```

It creates three tables:

- `iteration_engine_runs`: one row per run.
- `iteration_engine_versions`: one row per version node. `base_version_id` is
  the parent pointer, so versions form a tree.
- `iteration_engine_events`: append-only lifecycle events for audit and UI.

## Ownership

The engine owns lifecycle semantics:

- run status
- version status
- version tree parent/depth
- active generate/review job
- review decision state
- adoption state

The owning service owns deployment:

- running the migration
- choosing the concrete DB/schema
- generating string IDs
- transaction and retry policy
- retention and backup policy

The engine should not be granted runtime DDL permissions.

## Store Mapping

`domain.Run` maps to `iteration_engine_runs`.

Important columns:

- `scene_key`: used to resolve the registered `SceneAdapter` after restart.
- `target_type`, `target_id`: the business target reference.
- `config`: scene-specific runtime config, such as model routing or seed input.
- `generate_context`: run-level defaults for which base-version fields and
  review fields a generate job receives.
- `default_directives`: default structured controls for generated versions.
- `rule_set_snapshot`: adapter `SceneSpec` captured when the run was created.
- `max_iterations`: maximum candidate versions the run can create before auto
  continuation stops.
- `active_job_id`: current generate/review job for callback validation.
- `version_count`: next version number is based on this value.

`domain.Version` maps to `iteration_engine_versions`.

Important columns:

- `run_id`, `version_no`: unique run-local version identity.
- `base_version_id`: parent version ID for branch-aware continuation.
- `depth`: tree depth; root versions start at 1.
- `iteration_plan`: source, instruction, explanation, and directives.
- `generate_job_id`, `review_job_id`: used by callback handlers to find a
  version from a job ID.
- `generated_content`: candidate content used for review and adopt. Manual edits
  overwrite this field in place.
- `edited_content`: legacy/optional compatibility field; the current engine edit
  path does not populate it.
- `review_json`: raw normalized review JSON for audit/debug.
- `review_extensions`: scene-specific structured review fields.

`domain.Event` maps to `iteration_engine_events`.

Events are append-only and should be listed by `(run_id, created_at, id)`.

## Null Handling

The domain model uses empty strings and nil slices/maps. The SQL schema uses
`NULL` for optional columns. A DB Store should convert:

- SQL `NULL` string columns to `""`
- SQL `NULL` JSON columns to nil `json.RawMessage`
- SQL `NULL` booleans/numbers to nil pointers where the domain type uses a
  pointer

For nullable job IDs, store `NULL` instead of `""`; unique indexes on nullable
job columns depend on this behavior.

## Callback Recovery

After a service restart, `ReceiveGenerateResult` and `ReceiveReviewResult` can
continue from persisted state because the Store can resolve callbacks through:

- `FindVersionByJobID(generate_job_id)`
- `FindVersionByJobID(review_job_id)`

The run row must also preserve `active_job_id`, `active_version_id`, and
`active_role_key`. The engine ignores stale callbacks when the active job no
longer matches.

## Listening Generate

For `listening_generate`, one shared adapter can serve all practices. Practice
specific data should live in:

- `Run.Target`: listening practice ID
- `Run.Config`: model/api key/topic/seed/initial passage
- `Version.GeneratedContent`: generated candidate JSON
- `Version.ReviewExtensions`: fields such as `feedback_items`,
  `difficulty_profile`, and `manual_controls`

The adapter should stay stateless or thread-safe. It may hold shared DAO,
prompt-runtime, and schema-loader dependencies, but it should not hold
per-practice mutable state.

## Implementation Notes

Recommended DB Store behavior:

- Generate `Run.ID`, `Version.ID`, and `Event.ID` when the engine passes them as
  empty strings.
- Enforce `CreateVersion` invariants before insert:
  - `run_id` exists
  - `(run_id, version_no)` is unique
  - `depth > 0`
  - if `base_version_id` is set, it belongs to the same run and `depth` is
    parent depth plus one
- Implement `ListVersions(runID)` ordered by `version_no ASC`.
- Implement `ListRuns(filter)` ordered by `created_at DESC`.
- Implement `ListEvents(runID)` ordered by `created_at ASC, id ASC`.
- Keep writes idempotent at the service boundary where possible; v0 does not
  include distributed locks or cross-service idempotency keys.
