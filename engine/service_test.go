package engine_test

import (
	"context"
	"encoding/json"
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
		MaxDepth:      2,
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
		PreviousReview *domain.ReviewResult `json:"previous_review"`
	}
	if err := json.Unmarshal(executor.Jobs[len(executor.Jobs)-1].Input, &generateInput); err != nil {
		t.Fatalf("failed to parse next generate input: %v", err)
	}
	if generateInput.PreviousReview == nil || string(generateInput.PreviousReview.Extensions["difficulty_profile"]) != `{"level":"too_easy"}` {
		t.Fatalf("review extension was not passed to next generate: %#v", generateInput.PreviousReview)
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

func TestManualEditCreatesChildVersionWithoutOverwritingGenerated(t *testing.T) {
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
	if edited.BaseVersionID != version.ID || edited.VersionNo != 2 || edited.Depth != version.Depth+1 {
		t.Fatalf("manual edit should create a child version, got %#v", edited)
	}
	detail, _ := service.GetRunDetail(ctx, run.ID)
	original := detail.Versions[0]
	if string(original.GeneratedContent) != `{"text":"generated"}` {
		t.Fatalf("generated content was overwritten: %s", original.GeneratedContent)
	}
	if len(edited.GeneratedContent) != 0 {
		t.Fatalf("edited child should not pretend to be generated: %s", edited.GeneratedContent)
	}
	if string(edited.EffectiveContent()) != `{"text":"edited"}` {
		t.Fatalf("effective content did not use edit: %s", edited.EffectiveContent())
	}
	if edited.ReviewPass != nil || edited.Status != domain.VersionStatusEdited {
		t.Fatalf("edit should clear stale review and mark edited: %#v", edited)
	}
	if detail.Run.Status != domain.RunStatusWaitingManual {
		t.Fatalf("edit should move succeeded run back to waiting manual, got %#v", detail.Run.Status)
	}
	if len(detail.VersionTree) != 1 || len(detail.VersionTree[0].Children) != 1 {
		t.Fatalf("expected edited version to appear as child in version tree: %#v", detail.VersionTree)
	}

	reviewed, err := service.ReviewVersion(ctx, engine.ReviewVersionRequest{
		RunID:     run.ID,
		VersionID: edited.ID,
	})
	if err != nil {
		t.Fatalf("ReviewVersion returned error: %v", err)
	}
	if reviewed.ReviewJobID == "" {
		t.Fatalf("review job was not submitted: %#v", reviewed)
	}
	if err := service.ReceiveReviewResult(ctx, engine.ReviewResultRequest{
		JobID: reviewed.ReviewJobID,
		Raw:   testkit.RawJSON(`{"pass":true,"score":8.8,"feedback":"edited ready"}`),
	}); err != nil {
		t.Fatalf("ReceiveReviewResult returned error: %v", err)
	}
	result, err := service.AdoptVersion(ctx, engine.AdoptVersionRequest{
		RunID:     run.ID,
		VersionID: edited.ID,
		Actor:     "admin",
	})
	if err != nil {
		t.Fatalf("AdoptVersion returned error: %v", err)
	}
	if result.VersionID != edited.ID || len(adapter.Adopted) != 1 || string(adapter.Adopted[0]) != `{"text":"edited"}` {
		t.Fatalf("adopt did not use edited effective content: result=%#v adopted=%s", result, adapter.Adopted)
	}
	detail, _ = service.GetRunDetail(ctx, run.ID)
	if detail.Run.Status != domain.RunStatusAdopted || detail.Run.AdoptedVersionID != edited.ID {
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
		t.Fatalf("manual review failure should not auto-create a version, got %d", len(detail.Versions))
	}
}

func TestAutoContinueStopsAtMaxDepth(t *testing.T) {
	ctx := context.Background()
	service := engine.NewService(memory.NewStore(), &testkit.Executor{}, ports.NewSceneRegistry(&testkit.Adapter{CanAuto: true}), engine.WithAutoContinue())
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeAuto,
		MaxDepth:      1,
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
	if detail.Run.Status != domain.RunStatusMaxDepth || len(detail.Versions) != 1 {
		t.Fatalf("expected max depth without auto child, got run=%#v versions=%d", detail.Run, len(detail.Versions))
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
	if run.IterationMode != domain.IterationModeManual || run.MaxDepth != 50 {
		t.Fatalf("expected safe manual defaults with max depth 50, got %#v", run)
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

func createStartedRun(t *testing.T, ctx context.Context, service *engine.Service) *domain.Run {
	t.Helper()
	run, err := service.CreateRun(ctx, engine.CreateRunRequest{
		SceneKey:      "fake",
		Target:        domain.TargetRef{Type: "document", ID: "doc-1"},
		IterationMode: domain.IterationModeAuto,
		MaxDepth:      3,
	})
	if err != nil {
		t.Fatalf("CreateRun returned error: %v", err)
	}
	if _, err := service.StartRun(ctx, run.ID); err != nil {
		t.Fatalf("StartRun returned error: %v", err)
	}
	return run
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
