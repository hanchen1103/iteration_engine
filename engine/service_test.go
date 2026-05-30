package engine_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hanchen1103/iteration_engine/domain"
	engine "github.com/hanchen1103/iteration_engine/engine"
	"github.com/hanchen1103/iteration_engine/ports"
	"github.com/hanchen1103/iteration_engine/store/memory"
	"github.com/hanchen1103/iteration_engine/testkit"
)

func TestAutoContinueThenPass(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	executor := &testkit.Executor{}
	adapter := &testkit.Adapter{CanAuto: true}
	service := engine.NewService(store, executor, ports.NewSceneRegistry(adapter), engine.WithAutoContinue())

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeAuto,
		MaxIterations: 2,
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	first, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if first.VersionNo != 1 || first.Depth != 1 || first.GenerateJobID == "" {
		t.Fatalf("unexpected first version: %#v", first)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: first.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 1"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	reviewJobID := detail.Versions[0].ReviewJobID
	if reviewJobID == "" || detail.Run.Status != domain.RunStatusReviewing {
		t.Fatalf("expected review job, got run=%#v version=%#v", detail.Run, detail.Versions[0])
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: reviewJobID,
		Raw:   testkit.RawJSON(`{"pass":false,"score":4.5,"feedback":"needs work","difficulty_profile":{"level":"too_easy"}}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult fail review returned error: %v", err)
	}
	detail, err = service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if detail.Run.Status != domain.RunStatusGenerating || detail.Run.VersionCount != 2 || len(detail.Versions) != 2 {
		t.Fatalf("expected auto-generated second version, got run=%#v versions=%d", detail.Run, len(detail.Versions))
	}
	second := detail.Versions[1]
	first = detail.Versions[0]
	if string(first.ReviewExtensions["difficulty_profile"]) != `{"level":"too_easy"}` {
		t.Fatalf("review extension was not stored: %#v", first.ReviewExtensions)
	}
	var generateInput struct {
		Context domain.GenerateContext `json:"context"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &generateInput); err != nil {
		t.Fatalf("failed to parse next generate input: %v", err)
	}
	if generateInput.Context.PreviousReview == nil || string(generateInput.Context.PreviousReview.Extensions["difficulty_profile"]) != `{"level":"too_easy"}` {
		t.Fatalf("review extension was not passed to next generate: %#v", generateInput.Context.PreviousReview)
	}
	if generateInput.Context.BaseVersion == nil || string(generateInput.Context.BaseVersion.Content) != `{"text":"draft 1"}` {
		t.Fatalf("base content was not passed to next generate: %#v", generateInput.Context.BaseVersion)
	}
	if second.BaseVersionID != first.ID || second.VersionNo != 2 || second.Depth != 2 || second.IterationPlan.Source != domain.PlanSourceAutoReview {
		t.Fatalf("unexpected auto plan: %#v", second)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: second.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 2"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult second returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: detail.Versions[1].ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true,"score":9.2,"feedback":"ready"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult pass returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusSucceeded || detail.Run.FinalScore == nil || *detail.Run.FinalScore != 9.2 {
		t.Fatalf("expected succeeded run with final score, got %#v", detail.Run)
	}
}

func TestAutoContinueCanOmitPreviousReviewContext(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	executor := &testkit.Executor{}
	adapter := &testkit.Adapter{CanAuto: true}
	service := engine.NewService(store, executor, ports.NewSceneRegistry(adapter), engine.WithAutoContinue())

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:        "fake",
		Target:          domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode:   domain.IterationModeAuto,
		MaxIterations:   2,
		GenerateContext: domain.NewGenerateContextOptions(domain.WithoutReview()),
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	first, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: first.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 1"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: detail.Versions[0].ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":false,"score":4.5,"feedback":"needs work"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}

	detail, err = service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if detail.Run.Status != domain.RunStatusGenerating || len(detail.Versions) != 2 {
		t.Fatalf("expected auto-generated second version, got run=%#v versions=%d", detail.Run, len(detail.Versions))
	}
	var generateInput struct {
		Context domain.GenerateContext `json:"context"`
		Plan    domain.IterationPlan   `json:"plan"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &generateInput); err != nil {
		t.Fatalf("failed to parse next generate input: %v", err)
	}
	if generateInput.Context.BaseVersion == nil {
		t.Fatalf("expected base version context to remain enabled")
	}
	if generateInput.Context.PreviousReview != nil {
		t.Fatalf("previous review context should be omitted, got input=%#v", generateInput)
	}
	if generateInput.Plan.Explanation != "" || strings.Contains(generateInput.Plan.Instruction, "review context") {
		t.Fatalf("review-derived plan context should be omitted, got plan=%#v", generateInput.Plan)
	}
}

func TestGenerateContextCanFilterReviewExtensions(t *testing.T) {
	ctx := context.Background()
	executor := &testkit.Executor{}
	service := engine.NewService(memory.NewStore(), executor, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}), engine.WithAutoContinue())

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:        "fake",
		Target:          domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode:   domain.IterationModeAuto,
		MaxIterations:   2,
		GenerateContext: domain.NewGenerateContextOptions(domain.WithReviewExtensions("difficulty_profile")),
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	first, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: first.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 1"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: detail.Versions[0].ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":false,"score":4.5,"feedback":"needs work","difficulty_profile":{"level":"too_easy"},"manual_controls":{"length_direction":"shorter"}}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}

	var generateInput struct {
		Context domain.GenerateContext `json:"context"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &generateInput); err != nil {
		t.Fatalf("failed to parse next generate input: %v", err)
	}
	if generateInput.Context.PreviousReview == nil {
		t.Fatal("expected previous review context")
	}
	extensions := generateInput.Context.PreviousReview.Extensions
	if string(extensions["difficulty_profile"]) != `{"level":"too_easy"}` || len(extensions) != 1 {
		t.Fatalf("expected filtered review extensions, got %#v", extensions)
	}
}

func TestManualRunWaitsAfterGenerate(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	version, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: version.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"manual draft"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if detail.Run.Status != domain.RunStatusWaitingManual {
		t.Fatalf("manual run should wait after generate, got %#v", detail.Run)
	}
	if detail.Versions[0].ReviewJobID != "" || detail.Versions[0].Status != domain.VersionStatusGenerated {
		t.Fatalf("manual run should not auto dispatch review, got %#v", detail.Versions[0])
	}
}

func TestVersionConfigsAreStoredAndPassedToAdapter(t *testing.T) {
	ctx := context.Background()
	executor := &testkit.Executor{}
	service := engine.NewService(memory.NewStore(), executor, ports.NewSceneRegistry(&testkit.Adapter{}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
		Config: domain.Config{
			"model": "run-default",
		},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	first, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if string(first.GenerateConfig) != `{"model":"run-default"}` {
		t.Fatalf("first version should snapshot run config, got %s", first.GenerateConfig)
	}
	var firstGenerateInput struct {
		GenerateConfig json.RawMessage `json:"generate_config"`
	}
	if err := json.Unmarshal(executor.Jobs[0].Input, &firstGenerateInput); err != nil {
		t.Fatalf("parse first generate input: %v", err)
	}
	if string(firstGenerateInput.GenerateConfig) != `{"model":"run-default"}` {
		t.Fatalf("adapter did not receive default generate config: %s", firstGenerateInput.GenerateConfig)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: first.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 1"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}

	second, err := service.ContinueRun(ctx, engine.ContinueRunRequest{
		RunID:         run.ID,
		BaseVersionID: first.ID,
		GenerateConfig: domain.Config{
			"model": "generate-override",
		},
		Plan: domain.IterationPlan{Source: domain.PlanSourceManual, Instruction: "try another model"},
	})
	if err != nil {
		t.Fatalf("ContinueRun returned error: %v", err)
	}
	if string(second.GenerateConfig) != `{"model":"generate-override"}` {
		t.Fatalf("second version should store override generate config, got %s", second.GenerateConfig)
	}
	var secondGenerateInput struct {
		GenerateConfig json.RawMessage `json:"generate_config"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &secondGenerateInput); err != nil {
		t.Fatalf("parse second generate input: %v", err)
	}
	if string(secondGenerateInput.GenerateConfig) != `{"model":"generate-override"}` {
		t.Fatalf("adapter did not receive override generate config: %s", secondGenerateInput.GenerateConfig)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: second.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft 2"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult second returned error: %v", err)
	}

	reviewed, err := service.ReviewVersion(ctx, engine.ReviewVersionRequest{
		RunID:     run.ID,
		VersionID: second.ID,
		ReviewConfig: domain.Config{
			"model": "review-override",
		},
		OnFail: domain.ReviewPolicyWaitManual,
		Plan:   domain.IterationPlan{Source: domain.PlanSourceReviewOnly, Instruction: "review with another model"},
	})
	if err != nil {
		t.Fatalf("ReviewVersion returned error: %v", err)
	}
	if string(reviewed.GenerateConfig) != `{"model":"generate-override"}` {
		t.Fatalf("review-only version should retain base generate config, got %s", reviewed.GenerateConfig)
	}
	if string(reviewed.ReviewConfig) != `{"model":"review-override"}` {
		t.Fatalf("review-only version should store review config, got %s", reviewed.ReviewConfig)
	}
	var reviewInput struct {
		ReviewConfig json.RawMessage `json:"review_config"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &reviewInput); err != nil {
		t.Fatalf("parse review input: %v", err)
	}
	if string(reviewInput.ReviewConfig) != `{"model":"review-override"}` {
		t.Fatalf("adapter did not receive review config: %s", reviewInput.ReviewConfig)
	}
}

func TestVersionConfigRejectsNonJSONSerializableValues(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{}))

	_, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
		Config: map[string]any{
			"bad": func() {},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "config is not JSON serializable") {
		t.Fatalf("expected non-serializable config error, got %v", err)
	}
}

func TestSubmitCandidateForReviewCreatesVersionWithoutGenerate(t *testing.T) {
	ctx := context.Background()
	executor := &testkit.Executor{}
	service := engine.NewService(memory.NewStore(), executor, ports.NewSceneRegistry(&testkit.Adapter{}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	version, err := service.SubmitCandidateForReview(ctx, engine.SubmitCandidateForReviewRequest{
		RunID:   run.ID,
		Content: testkit.RawJSON(`{"text":"seed candidate"}`),
		Actor:   "admin",
	})
	if err != nil {
		t.Fatalf("SubmitCandidateForReview returned error: %v", err)
	}
	if version.VersionNo != 1 || version.GenerateJobID != "" || version.ReviewJobID == "" {
		t.Fatalf("expected review-only first version, got %#v", version)
	}
	if string(version.GeneratedContent) != `{"text":"seed candidate"}` || version.IterationPlan.Source != domain.PlanSourceSubmittedCandidate {
		t.Fatalf("unexpected submitted candidate version: %#v", version)
	}
	if len(executor.Jobs) != 1 || executor.Jobs[0].TaskName != "fake_review" {
		t.Fatalf("expected only review job to be submitted, got %#v", executor.Jobs)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if detail.Run.Status != domain.RunStatusReviewing || detail.Run.VersionCount != 1 {
		t.Fatalf("unexpected run after candidate submit: %#v", detail.Run)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: version.ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true,"score":8.5,"feedback":"seed ready"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusSucceeded || len(detail.Versions) != 1 {
		t.Fatalf("expected submitted candidate to pass, got run=%#v versions=%d", detail.Run, len(detail.Versions))
	}
}

func TestManualEditOverwritesGeneratedContentInPlace(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	executor := &testkit.Executor{}
	adapter := &testkit.Adapter{CanAuto: true}
	service := engine.NewService(store, executor, ports.NewSceneRegistry(adapter), engine.WithAutoContinue())

	run := createStartedRun(t, ctx, service)
	version := completeGenerateAndReview(t, ctx, service, run.ID, true)

	edited, err := service.EditVersion(ctx, engine.EditVersionRequest{
		RunID:     run.ID,
		VersionID: version.ID,
		Content:   testkit.RawJSON(`{"text":"edited"}`),
		Actor:     "admin",
	})
	if err != nil {
		t.Fatalf("EditVersion returned error: %v", err)
	}
	if edited.ID != version.ID || edited.VersionNo != version.VersionNo || edited.BaseVersionID != version.BaseVersionID {
		t.Fatalf("manual edit should update the same version, got %#v", edited)
	}
	detail, _ := service.GetRunDetail(ctx, run.ID)
	if len(detail.Versions) != 1 {
		t.Fatalf("manual edit should not create a new version, got %d", len(detail.Versions))
	}
	edited = detail.Versions[0]
	if string(edited.GeneratedContent) != `{"text":"edited"}` {
		t.Fatalf("generated content was not overwritten: %s", edited.GeneratedContent)
	}
	if string(edited.EffectiveContent()) != `{"text":"edited"}` {
		t.Fatalf("effective content did not use edit: %s", edited.EffectiveContent())
	}
	if edited.ReviewPass != nil || edited.ReviewJobID != "" || edited.Status != domain.VersionStatusGenerated {
		t.Fatalf("edit should clear stale review and mark generated: %#v", edited)
	}
	if detail.Run.Status != domain.RunStatusWaitingManual {
		t.Fatalf("edit should move succeeded run back to waiting manual, got %#v", detail.Run.Status)
	}
	if len(detail.VersionTree) != 1 || len(detail.VersionTree[0].Children) != 0 {
		t.Fatalf("expected edit to keep a single root version: %#v", detail.VersionTree)
	}

	reviewed, err := service.ReviewVersion(ctx, engine.ReviewVersionRequest{
		RunID:     run.ID,
		VersionID: edited.ID,
	})
	if err != nil {
		t.Fatalf("ReviewVersion returned error: %v", err)
	}
	if reviewed.ID == edited.ID || reviewed.BaseVersionID != edited.ID || reviewed.VersionNo != 2 || reviewed.ReviewJobID == "" {
		t.Fatalf("review should create a child review version, got %#v", reviewed)
	}
	if string(reviewed.GeneratedContent) != `{"text":"edited"}` {
		t.Fatalf("review child did not copy edited content: %s", reviewed.GeneratedContent)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: reviewed.ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true,"score":8.8,"feedback":"edited ready"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	result, err := service.AdoptVersion(ctx, engine.AdoptVersionRequest{
		RunID:     run.ID,
		VersionID: reviewed.ID,
		Actor:     "admin",
	})
	if err != nil {
		t.Fatalf("AdoptVersion returned error: %v", err)
	}
	if result.VersionID != reviewed.ID || len(adapter.Adopted) != 1 || string(adapter.Adopted[0]) != `{"text":"edited"}` {
		t.Fatalf("adopt did not use edited effective content: result=%#v adopted=%s", result, adapter.Adopted)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusAdopted || detail.Run.AdoptedVersionID != reviewed.ID {
		t.Fatalf("run was not adopted correctly: %#v", detail.Run)
	}
}

func TestReviewVersionDefaultsToWaitManualOnFail(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	executor := &testkit.Executor{}
	adapter := &testkit.Adapter{CanAuto: true}
	service := engine.NewService(store, executor, ports.NewSceneRegistry(adapter), engine.WithAutoContinue())

	run := createStartedRun(t, ctx, service)
	version := completeGenerateAndReview(t, ctx, service, run.ID, true)
	edited, err := service.EditVersion(ctx, engine.EditVersionRequest{
		RunID:     run.ID,
		VersionID: version.ID,
		Content:   testkit.RawJSON(`{"text":"edited but weak"}`),
	})
	if err != nil {
		t.Fatalf("EditVersion returned error: %v", err)
	}
	reviewed, err := service.ReviewVersion(ctx, engine.ReviewVersionRequest{RunID: run.ID, VersionID: edited.ID})
	if err != nil {
		t.Fatalf("ReviewVersion returned error: %v", err)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: reviewed.ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":false,"score":5,"feedback":"still weak"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	detail, _ := service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusWaitingManual {
		t.Fatalf("manual review failure should wait for manual action, got %#v", detail.Run)
	}
	if len(detail.Versions) != 2 {
		t.Fatalf("manual review should create exactly one review-only version, got %d", len(detail.Versions))
	}
	if detail.Versions[1].BaseVersionID != edited.ID || detail.Versions[1].IterationPlan.Source != domain.PlanSourceReviewOnly {
		t.Fatalf("unexpected review-only version: %#v", detail.Versions[1])
	}
}

func TestAutoContinueStopsAtMaxIterations(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}), engine.WithAutoContinue())
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeAuto,
		MaxIterations: 1,
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	version, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: version.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, _ := service.GetRunDetail(ctx, run.ID)
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: detail.Versions[0].ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":false,"score":5,"feedback":"still weak"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusMaxIterations || len(detail.Versions) != 1 {
		t.Fatalf("expected max iterations without auto child, got run=%#v versions=%d", detail.Run, len(detail.Versions))
	}
}

func TestContinueRunCanBranchFromEarlierVersion(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}), engine.WithAutoContinue())
	run := createStartedRun(t, ctx, service)
	root := completeGenerateAndReview(t, ctx, service, run.ID, true)

	firstBranch, err := service.ContinueRun(ctx, engine.ContinueRunRequest{
		RunID:         run.ID,
		BaseVersionID: root.ID,
		Plan: domain.IterationPlan{
			Source:      domain.PlanSourceManual,
			Instruction: "Create branch A.",
		},
	})
	if err != nil {
		t.Fatalf("ContinueRun branch A returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: firstBranch.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"branch a"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult branch A returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	firstBranch = detail.Versions[len(detail.Versions)-1]
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: firstBranch.ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true,"score":8,"feedback":"branch a ready"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult branch A returned error: %v", err)
	}

	secondBranch, err := service.ContinueRun(ctx, engine.ContinueRunRequest{
		RunID:         run.ID,
		BaseVersionID: root.ID,
		Plan: domain.IterationPlan{
			Source:      domain.PlanSourceManual,
			Instruction: "Create branch B.",
		},
	})
	if err != nil {
		t.Fatalf("ContinueRun branch B returned error: %v", err)
	}
	detail, err = service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if secondBranch.BaseVersionID != root.ID || secondBranch.Depth != root.Depth+1 || secondBranch.VersionNo != 3 {
		t.Fatalf("unexpected second branch version: %#v", secondBranch)
	}
	if len(detail.VersionTree) != 1 || len(detail.VersionTree[0].Children) != 2 {
		t.Fatalf("expected a root with two children, got %#v", detail.VersionTree)
	}
}

func TestContinueRunCanOmitGenerateContext(t *testing.T) {
	ctx := context.Background()
	executor := &testkit.Executor{}
	service := engine.NewService(memory.NewStore(), executor, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}), engine.WithAutoContinue())
	run := createStartedRun(t, ctx, service)
	root := completeGenerateAndReview(t, ctx, service, run.ID, true)

	branch, err := service.ContinueRun(ctx, engine.ContinueRunRequest{
		RunID:           run.ID,
		BaseVersionID:   root.ID,
		GenerateContext: domain.GenerateContextNone(),
		Plan: domain.IterationPlan{
			Source:      domain.PlanSourceManual,
			Instruction: "Create a fresh branch without previous context.",
		},
	})
	if err != nil {
		t.Fatalf("ContinueRun returned error: %v", err)
	}

	var generateInput struct {
		Context domain.GenerateContext `json:"context"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &generateInput); err != nil {
		t.Fatalf("failed to parse continue generate input: %v", err)
	}
	if generateInput.Context.BaseVersion != nil || generateInput.Context.PreviousReview != nil {
		t.Fatalf("generate context should be omitted, got input=%#v", generateInput)
	}
	if branch.BaseVersionID != root.ID || branch.IterationPlan.BaseVersionID != root.ID {
		t.Fatalf("lineage should remain persisted even when context is omitted: %#v", branch)
	}
}

func TestDirectiveValidationRejectsUnstableShape(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{}))
	_, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeManual,
		DefaultDirectives: []domain.IterationDirective{
			{Key: "difficulty", Value: testkit.RawJSON(`{"direction":"harder"}`), Description: "Make it harder."},
			{Key: "difficulty", Value: testkit.RawJSON(`{"direction":"easier"}`), Description: "Make it easier."},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate directive error")
	}
	_, err = service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeManual,
		DefaultDirectives: []domain.IterationDirective{
			{Key: "length", Value: testkit.RawJSON(`["longer"]`), Description: "Make it longer."},
		},
	})
	if err == nil {
		t.Fatal("expected non-object directive value error")
	}
}

func TestDirectiveHelpersAreAcceptedByEngine(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeManual,
		DefaultDirectives: []domain.IterationDirective{
			domain.IntDirective("difficulty", 1, "1 means harder, -1 means easier."),
			domain.BoolDirective("preserve_topic", true, "Keep the same topic."),
			domain.StringDirective("tone", "formal", "Use a formal tone."),
		},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	if len(run.DefaultDirectives) != 3 || string(run.DefaultDirectives[0].Value) != `{"value":1}` {
		t.Fatalf("unexpected directives: %#v", run.DefaultDirectives)
	}
}

func TestCreateRunDefaultsToManualAndRequiresAutoOption(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun manual default returned error: %v", err)
	}
	if run.IterationMode != domain.IterationModeManual || run.MaxIterations != 50 {
		t.Fatalf("expected safe manual defaults with max iterations 50, got %#v", run)
	}

	_, err = service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-2"},
		IterationMode: domain.IterationModeAuto,
	})
	if err == nil {
		t.Fatal("expected auto continue to require WithAutoContinue")
	}
}

func TestAdapterRejectsSceneKeyMismatch(t *testing.T) {
	ctx := context.Background()
	adapter := &sceneKeyChangingAdapter{}
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(adapter))

	_, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err == nil {
		t.Fatal("expected scene key mismatch error")
	}
	if !strings.Contains(err.Error(), "sceneKey mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdapterNilTargetReturnsClearError(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&nilTargetAdapter{}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	_, err = service.StartRun(ctx, run.ID)
	if err == nil || !strings.Contains(err.Error(), "adapter returned nil target") {
		t.Fatalf("expected nil target error, got %v", err)
	}
}

func TestAdapterNilGenerateJobFailsClearly(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&nilGenerateJobAdapter{}))

	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	_, err = service.StartRun(ctx, run.ID)
	if err == nil || !strings.Contains(err.Error(), "adapter returned nil generate job request") {
		t.Fatalf("expected nil generate job error, got %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if detail.Run.Status != domain.RunStatusFailed || len(detail.Versions) != 1 || detail.Versions[0].Status != domain.VersionStatusFailed {
		t.Fatalf("expected failed run/version, got run=%#v versions=%#v", detail.Run, detail.Versions)
	}
}

func TestAdapterNilGenerateResultFailsClearly(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&nilGenerateResultAdapter{}))
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	if _, err := service.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}

	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	err = service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: detail.Versions[0].GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft"}`),
	})
	if err == nil || !strings.Contains(err.Error(), "adapter returned nil generate result") {
		t.Fatalf("expected nil generate result error, got %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusFailed || detail.Versions[0].Status != domain.VersionStatusFailed {
		t.Fatalf("expected failed run/version, got run=%#v version=%#v", detail.Run, detail.Versions[0])
	}
}

func TestAdapterNilReviewJobFailsClearly(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&nilReviewJobAdapter{}))
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey: "fake",
		Target:   domain.TargetRef{Type: "document", ID: "doc-1"},
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	version, err := service.StartRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: version.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	version = detail.Versions[0]

	_, err = service.ReviewVersion(ctx, engine.ReviewVersionRequest{
		RunID:     run.ID,
		VersionID: version.ID,
	})
	if err == nil || !strings.Contains(err.Error(), "adapter returned nil review job request") {
		t.Fatalf("expected nil review job error, got %v", err)
	}
}

func TestAdapterNilReviewResultFailsClearly(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(
		memory.NewStore(),
		&testkit.Executor{},
		ports.NewSceneRegistry(&nilReviewResultAdapter{Adapter: testkit.Adapter{CanAuto: true}}),
		engine.WithAutoContinue(),
	)
	run := createStartedRun(t, ctx, service)

	detail, err := service.GetRunDetail(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: detail.Versions[0].GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"draft"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	err = service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: detail.Versions[0].ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true}`),
	})
	if err == nil || !strings.Contains(err.Error(), "adapter returned nil review result") {
		t.Fatalf("expected nil review result error, got %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusFailed || detail.Versions[0].Status != domain.VersionStatusFailed {
		t.Fatalf("expected failed run/version, got run=%#v version=%#v", detail.Run, detail.Versions[0])
	}
}

func createStartedRun(t *testing.T, ctx context.Context, service *engine.Service) *domain.Run {
	t.Helper()
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeAuto,
		MaxIterations: 3,
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	if _, err := service.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	return run
}

type sceneKeyChangingAdapter struct {
	testkit.Adapter
	calls int
}

func (a *sceneKeyChangingAdapter) Spec() domain.SceneSpec {
	a.calls++
	spec := a.Adapter.Spec()
	if a.calls > 1 {
		spec.SceneKey = "other"
	}
	return spec
}

type nilTargetAdapter struct {
	testkit.Adapter
}

func (a *nilTargetAdapter) LoadTarget(ctx context.Context, target domain.TargetRef) (*domain.TargetSnapshot, error) {
	return nil, nil
}

type nilGenerateJobAdapter struct {
	testkit.Adapter
}

func (a *nilGenerateJobAdapter) BuildGenerateJob(ctx context.Context, req ports.GenerateRequest) (*ports.JobRequest, error) {
	return nil, nil
}

type nilGenerateResultAdapter struct {
	testkit.Adapter
}

func (a *nilGenerateResultAdapter) ParseGenerateResult(ctx context.Context, raw []byte) (*domain.VersionContent, error) {
	return nil, nil
}

type nilReviewJobAdapter struct {
	testkit.Adapter
}

func (a *nilReviewJobAdapter) BuildReviewJob(ctx context.Context, req ports.ReviewRequest) (*ports.JobRequest, error) {
	return nil, nil
}

type nilReviewResultAdapter struct {
	testkit.Adapter
}

func (a *nilReviewResultAdapter) ParseReviewResult(ctx context.Context, raw []byte) (*domain.ReviewResult, error) {
	return nil, nil
}

func completeGenerateAndReview(t *testing.T, ctx context.Context, service *engine.Service, runID string, pass bool) *domain.Version {
	t.Helper()
	detail, err := service.GetRunDetail(ctx, runID)
	if err != nil {
		t.Fatalf("GetRunDetail returned error: %v", err)
	}
	version := detail.Versions[len(detail.Versions)-1]
	if err := service.ReceiveGenerateResult(ctx, engine.GenerateResultRequest{
		JobID: version.GenerateJobID,
		Raw:   testkit.RawJSON(`{"text":"generated"}`),
	}); err != nil {
		t.Fatalf("ReceiveGenerateResult returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, runID)
	version = detail.Versions[len(detail.Versions)-1]
	review := `{"pass":false,"score":4,"feedback":"not ready"}`
	if pass {
		review = `{"pass":true,"score":9,"feedback":"ready"}`
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: version.ReviewJobID,
		Raw:   testkit.RawJSON(review),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	detail, _ = service.GetRunDetail(ctx, runID)
	return detail.Versions[len(detail.Versions)-1]
}
