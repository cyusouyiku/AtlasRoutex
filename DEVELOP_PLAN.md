# 开发计划（2 周 / Sprint）

前提：`internal/domain`（entity / valueobject / repository）已写完。下列按 README 目录排期，未列文件表示该 Sprint 不动或后移。

---

## 进度快照（与当前仓库对齐）

| 范围 | 状态 | 说明 |
|------|------|------|
| `internal/domain/*` | ✅ 已具备 | entity / valueobject / repository 接口 |
| `internal/application/planner/*`（`plan_usecase`、`adjust_usecase`、`dto`） | ✅ 已具备 | 编排与 `TripSolver` 端口已定义 |
| `internal/infrastructure/persistence/postgres/*`（`db`、`poi`/`itinerary`/`user` repo） | ✅ 已具备 | 尚无迁移脚本时，表结构依赖人工对齐，联调易漂移 |
| `cmd/migration`、`scripts/*.sql` | 🔲 未做 | **建议下一批最高优先级**，与已有 PG 实现配套 |
| `internal/core/solver`（及 `TripSolver` 的真实现 / stub） | 🔲 未做 | 用例已依赖接口；无实现则 `Execute` 无法端到端跑通 |
| Recommender / Feedback、Redis、HTTP、gRPC 等 | 🔲 未做 | 按下方调整后的顺序推进 |

---

## 顺序调整说明（原计划的修正）

1. **迁移应与 PG 仓储同频，而不是「写完所有 repo 再补 migration」**  
   当前已有 postgres 实现但缺少 `cmd/migration`，新人无法从零建库。应优先补齐 `001_init_schema` 等与表结构一致的脚本，再迭代 `002` 索引、`003` seed。

2. **求解器应「接口 + 可运行 stub 早实现」，全量 GA/AC-3 仍可后置**  
   `PlanUsecase` 已注入 `TripSolver`；若等到 Sprint 3 才在 `internal/core/solver` 落文件，中间几周主链路无法集成测试。建议在 **`internal/core/solver` 先交付：正式 `Problem`/`Solution`（或与现有 `SolverInput`/`SolverOutput` 对齐）+ 一个 naive/stub 实现**（例如贪心排程），保证「PG + migration + stub」可打通；遗传算法、AC-3、蚁群等仍在后续 Sprint 加深。

3. **Redis 不应急于与「第一次 PG 联调」强绑定**  
   热点读、距离矩阵缓存可在主路径「规划 + 持久化」稳定后再加；否则调试同时要排缓存失效与 TTL，问题面过大。可先完成 PG + migration + stub solver 闭环，再将 Redis 作为性能与预计算缓存层接入。

4. **推荐与反馈用例不阻塞规划 MVP**  
   `recommender` / `feedback` 可与求解器深化并行，但不必排在「能存能算一条行程」之前。

5. **最小 HTTP 入口可早于「完整 Wire + 全套中间件」**  
   原 Sprint 6 一口气堆满 bootstrap/handler/middleware 周期较长。更合理的是：**在 stub solver + migration + PG 可用后**，用极简 `cmd/api`（哪怕暂时手动 `NewPlanUsecase`）打通一条 POST 规划，再逐步换 Wire、鉴权、限流等。文档里将 Sprint 6 拆成「6A 最小可联调 API」与「6B 工程化完善」两条线描述。

---

## Sprint 1（第 1–2 周）— 应用层用例，已完成所有交付

**需求背景：** 领域模型和仓储接口已经定义，但还需要「一条可执行的业务流程」把用户意图串起来。

**要解决什么问题：** 把「规划行程、调整行程」等拆成独立用例；用例只依赖 `domain` 类型和 `repository` 接口以及求解端口。

### 交付文件

| 状态 | 文件 | 职责与要点 |
|------|------|------------|
| ✅ | `internal/application/planner/plan_usecase.go` | 从输入到可保存行程的编排；错误语义清晰；不写 SQL/HTTP。 |
| ✅ | `internal/application/planner/adjust_usecase.go` | 调整日/换点/改时段；增量或全量策略占位；乐观锁占位。 |
| ✅ | `internal/application/planner/dto.go` | 规划与调整 IO；与 entity 解耦；校验与 validator 衔接点。 |
| 🔲 | `internal/application/recommender/recommend_usecase.go` | 给定上下文返回 POI 列表；可 **与 Sprint 3–4 并行**，不挡 MVP。 |
| 🔲 | `internal/application/recommender/strategy.go` | 规则过滤与打分；无状态或可注入策略。 |
| 🔲 | `internal/application/feedback/feedback_usecase.go` | 反馈入口；异步占位。 |
| 🔲 | `internal/application/feedback/collector.go` | 校验、去重、聚合后交 usecase。 |

**本 Sprint 剩余工作：** 推荐与反馈；规划主路径已在代码中，需与下方「迁移 + stub 求解器」联调验证。

---

## Sprint 2（第 3–4 周）— 迁移 + PostgreSQL（Redis 后置），预计4.13开始开发

**需求背景：** 用例需要真实数据与可复现 schema；**迁移优先于或紧同步于仓储迭代**。

**要解决什么问题：** 表结构与领域一致；从零建库可复现；**Redis 可在主路径打通后再接**（见顺序说明第 3 点）。

### 交付文件

| 状态 | 文件 | 职责与要点 |
|------|------|------------|
| 🔲 | `cmd/migration/main.go` | CLI：DSN、按版本执行 SQL、schema_migrations；失败非零退出。 |
| 🔲 | `cmd/migration/scripts/001_init_schema.sql` | 与 entity / 现有 repo 查询一致；主键外键与时间默认值与 JSON tag 对齐。 |
| 🔲 | `cmd/migration/scripts/002_add_indexes.sql` | 城市/类别/用户行程等高频查询索引。 |
| 🔲 | `cmd/migration/scripts/003_seed_data.sql` | 最小 POI + 示例行程，供集成测试稳定输入。 |
| ✅ | `internal/infrastructure/persistence/postgres/db.go` | 连接池与健康检查。 |
| ✅ | `internal/infrastructure/persistence/postgres/poi_repo_impl.go` | `PoiRepository` 实现。 |
| ✅ | `internal/infrastructure/persistence/postgres/itinerary_repo_impl.go` | `ItineraryRepository` 实现。 |
| ✅ | `internal/infrastructure/persistence/postgres/user_repo_impl.go` | `UserRepository` 实现（README 未单列但实际需要）。 |
| 🔲 → **Sprint 2 末或 Sprint 3 初** | `internal/infrastructure/persistence/redis/client.go` | 与 PG 联调稳定后再接。 |
| 🔲 | `internal/infrastructure/persistence/redis/cache_repo_impl.go` | 规划结果 / POI 列表 / 矩阵片段；TTL 与击穿策略。 |

---

## Sprint 3（第 5–6 周）— 求解器接口 + stub + 约束与遗传算法（分段交付）

**需求背景：** 规划是组合优化；需要可替换 solver。**必须先有可运行的默认实现（stub），再叠 GA/约束深化**。

**要解决什么问题：** 统一问题/解抽象；**尽早**让 `PlanUsecase` 能调用真实包路径下的实现；再实现约束传播、AC-3、GA 闭环。

### 交付文件（建议实现顺序）

| 顺序 | 文件 | 职责与要点 |
|------|------|------------|
| ① | `internal/core/solver/solver.go` | `Solver` 接口与 `Problem`/`Solution`；`context` 取消与超时；**与 `planner.SolverInput`/`SolverOutput` 映射层可放在 application 边缘或 solver 包内适配函数，避免两套模型长期分裂**。 |
| ② | `internal/core/solver/stub.go`（或 `naive/`） | **简单可行解**：固定规则或贪心，供集成测试与演示；后续由 GA/ACO 替换或并存。 |
| ③ | `internal/core/solver/constraint/propagator.go` | 约束图与域缩减。 |
| ④ | `internal/core/solver/constraint/ac3.go` | AC-3 弧相容。 |
| ⑤ | `internal/core/solver/constraint/validator.go` | 硬约束检查与软约束惩罚。 |
| ⑥ | `internal/core/solver/genetic/chromosome.go` | 编码/解码与合法化。 |
| ⑦ | `internal/core/solver/genetic/crossover.go` | OX/PMX 等。 |
| ⑧ | `internal/core/solver/genetic/mutation.go` | 变异与修复。 |
| ⑨ | `internal/core/solver/genetic/selection.go` | 锦标赛 / 轮盘 / 精英。 |
| ⑩ | `internal/core/solver/genetic/fitness.go` | 路程、时间窗、预算、偏好综合。 |

---

## Sprint 4（第 7–8 周）— 蚁群 + 混合求解 + 规划引擎

（内容与原计划一致；**建议在 `internal/core/planner/engine.go` 接线前，先用 stub + 单算法验证 `PlanUsecase`**，引擎作为对多算法与流水线的聚合。）

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/core/solver/antcolony/colony.go` | 蚂蚁主循环；对接 `validator`。 |
| `internal/core/solver/antcolony/pheromone.go` | 信息素矩阵与参数。 |
| `internal/core/solver/antcolony/heuristic.go` | 启发式可见度。 |
| `internal/core/solver/hybrid/orchestrator.go` | 多算法与超时降级。 |
| `internal/core/solver/hybrid/ensemble.go` | 多解合并与 top-k。 |
| `internal/core/planner/engine.go` | 对上层暴露统一 `Plan`；诊断信息。 |
| `internal/core/planner/hierarchical.go` | 分区再子求解。 |
| `internal/core/planner/incremental.go` | 增量重算与回退全量。 |
| `internal/core/planner/pipeline.go` | 召回 → 剪枝 → 求解 → 后处理。 |

---

## Sprint 5（第 9–10 周）— 多目标优化 + 图谱 + Neo4j + 地图 API

（与原计划一致；Neo4j / 地图 API 可按 scope 裁剪，**地图距离建议在真实 GA 调参前至少提供「直线 + 可配置系数」或缓存矩阵占位**，避免 fitness 完全失真。）

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/core/optimization/objective.go` | 多目标与加权合成。 |
| `internal/core/optimization/weight.go` | 配置与画像权重。 |
| `internal/core/optimization/pareto.go` | 非支配解集。 |
| `internal/core/graph/graph.go` | 图抽象。 |
| `internal/core/graph/builder.go` | 批量构建/导入。 |
| `internal/core/graph/query.go` | 高层查询。 |
| `internal/core/graph/index/geohash.go` | 网格粗筛。 |
| `internal/core/graph/index/rtree.go` | 近邻查询。 |
| `internal/infrastructure/persistence/neo4j/driver.go` | Driver 生命周期。 |
| `internal/infrastructure/persistence/neo4j/graph_repo_impl.go` | Cypher 实现。 |
| `internal/infrastructure/external/mapapi/client.go` | REST 封装与重试。 |
| `internal/infrastructure/external/mapapi/distance.go` | 矩阵与降级策略。 |

---

## Sprint 6（第 11–12 周）— Bootstrap + HTTP + cmd/api + Wire

**调整：** 优先 **6A** 可运行的一条规划 API，再 **6B** 补全中间件与 Wire。

### 6A（可与 Sprint 2–3 交叉，时间允许时提前）

- `cmd/api/main.go`：读配置、连 PG、构造 `PlanUsecase`（注入 stub 或真实 solver）、注册 **1 条** 规划路由。
- `internal/api/handler/plan_handler.go` 骨架 + `pkg/errors`、`pkg/logger` 最小可用。

### 6B（工程化）

| 文件 | 职责与要点 |
|------|------------|
| `internal/bootstrap/config.go` | 配置与校验。 |
| `internal/bootstrap/logger.go` | zap 等。 |
| `internal/bootstrap/database.go` | PG/Redis/Neo4j provider。 |
| `internal/bootstrap/server.go` | `http.Server` 与优雅退出。 |
| `internal/bootstrap/app.go` | Router + 中间件 + `Run()`。 |
| `internal/api/handler/poi_handler.go` | POI 列表/详情。 |
| `internal/api/handler/user_handler.go` | 用户相关。 |
| `internal/api/handler/middleware/auth.go` | JWT/API Key。 |
| `internal/api/handler/middleware/ratelimit.go` | 限流。 |
| `internal/api/handler/middleware/logger.go` | 访问日志。 |
| `internal/api/handler/middleware/recovery.go` | Panic 恢复。 |
| `internal/api/dto/request.go` | HTTP 入参。 |
| `internal/api/dto/response.go` | 统一响应体。 |
| `internal/api/dto/converter.go` | 与 application dto 转换。 |
| `cmd/api/wire.go` / `wire_gen.go` | 编译期注入。 |
| `pkg/errors/errors.go`、`codes.go` | 错误包装与码表。 |
| `pkg/logger/logger.go`、`zap.go` | 日志接口与实现。 |

---

## Sprint 7（第 13–14 周）— MQ + 预订/支付 + Worker + 指标与集成测

（与原计划一致。）

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/infrastructure/mq/producer.go` | 发送任务；trace id。 |
| `internal/infrastructure/mq/consumer.go` | 订阅与分发。 |
| `internal/infrastructure/mq/topics.go` | Topic 常量。 |
| `internal/infrastructure/external/booking/client.go` | 预订网关基类。 |
| `internal/infrastructure/external/booking/hotel.go` | 酒店下单等。 |
| `internal/infrastructure/external/payment/client.go` | 支付与回调占位。 |
| `cmd/worker/main.go` | Worker 入口。 |
| `cmd/worker/handlers/precompute_handler.go` | 预计算与幂等。 |
| `cmd/worker/handlers/feedback_handler.go` | 异步反馈。 |
| `pkg/queue/queue.go`、`kafka.go` | 队列抽象与实现。 |
| `pkg/metrics/prometheus.go`、`recorder.go` | 指标。 |
| `pkg/utils/hash.go`、`time.go`、`math.go`、`validator.go` | 通用工具。 |
| `pkg/cache/cache.go`、`memory.go`、`redis.go` | 业务级缓存（可与 persistence/redis 分工合并）。 |
| `tests/integration/api_test.go` | HTTP 主链路。 |
| `tests/integration/database_test.go` | 迁移 + repo。 |
| `tests/integration/cache_test.go` | Redis / testcontainer。 |

---

## Sprint 8（第 15–16 周）— gRPC、OpenAPI、配置与性能测试（可选）

（与原计划一致。）

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/api/grpc/server.go` | gRPC 服务与 graceful stop。 |
| `internal/api/grpc/plan_service.go` | PlanService 实现。 |
| `internal/api/grpc/interceptor/auth.go` | Metadata 鉴权。 |
| `internal/api/grpc/interceptor/logging.go` | 日志与耗时。 |
| `api/openapi/plan.yaml`、`poi.yaml`、`swagger.yaml` | OpenAPI。 |
| `api/proto/plan.proto`、`poi.proto`、`common.proto` | Proto。 |
| `configs/config*.yaml` | 多环境配置。 |
| `configs/solver/*.yaml` | 求解器参数。 |
| `tests/benchmark/solver_bench_test.go` | 求解器基准。 |
| `tests/benchmark/planner_bench_test.go` | 引擎端到端基准。 |

---

## 推荐执行顺序（一线开发 checklist）

1. ✅ Domain、✅ planner 三文件、✅ postgres 仓储 — 保持与 migration 一致。  
2. 🔲 **Migration + seed**（对齐现有 SQL 与 repo）。  
3. 🔲 **`internal/core/solver`：`solver.go` + `stub`（或 naive）**，接好 `TripSolver`，跑通集成测试。  
4. 🔲 约束 + 遗传算法深化；蚁群 + hybrid + `core/planner` 引擎。  
5. 🔲 Redis（可选紧跟 stub 压力测试前）。  
6. 🔲 Recommender + Feedback（与 4–5 并行亦可）。  
7. 🔲 最小 HTTP（6A）→ Wire 与全套 bootstrap（6B）。  
8. 🔲 MQ、Worker、观测性、集成测试；最后 gRPC/OpenAPI/压测视需求。

---

## 说明

- Sprint 5、8 体量大：可砍掉 Neo4j / gRPC / 部分 yaml，把周次顺延或合并。  
- 单测文件（如 `tests/unit/domain`、`tests/unit/solver`）不逐文件列：开发对应包时同步补 `_test.go` 即可。

- 单测文件（如 `tests/unit/domain`、`tests/unit/solver`）不逐文件列：开发对应包时同步补 `_test.go` 即可。
