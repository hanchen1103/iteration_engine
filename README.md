# Iteration Engine

`iteration_engine` 是一个用于“生成 -> 审核 -> 继续迭代 -> 人工处理 -> 采纳”的 Go 编排组件。

它只负责通用生命周期和状态流转，不绑定具体业务内容、Prompt、LLM、队列、Worker 或数据库实现。业务系统通过 `ports.SceneAdapter` 接入场景逻辑，执行系统通过 `ports.JobExecutor` 接入异步任务。

## 适用场景

- 需要对同一个业务对象反复生成候选版本，例如文章、题目、听力材料、配置草稿。
- 每个候选版本都需要结构化审核结果，并根据审核结果决定通过、等待人工、继续自动迭代或达到上限。
- 需要保留版本树、审核记录、人工编辑、采纳记录和事件日志。
- 希望编排逻辑独立于业务表、LLM 平台和任务队列。

## 核心概念

- `Run`：一次迭代任务，绑定一个 `SceneKey` 和一个业务目标 `TargetRef`。
- `Version`：一次生成或提交的候选版本。版本通过 `BaseVersionID` 形成树结构。
- `SceneAdapter`：业务场景适配器，负责加载目标、构造生成/审核任务、解析任务输出、采纳版本。
- `JobExecutor`：任务执行入口，负责提交生成/审核任务并返回 `JobID`。
- `Store`：持久化接口，负责保存 Run、Version、Event。
- `IterationPlan`：一次生成或审核的意图，包括来源、说明、指令和结构化控制项。
- `ReviewResult`：审核结果，包括是否通过、分数、反馈、问题列表和业务扩展字段。

## 目录结构

```text
iteration_engine/
  domain/              稳定领域类型、决策逻辑、版本树构建
  ports/               业务适配器、任务执行器、存储接口
  engine/              编排服务、状态流转、回调入口
  store/memory/        内存 Store，适合测试和本地开发
  store/sql/mysql/     推荐的 MySQL 表结构
  docs/                生产存储映射说明
  testkit/             测试用适配器和执行器
```

## 安装

```bash
go get github.com/hanchen1103/iteration_engine
```

本仓库是库，不提供独立 CLI 或 HTTP 服务。调用方需要在自己的服务里初始化 `engine.Service`，并把 API、队列回调或 Worker 回调转发给它。

## 快速开始

最小接入需要准备三个对象：

1. `ports.Store`：开发期可以用 `memory.NewStore()`。
2. `ports.SceneAdapter`：业务侧实现，定义某个场景如何生成、审核和采纳。
3. `ports.JobExecutor`：提交异步生成/审核任务，返回可回调的 `JobID`。

```go
package example

import (
    "context"

    "github.com/hanchen1103/iteration_engine/domain"
    iterengine "github.com/hanchen1103/iteration_engine/engine"
    "github.com/hanchen1103/iteration_engine/ports"
    "github.com/hanchen1103/iteration_engine/store/memory"
)

func start(ctx context.Context, adapter ports.SceneAdapter, executor ports.JobExecutor) error {
    store := memory.NewStore()
    registry := ports.NewSceneRegistry(adapter)

    service := iterengine.NewService(store, executor, registry)

    run, err := service.CreateRun(ctx, iterengine.CreateRunRequest{
        SceneKey:      "article_rewrite",
        Target:        domain.TargetRef{Type: "article", ID: "article-123"},
        IterationMode: domain.IterationModeManual,
        MaxIterations: 3,
        DefaultDirectives: []domain.IterationDirective{
            domain.StringDirective("tone", "formal", "Use a formal tone."),
        },
        Actor: "user-1",
    })
    if err != nil {
        return err
    }

    version, err := service.StartRun(ctx, run.ID)
    if err != nil {
        return err
    }

    // version.GenerateJobID 需要交给你的队列/Worker。
    // Worker 完成后调用 service.ReceiveGenerateResult。
    _ = version.GenerateJobID
    return nil
}
```

## 实现 SceneAdapter

`SceneAdapter` 是业务接入的核心。一个场景通常对应一种目标类型和一组生成/审核规则。

```go
type ArticleAdapter struct{}

func (a *ArticleAdapter) Spec() domain.SceneSpec {
    return domain.SceneSpec{
        SceneKey:   "article_rewrite",
        TargetType: "article",
        GenerateRule: domain.RuleSpec{
            Role:        "writer",
            RuleKey:     "article_rewrite_generate",
            RuleVersion: "v1",
        },
        ReviewRule: domain.RuleSpec{
            Role:        "reviewer",
            RuleKey:     "article_rewrite_review",
            RuleVersion: "v1",
        },
        Capability: domain.SceneCapability{
            CanAutoContinue:      true,
            CanManualEdit:        true,
            CanReviewOnly:        true,
            CanAdopt:             true,
            DefaultMaxIterations: 5,
        },
    }
}
```

需要实现的方法：

- `LoadTarget`：根据 `domain.TargetRef` 读取业务对象快照。
- `BuildGenerateJob`：把目标、上下文和计划转换成生成任务 `ports.JobRequest`。
- `ParseGenerateResult`：把 Worker/LLM 原始输出解析成 `domain.VersionContent`。
- `BuildReviewJob`：把候选内容转换成审核任务 `ports.JobRequest`。
- `ParseReviewResult`：把原始审核输出解析成 `domain.ReviewResult`。
- `Adopt`：把某个版本写回业务系统。

`BuildGenerateJob` 和 `BuildReviewJob` 不需要填 `RunID`、`VersionID`、`SceneKey`、`RoleKey` 等通用字段，engine 会在提交前补齐。

## 实现 JobExecutor

`JobExecutor` 只负责提交任务并返回任务 ID。它可以封装队列、HTTP 调用、数据库任务表或任意 Worker 系统。

```go
type QueueExecutor struct {
    queue Queue
}

func (e *QueueExecutor) Submit(ctx context.Context, req *ports.JobRequest) (*ports.JobHandle, error) {
    jobID, err := e.queue.Enqueue(ctx, req)
    if err != nil {
        return nil, err
    }
    return &ports.JobHandle{JobID: jobID}, nil
}
```

Worker 完成后，调用方需要根据任务类型回调 engine：

```go
err := service.ReceiveGenerateResult(ctx, iterengine.GenerateResultRequest{
    JobID: "job-123",
    Raw:   rawGenerateOutput,
})
```

```go
err := service.ReceiveReviewResult(ctx, iterengine.ReviewResultRequest{
    JobID: "job-456",
    Raw:   rawReviewOutput,
})
```

如果任务失败，传 `ErrorMessage`：

```go
err := service.ReceiveGenerateResult(ctx, iterengine.GenerateResultRequest{
    JobID:        "job-123",
    ErrorMessage: "model timeout",
})
```

engine 会通过 `Store.FindVersionByJobID` 找到对应版本，并忽略已经过期的旧回调。

## 手动模式

手动模式是默认模式。`CreateRunRequest.IterationMode` 为空时等同于 `domain.IterationModeManual`。

流程：

1. `CreateRun` 创建 Run，状态为 `PENDING`。
2. `StartRun` 创建第一个 Version，并提交生成任务。
3. `ReceiveGenerateResult` 保存生成内容。
4. Run 进入 `WAITING_MANUAL`，不会自动发起审核。
5. 调用方可以继续人工操作：编辑、审核、继续生成、采纳。

常用操作：

```go
// 基于某个版本继续生成一个子版本。
next, err := service.ContinueRun(ctx, iterengine.ContinueRunRequest{
    RunID:         runID,
    BaseVersionID: versionID,
    Plan: domain.IterationPlan{
        Source:      domain.PlanSourceManual,
        Instruction: "Make the answer shorter and more direct.",
    },
    Actor: "editor-1",
})
```

```go
// 直接编辑某个版本的内容。编辑会覆盖该版本的生成内容，并清空旧审核结果。
edited, err := service.EditVersion(ctx, iterengine.EditVersionRequest{
    RunID:     runID,
    VersionID: versionID,
    Content:   []byte(`{"text":"edited content"}`),
    Actor:     "editor-1",
})
```

```go
// 对某个已有版本发起审核。该方法会创建一个 review-only 子版本，用于保留审核历史。
reviewVersion, err := service.ReviewVersion(ctx, iterengine.ReviewVersionRequest{
    RunID:     runID,
    VersionID: versionID,
    OnFail:    domain.ReviewPolicyWaitManual,
})
```

```go
// 不经过生成，直接提交外部候选内容进行审核。
candidate, err := service.SubmitCandidateForReview(ctx, iterengine.SubmitCandidateForReviewRequest{
    RunID:   runID,
    Content: []byte(`{"text":"external candidate"}`),
    Actor:   "editor-1",
})
```

```go
// 将某个版本采纳到业务系统。
result, err := service.AdoptVersion(ctx, iterengine.AdoptVersionRequest{
    RunID:     runID,
    VersionID: versionID,
    Actor:     "editor-1",
})
```

## 自动模式

自动模式需要同时满足两个条件：

- 创建 service 时显式启用 `engine.WithAutoContinue()`。
- 场景能力 `SceneCapability.CanAutoContinue` 为 `true`。

```go
service := iterengine.NewService(
    store,
    executor,
    ports.NewSceneRegistry(adapter),
    iterengine.WithAutoContinue(),
)

run, err := service.CreateRun(ctx, iterengine.CreateRunRequest{
    SceneKey:      "article_rewrite",
    Target:        domain.TargetRef{Type: "article", ID: "article-123"},
    IterationMode: domain.IterationModeAuto,
    MaxIterations: 5,
})
```

自动模式流程：

1. `StartRun` 提交生成任务。
2. `ReceiveGenerateResult` 保存内容后自动提交审核任务。
3. `ReceiveReviewResult` 解析审核结果。
4. 审核通过时 Run 进入 `SUCCEEDED`。
5. 审核失败且未达到 `MaxIterations` 时自动基于当前版本继续生成。
6. 达到上限时 Run 进入 `MAX_ITERATIONS`。

自动继续默认关闭，是为了避免接入方在未明确配置任务队列、成本控制和审核策略时产生无限任务链。

## 控制生成上下文

继续生成时，engine 可以把父版本和上一次审核结果传给 `BuildGenerateJob`：

```go
func (a *ArticleAdapter) BuildGenerateJob(ctx context.Context, req ports.GenerateRequest) (*ports.JobRequest, error) {
    // req.Context.BaseVersion 是父版本上下文。
    // req.Context.PreviousReview 是上一轮审核上下文。
    // req.Plan 是本轮迭代计划。
    return &ports.JobRequest{
        TaskName: "article_generate",
        Input:    buildPromptInput(req),
    }, nil
}
```

默认情况下，父版本和审核字段都会包含。可以在创建 Run 或手动 Continue 时调整：

```go
run, err := service.CreateRun(ctx, iterengine.CreateRunRequest{
    SceneKey: "article_rewrite",
    Target:   domain.TargetRef{Type: "article", ID: "article-123"},
    GenerateContext: domain.NewGenerateContextOptions(
        domain.WithBaseContent(),
        domain.WithoutBaseArtifacts(),
        domain.WithReviewFeedback(),
        domain.WithReviewExtensions("difficulty_profile", "manual_controls"),
        domain.WithoutReviewRawJSON(),
    ),
})
```

常用预设：

- `domain.GenerateContextFull()`：包含全部上下文。
- `domain.GenerateContextNone()`：不传父版本和审核上下文。
- `domain.GenerateContextBaseContentOnly()`：只传父版本内容。
- `domain.GenerateContextReviewFeedbackOnly()`：只传审核反馈。

## IterationPlan 和 Directive

`IterationPlan` 描述本轮迭代意图：

```go
plan := domain.IterationPlan{
    Source:      domain.PlanSourceManual,
    Explanation: "User wants a more concise version.",
    Instruction: "Shorten the content while preserving the main argument.",
    Directives: []domain.IterationDirective{
        domain.IntDirective("length_delta", -1, "-1 means shorter, 1 means longer."),
        domain.BoolDirective("preserve_facts", true, "Keep all factual claims."),
    },
}
```

内置 `PlanSource`：

- `PlanSourceInitial`
- `PlanSourceManual`
- `PlanSourceAutoReview`
- `PlanSourceManualEdit`
- `PlanSourceSubmittedCandidate`
- `PlanSourceReviewOnly`

Directive 的 `Value` 必须是 JSON object，同一个 Plan 内不能重复 `Key`。标量值建议使用辅助函数，它们会自动包装成 `{"value": ...}`：

- `domain.IntDirective`
- `domain.BoolDirective`
- `domain.StringDirective`
- `domain.ValueDirective`
- `domain.ObjectDirective`

## 查询 Run 和版本树

```go
detail, err := service.GetRunDetail(ctx, runID)
if err != nil {
    return err
}

run := detail.Run
versions := detail.Versions      // 按 VersionNo 升序
tree := detail.VersionTree       // 按 BaseVersionID 构建的树
events := detail.Events          // 生命周期事件
```

也可以按条件列出 Run：

```go
runs, err := service.ListRuns(ctx, ports.ListRunsFilter{
    SceneKey: "article_rewrite",
    Target:   domain.TargetRef{Type: "article", ID: "article-123"},
    Status:   domain.RunStatusWaitingManual,
    Limit:    20,
})
```

## 状态说明

Run 状态：

- `PENDING`：已创建，未开始。
- `GENERATING`：正在生成。
- `REVIEWING`：正在审核。
- `WAITING_MANUAL`：等待人工继续、编辑、审核或采纳。
- `SUCCEEDED`：审核通过。
- `MAX_ITERATIONS`：自动迭代达到上限。
- `FAILED`：生成、审核或适配器处理失败。
- `ADOPTED`：已采纳到业务系统。

Version 状态：

- `GENERATING`：生成任务已提交。
- `GENERATED`：已有候选内容。
- `REVIEWING`：审核任务已提交。
- `REVIEWED`：审核结果已保存。
- `EDITED`：保留给兼容状态；当前编辑路径会覆盖生成内容并回到 `GENERATED`。
- `FAILED`：该版本处理失败。
- `ADOPTED`：该版本已采纳。

## 错误处理

engine 返回的领域错误类型是 `*domain.Error`，包含稳定的 `Code`：

- `INVALID`
- `NOT_FOUND`
- `CONFLICT`
- `FAILED`
- `FORBIDDEN`

```go
var domainErr *domain.Error
if errors.As(err, &domainErr) {
    switch domainErr.Code {
    case domain.ErrorCodeConflict:
        // 例如 Run 正在生成/审核，不能继续编辑。
    case domain.ErrorCodeForbidden:
        // 例如未启用自动继续，或场景不支持某个能力。
    }
}
```

## 生产存储

`store/memory` 只适合测试和本地开发。生产环境应实现 `ports.Store` 并接入自己的数据库、事务和重试策略。

推荐 MySQL 基线表结构：

```text
store/sql/mysql/schema.sql
```

详细映射说明见：

```text
docs/mysql_store.md
```

生产实现要点：

- `Run.ID`、`Version.ID`、`Event.ID` 可以由 Store 在创建时生成。
- `ListVersions(runID)` 应按 `version_no ASC` 返回。
- `ListRuns(filter)` 建议按 `created_at DESC` 返回。
- `FindVersionByJobID` 必须能通过 `generate_job_id` 或 `review_job_id` 找回版本，用于服务重启后的任务回调恢复。
- `active_job_id`、`active_version_id`、`active_role_key` 需要持久化，engine 会用它们忽略旧回调。
- 业务系统自行负责迁移、事务边界、锁、幂等、备份和保留策略。

## 设计边界

engine 负责：

- Run 和 Version 生命周期。
- 版本树关系。
- 生成/审核任务提交。
- 审核结果标准化和决策。
- 手动继续、手动编辑、外部候选审核。
- 采纳生命周期和事件记录。

engine 不负责：

- 业务内容结构。
- Prompt 文本。
- 模型路由和 LLM 调用。
- 队列、Worker 和 HTTP 回调协议。
- 业务表写入。
- 生产数据库迁移。

这种边界使同一个 engine 可以服务多个业务场景，而每个场景只需要提供自己的 `SceneAdapter`。
