-- Recommended MySQL schema for a production iteration_engine Store.
--
-- The engine library does not run this file automatically. Apply it through the
-- owning service's normal migration system, then implement ports.Store against
-- these tables.
--
-- Notes:
-- - Domain IDs are strings. A DB Store should generate run/version/event IDs
--   before insert when the engine passes an empty ID.
-- - Optional string fields should be stored as NULL and mapped back to "" in
--   domain structs.
-- - Foreign keys are intentionally omitted so services can use their normal
--   online DDL, sharding, and retention policies. Enforce integrity in Store
--   code and keep the indexes below.

CREATE TABLE IF NOT EXISTS iteration_engine_runs (
    id VARCHAR(64) NOT NULL,
    scene_key VARCHAR(128) NOT NULL,
    target_type VARCHAR(128) NOT NULL,
    target_id VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    iteration_mode VARCHAR(32) NOT NULL,
    max_iterations INT NOT NULL,
    version_count INT NOT NULL DEFAULT 0,
    adopted_version_id VARCHAR(64) NULL,
    config JSON NULL,
    generate_context JSON NULL,
    default_directives JSON NULL,
    rule_set_snapshot JSON NULL,
    final_score DOUBLE NULL,
    final_feedback TEXT NULL,
    error_message TEXT NULL,
    active_version_id VARCHAR(64) NULL,
    active_job_id VARCHAR(128) NULL,
    active_role_key VARCHAR(128) NULL,
    created_by VARCHAR(128) NULL,
    updated_by VARCHAR(128) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_iteration_runs_scene_status_created (scene_key, status, created_at),
    KEY idx_iteration_runs_target_created (target_type, target_id, created_at),
    KEY idx_iteration_runs_active_job (active_job_id),
    KEY idx_iteration_runs_adopted_version (adopted_version_id),
    KEY idx_iteration_runs_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS iteration_engine_versions (
    id VARCHAR(64) NOT NULL,
    run_id VARCHAR(64) NOT NULL,
    version_no INT NOT NULL,
    base_version_id VARCHAR(64) NULL,
    depth INT NOT NULL,
    status VARCHAR(32) NOT NULL,
    review_policy VARCHAR(32) NULL,
    iteration_plan JSON NOT NULL,
    target_snapshot JSON NULL,
    generate_rule_snapshot JSON NOT NULL,
    review_rule_snapshot JSON NOT NULL,
    generate_attempt_no INT NOT NULL DEFAULT 0,
    review_attempt_no INT NOT NULL DEFAULT 0,
    generate_job_id VARCHAR(128) NULL,
    review_job_id VARCHAR(128) NULL,
    generate_input_json JSON NULL,
    generated_content JSON NULL,
    generated_artifacts JSON NULL,
    review_input_json JSON NULL,
    review_json JSON NULL,
    review_pass TINYINT(1) NULL,
    review_score DOUBLE NULL,
    review_summary TEXT NULL,
    review_feedback TEXT NULL,
    review_issues JSON NULL,
    review_extensions JSON NULL,
    edited_content JSON NULL,
    edited_artifacts JSON NULL,
    edited_by VARCHAR(128) NULL,
    edited_at DATETIME(6) NULL,
    error_message TEXT NULL,
    created_by VARCHAR(128) NULL,
    updated_by VARCHAR(128) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_iteration_versions_run_version_no (run_id, version_no),
    UNIQUE KEY uk_iteration_versions_generate_job (generate_job_id),
    UNIQUE KEY uk_iteration_versions_review_job (review_job_id),
    KEY idx_iteration_versions_run_created (run_id, created_at),
    KEY idx_iteration_versions_run_status (run_id, status),
    KEY idx_iteration_versions_base (base_version_id),
    KEY idx_iteration_versions_depth (run_id, depth),
    KEY idx_iteration_versions_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS iteration_engine_events (
    id VARCHAR(64) NOT NULL,
    run_id VARCHAR(64) NOT NULL,
    version_id VARCHAR(64) NULL,
    type VARCHAR(64) NOT NULL,
    actor VARCHAR(128) NULL,
    message TEXT NULL,
    payload JSON NULL,
    created_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_iteration_events_run_created (run_id, created_at, id),
    KEY idx_iteration_events_version_created (version_id, created_at),
    KEY idx_iteration_events_type_created (type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
