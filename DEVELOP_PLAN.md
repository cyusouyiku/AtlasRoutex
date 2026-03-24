# 开发计划（2 周 / Sprint）

前提：`internal/domain`（entity / valueobject / repository）已写完。下列按 README 目录排期，未列文件表示该 Sprint 不动或后移。

---

## Sprint 1（第 1–2 周）— 应用层用例

**需求背景：** 领域模型和仓储接口已经定义，但还没有「一条可执行的业务流程」把用户意图串起来。HTTP、gRPC、数据库都还没接时，必须有一层只关心业务步骤的代码，否则后面会在 handler 或 SQL 里堆逻辑，难以测试和替换实现。

**要解决什么问题：** 把「规划行程、调整行程、推荐 POI、收集反馈」拆成独立用例；用例只依赖 `domain` 类型和 `repository` 接口（以及后续会注入的 solver 端口），这样团队可以并行写基础设施和算法，用 Mock 就能验证业务流程是否正确。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/application/planner/plan_usecase.go` | 实现「从输入到可保存行程」的编排：读用户/约束、拉候选 POI（经仓储）、组装求解器输入、调用求解端口、把结果写回 `Itinerary` 并交给仓储。需定义清晰的错误语义（如无候选、求解超时、预算不可行），不在此文件写 SQL 或 HTTP。 |
| `internal/application/planner/adjust_usecase.go` | 覆盖「用户改某一天、换一个景点、改时段」等场景：加载现有行程与变更意图，计算受影响的最小片段，调用增量求解或全量重算策略（可先留接口实现为全量）。负责并发/版本冲突策略的占位（如乐观锁字段）。 |
| `internal/application/planner/dto.go` | 定义规划与调整用例的输入输出结构（目的地、天数、预算 VO 的序列化形式、偏好标签等），与领域实体解耦，避免 API 变更直接改 entity。可含校验方法或与 `pkg/utils/validator` 的衔接点。 |
| `internal/application/recommender/recommend_usecase.go` | 对外提供「给定上下文返回 POI 列表」的用例：上下文包括用户画像、当前行程片段、地理范围；内部调用 `strategy` 与 `PoiRepository`，输出带分数或解释字段的 DTO，供规划前筛选或前端展示。 |
| `internal/application/recommender/strategy.go` | 实现具体推荐逻辑：规则过滤（类别、距离、营业时间）、简单加权打分、或预留调用外部召回的钩子。保持无状态或显式依赖注入，便于 A/B 多种策略并存。 |
| `internal/application/feedback/feedback_usecase.go` | 接收用户对 POI/行程的反馈（喜欢/不喜欢/备注），决定写入数据库、发 MQ 或仅记日志的入口；与领域中的用户/行程 ID 对齐，并触发后续画像更新（可先异步占位）。 |
| `internal/application/feedback/collector.go` | 对原始反馈做校验（必填字段、枚举合法）、去重、批量聚合（如同步多条评分），再交给 `feedback_usecase`，避免用例里充斥数据清洗代码。 |

---

## Sprint 2（第 3–4 周）— 迁移 + PostgreSQL + Redis

**需求背景：** 用例层需要真实数据才能联调。没有版本化的表结构与迁移工具，多人协作时环境不一致会导致「我本地能跑你那边挂」。POI 与行程数据量大、读多写少，仅靠 PG 顶不住热点读与距离矩阵这类重复计算，需要 Redis 做缓存层。

**要解决什么问题：** 把 `domain/repository` 里的接口落到可运行的存储上；迁移脚本保证从零建库可复现；PG 负责权威数据与事务；Redis 负责缓存 POI 列表、规划中间结果等，并统一 key 前缀与 TTL，避免与业务代码散落字符串拼接。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `cmd/migration/main.go` | 小型 CLI：指向 DSN、按版本顺序执行 `scripts` 下 SQL、记录已执行版本（可用 schema_migrations 表）。支持 dry-run 或仅打印待执行脚本（可选），失败时退出非零码便于 CI。 |
| `cmd/migration/scripts/001_init_schema.sql` | 创建与领域实体一致的表：POI、行程主表与日表/项表、用户、约束存储方式等；主键、外键、时间字段默认值与 `entity` JSON tag 对齐，减少 mapper 歧义。 |
| `cmd/migration/scripts/002_add_indexes.sql` | 为高频查询加索引：按城市/类别查 POI、按 user_id 查行程、按状态与时间范围筛选等；避免在 001 里一次堆太多，便于审查。 |
| `cmd/migration/scripts/003_seed_data.sql` | 开发/CI 用最小数据集：若干 POI、一条示例行程，保证 `plan_usecase` 集成测试有稳定输入。 |
| `internal/infrastructure/persistence/postgres/db.go` | 封装 `sql.Open`、连接池参数（max open、idle、lifetime）、健康检查 `Ping`；可暴露 `*sql.DB` 或 `*sqlx.DB` 给各 repo 实现复用。 |
| `internal/infrastructure/persistence/postgres/poi_repo_impl.go` | 实现 `PoiRepository`：按 ID/区域/标签查询、分页、批量获取；行与 `entity.POI`（或项目实际类型）的转换集中在此，处理 NULL 与枚举字符串。 |
| `internal/infrastructure/persistence/postgres/itinerary_repo_impl.go` | 实现 `ItineraryRepository`：创建/更新/加载完整行程（含嵌套天与活动）；考虑事务边界（一次保存多表）；与 `Itinerary` 状态机字段一致。 |
| `internal/infrastructure/persistence/redis/client.go` | 封装 go-redis 或等价客户端：从配置读地址、密码、DB 号；提供统一 `Get/Set/Del`、`SetNX`、TTL；可在此做 key 命名空间常量。 |
| `internal/infrastructure/persistence/redis/cache_repo_impl.go` | 实现 `CacheRepository`：序列化规划结果、POI 列表、距离矩阵片段；定义 TTL 与缓存击穿策略（如单飞锁可选）；与 PG 数据版本或 etag 配合失效（若接口支持）。 |

---

## Sprint 3（第 5–6 周）— 求解器接口 + 约束 + 遗传算法

**需求背景：** 行程规划本质是带大量约束的组合优化问题，手工 if-else 无法扩展。README 选定遗传算法作为主力之一，需要先有可替换的 solver 抽象，再实现约束检查与 GA 闭环，否则 application 层无法稳定调用「黑盒优化器」。

**要解决什么问题：** 定义统一的「问题实例 → 可行/近优解」接口；用约束传播与 AC-3 缩小搜索空间或预处理变量域；用遗传算法在离散序列空间上搜索_visit 顺序_；validator 保证输出满足硬约束（营业时间、预算上限等），软约束进 fitness。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/core/solver/solver.go` | 声明 `Solver` 接口与 `Problem`/`Solution` 结构：输入包含 POI 集合、距离/时间矩阵引用、`Budget`/`Timeslot`、约束列表；输出为日粒度 POI 序列或可直接映射到 `Itinerary` 的中间结构；支持 `context` 取消与超时。 |
| `internal/core/solver/constraint/propagator.go` | 构建约束图（变量为时段/POI 槽位等），调度弧修订与传播顺序；与 `entity.Constraint` 类型映射，输出缩减后的域或不可行判定。 |
| `internal/core/solver/constraint/ac3.go` | 实现 AC-3：维护弧队列，对每条弧执行 Revise，直至不动点或发现域为空；可参数化弧生成策略以适配不同约束类型。 |
| `internal/core/solver/constraint/validator.go` | 对一条完整候选解（POI 顺序 + 每站时间）逐项检查硬约束，返回违规列表与软约束惩罚分；供 GA 终筛与 hybrid 层择优。 |
| `internal/core/solver/genetic/chromosome.go` | 定义染色体编码与解码：如按天划分的 POI 排列、固定起点终点；提供合法化修复（重复 POI、必选点缺失时变异修复）。 |
| `internal/core/solver/genetic/crossover.go` | 实现顺序相关交叉（OX/PMX 等）或按天切块交换；保证子代在业务上可解码为行程草案。 |
| `internal/core/solver/genetic/mutation.go` | 交换、插入、反序等变异；可结合 validator 拒绝完全非法变异或做修复。 |
| `internal/core/solver/genetic/selection.go` | 锦标赛、轮盘赌或精英保留；保留多样性避免早熟收敛。 |
| `internal/core/solver/genetic/fitness.go` | 综合路程、时间窗违反惩罚、预算超出、用户偏好匹配等打分；硬约束违反可给极大惩罚或视为无效个体。 |

---

## Sprint 4（第 7–8 周）— 蚁群 + 混合求解 + 规划引擎

**需求背景：** 单一遗传算法在规模变大时可能收敛慢或陷入局部最优。蚁群适合路径类问题，与 GA 互补。产品还需要「分层、流水线、增量」等产品级能力，不能每次都从零把 solver 参数传给 application。

**要解决什么问题：** 提供第二种搜索路径（ACO）；用 hybrid 在超时内组合多算法或多次运行取最优；`core/planner` 作为 application 与多个 solver 之间的门面，封装分层（先选区再排点）、流水线阶段与增量重算，降低用例层复杂度。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/core/solver/antcolony/colony.go` | 蚂蚁迭代主循环：每只蚂蚁构造解、更新最优、直到迭代上限或收敛判定；对接 `validator` 保证输出可用。 |
| `internal/core/solver/antcolony/pheromone.go` | 信息素矩阵存储、挥发、增强；边对应 POI 间转移或图上弧；参数（α、β、ρ）可从配置注入。 |
| `internal/core/solver/antcolony/heuristic.go` | 计算启发式可见度（通常 1/距离 或偏好加权），与信息素相乘决定转移概率。 |
| `internal/core/solver/hybrid/orchestrator.go` | 按时间预算依次或并行调用 GA、ACO 等，捕获 panic/错误并降级；支持 context 超时后返回当前最优。 |
| `internal/core/solver/hybrid/ensemble.go` | 合并多轮解：去重、按 fitness 排序、或投票；可选输出 top-k 供前端展示备选方案。 |
| `internal/core/planner/engine.go` | 对 `plan_usecase` 暴露单一 `Plan(ctx, input)`：内部选择默认 solver 或 hybrid，返回领域友好的结果与诊断信息（如违反的软约束）。 |
| `internal/core/planner/hierarchical.go` | 粗粒度先划分区域或主题日，再在各子问题内调用 solver，最后拼接并做全局 validator 修正。 |
| `internal/core/planner/incremental.go` | 输入旧解 + 变更集，仅对受影响天或子路径重算，其余拷贝；失败时回退全量规划。 |
| `internal/core/planner/pipeline.go` | 串联「召回候选 → 规则剪枝 → 求解 → 后处理（如插入用餐时段）」；每步可替换实现，便于单测每一步。 |

---

## Sprint 5（第 9–10 周）— 多目标优化 + 图谱 + Neo4j + 地图 API

**需求背景：** 真实用户往往同时在意时间、钱、体验、少走回头路，单一标量 fitness 不好调权。POI 之间存在类别协同、地理聚类，纯表查询难表达「附近还有什么同类」。距离不能全靠直线，需要地图 API 做矩阵或路段。Neo4j 适合表达 POI 关系网络（相似、共存、主题边）。

**要解决什么问题：** 把多目标显式建模，支持权重配置与 Pareto 备选；用图查询加速「相关推荐」与结构化剪枝；用外部 mapapi 拉真实 OD 矩阵或批量距离，写入缓存供 solver 使用；Neo4j 实现可选，若砍 scope 可先做 `graph` 内存版 + mapapi。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/core/optimization/objective.go` | 定义多个可计算目标（总时长、总费用、偏好分、步行量等）及加权合成；支持归一化避免量纲碾压。 |
| `internal/core/optimization/weight.go` | 从配置或用户画像加载权重；提供默认模板（如亲子/穷游）；可记录本次规划使用的权重便于解释性输出。 |
| `internal/core/optimization/pareto.go` | 维护非支配解集；在 hybrid 输出多解时筛选展示用子集。 |
| `internal/core/graph/graph.go` | 领域侧图抽象：节点 POI、边类型（地理邻近、主题相似）；接口方法如 `Neighbors`, `WithinDistance`, `RelatedByCategory`。 |
| `internal/core/graph/builder.go` | 从 PG 批量导入或增量更新图；批处理避免 N+1；可对接离线任务写入 Neo4j。 |
| `internal/core/graph/query.go` | 高层查询：K-hop、路径模式、带过滤的子图；底层调用 Neo4j driver 或内存实现。 |
| `internal/core/graph/index/geohash.go` | 按网格快速筛候选 POI，减少进精确距离计算的规模。 |
| `internal/core/graph/index/rtree.go` | 二维近邻查询，支撑「当前点周围 N 公里」类需求。 |
| `internal/infrastructure/persistence/neo4j/driver.go` | 管理 Neo4j session/driver 生命周期、超时与重试策略。 |
| `internal/infrastructure/persistence/neo4j/graph_repo_impl.go` | 把 `graph.go` 的接口落到 Cypher；参数化查询防注入；结果映射回领域 ID。 |
| `internal/infrastructure/external/mapapi/client.go` | 封装第三方地图 REST：鉴权、QPS 限制、错误重试；统一返回结构供 `distance.go` 使用。 |
| `internal/infrastructure/external/mapapi/distance.go` | 批量起终点距离/时间矩阵、单段路径；结果可落 Redis；对失败格点回退直线距离（可配置）。 |

---

## Sprint 6（第 11–12 周）— Bootstrap + HTTP API + cmd/api + Wire

**需求背景：** 内部用例和核心引擎可测，但外部系统与用户需要 HTTP 入口。手动 `new` 依赖在文件间复制会失控，Wire 适合 Go 的编译期注入。没有统一的 bootstrap，每个 cmd 会重复读配置、连库、挂中间件。

**要解决什么问题：** 提供可启动的 API 进程：配置与日志一致、依赖一次性组装、路由与中间件可观测可保护；handler 薄，只做协议转换与状态码映射；错误与日志通过 `pkg/errors`、`pkg/logger` 统一，便于前后端联调与排障。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/bootstrap/config.go` | 加载 yaml/环境变量，解析 DSN、端口、第三方 API key；提供校验缺失项与默认值。 |
| `internal/bootstrap/logger.go` | 初始化 zap（或项目选定实现），设置 level、采样、开发/生产编码器差异。 |
| `internal/bootstrap/database.go` | 调用 postgres/redis（及 neo4j）的构造函数，注册到 Wire 的 provider；暴露 Close 给优雅退出。 |
| `internal/bootstrap/server.go` | 创建 `http.Server`，配置读写超时；`ListenAndServe` 与 `Shutdown` 分协程；可选 TLS（若需要）。 |
| `internal/bootstrap/app.go` | 组合 router、middleware、handler、全局中间件；`Run()` 阻塞直至信号。 |
| `internal/api/handler/plan_handler.go` | POST 创建规划、GET 查询结果、POST 调整等；绑定 JSON，调用 application usecase，映射错误到 HTTP 状态与 body。 |
| `internal/api/handler/poi_handler.go` | POI 列表、详情、搜索参数；分页与过滤查询参数解析。 |
| `internal/api/handler/user_handler.go` | 用户注册/登录占位或对接 IDP；与 `entity.User` 相关读写。 |
| `internal/api/handler/middleware/auth.go` | 解析 JWT/API Key，注入 user id 到 context；未授权返回 401。 |
| `internal/api/handler/middleware/ratelimit.go` | 按 IP 或用户限流，防刷规划接口；可对接 Redis 令牌桶。 |
| `internal/api/handler/middleware/logger.go` | 记录 method、path、status、latency、request id。 |
| `internal/api/handler/middleware/recovery.go` | Panic 捕获，打栈，返回 500 不泄露内部细节。 |
| `internal/api/dto/request.go` | HTTP 入参结构体与 `json`/`form` tag；与 application dto 分离。 |
| `internal/api/dto/response.go` | 统一包装 success/data/error 或分页结构。 |
| `internal/api/dto/converter.go` | request ↔ application dto ↔ domain 的转换函数，集中处理时间字符串解析等。 |
| `cmd/api/main.go` | 解析 flag/env，调用 bootstrap 启动；处理 SIGINT/SIGTERM。 |
| `cmd/api/wire.go` | 声明 `Injector` 与各层 `ProviderSet`，`//go:generate wire` 注释。 |
| `cmd/api/wire_gen.go` | Wire 生成，勿手改；CI 校验与 `wire.go` 一致。 |
| `pkg/errors/errors.go` | `Is`/`As` 友好包装、`WithStack` 可选；业务错误与系统错误区分。 |
| `pkg/errors/codes.go` | 稳定错误码字符串，供客户端 i18n 与监控聚合。 |
| `pkg/logger/logger.go` | 接口 `Info/Warn/Error/WithContext`，便于测试替换。 |
| `pkg/logger/zap.go` | 实现上述接口，对接具体 zap logger。 |

---

## Sprint 7（第 13–14 周）— MQ + 预订/支付 + Worker + 指标与集成测

**需求背景：** 规划与预计算耗 CPU，不能全堵在 API 请求线程里；反馈写入后可能要异步更新画像、触发重训。商业化需要对接 OTA 预订与支付，这些调用慢且易失败，适合独立客户端封装。上线需要可观测：QPS、延迟、缓存命中率、队列积压。

**要解决什么问题：** 用消息队列解耦 API 与后台任务；Worker 消费预计算与反馈主题；booking/payment 封装重试、幂等与超时；Prometheus 暴露指标；集成测试验证 API+DB+Redis 在近似生产配置下可用；pkg 提供队列抽象与通用工具减少重复。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/infrastructure/mq/producer.go` | 发送任务消息（JSON/Protobuf），带 trace id；失败重试与日志。 |
| `internal/infrastructure/mq/consumer.go` | 订阅、反序列化、分发到 handler 注册表；ack/nack 策略与并发度配置。 |
| `internal/infrastructure/mq/topics.go` | Topic 名、消费者组名常量，避免魔法字符串。 |
| `internal/infrastructure/external/booking/client.go` | 预订网关基类：通用 header、错误码映射、断路器占位。 |
| `internal/infrastructure/external/booking/hotel.go` | 酒店可订性查询、下单、取消；DTO 与内部领域字段映射。 |
| `internal/infrastructure/external/payment/client.go` | 创建支付单、查询状态、回调验签占位；密钥从配置注入。 |
| `cmd/worker/main.go` | 初始化与 API 相同的底层依赖（可共用 bootstrap 子集），启动 consumer 循环。 |
| `cmd/worker/handlers/precompute_handler.go` | 消费预计算任务：拉 POI 子集、算距离矩阵、写 Redis；幂等键防重复写。 |
| `cmd/worker/handlers/feedback_handler.go` | 异步落库、更新统计或转发分析管道；失败入死信或重试队列。 |
| `pkg/queue/queue.go` | 抽象 `Publish/Subscribe/Close`，便于单测 mock 与换 Kafka/Rabbit。 |
| `pkg/queue/kafka.go` | Kafka 实现（若不用可换别的 MQ 包名）。 |
| `pkg/metrics/prometheus.go` | 注册 histogram/counter/gauge，挂 `/metrics` 路由。 |
| `pkg/metrics/recorder.go` | 封装 `ObservePlanLatency`、`IncCacheHit` 等，业务代码只调语义化方法。 |
| `pkg/utils/hash.go` | 缓存键、幂等 id 用的稳定哈希。 |
| `pkg/utils/time.go` | 时区、当天边界、营业日对齐等与规划强相关的日期工具。 |
| `pkg/utils/math.go` | 距离、插值、简单统计，供 fitness 与报表复用。 |
| `pkg/utils/validator.go` | 邮箱、坐标范围、分页参数等通用校验，handler 与 dto 共用。 |
| `pkg/cache/cache.go` | 可选第二套缓存抽象：若与 `persistence/redis` 重复，可合并为单一实现，此处保留「业务级 GetPlanResult」等语义方法。 |
| `pkg/cache/memory.go` | 进程内 LRU/TTL，用于单测或降级。 |
| `pkg/cache/redis.go` | 对接 Redis 的通用 get/set；与 `cache_repo_impl` 分工：后者面向仓储接口，此处面向通用 KV。 |
| `tests/integration/api_test.go` | 起测试 server 或对接本地 docker，跑规划主链路 HTTP 断言。 |
| `tests/integration/database_test.go` | 迁移后插入数据，测 repo 与事务。 |
| `tests/integration/cache_test.go` | 测 Redis 读写与 TTL 行为（可用 miniredis 或 testcontainer）。 |

---

## Sprint 8（第 15–16 周）— gRPC、OpenAPI、配置与性能测试（可选）

**需求背景：** 移动端或 BFF 可能需要二进制协议降延迟；对外合作方需要机器可读的 API 合同。求解器参数调优依赖配置文件而非改代码。上线前要防止算法回归导致延迟爆炸。

**要解决什么问题：** 提供与 REST 并行的 gRPC 服务（或二选一）；OpenAPI/Proto 作为单一事实来源生成文档与客户端；环境分层配置与 solver 专项 yaml；基准测试建立 P95/P99 基线，CI 可选跑 benchmark 对比阈值。

### 交付文件

| 文件 | 职责与要点 |
|------|------------|
| `internal/api/grpc/server.go` | 注册 reflection（开发用）、TLS 可选、graceful stop；与 HTTP 共享部分 usecase。 |
| `internal/api/grpc/plan_service.go` | 实现 proto 生成的 `PlanService`：参数校验、调用 application、错误转 gRPC status。 |
| `internal/api/grpc/interceptor/auth.go` | 从 metadata 取 token，与 HTTP auth 逻辑复用同一校验函数。 |
| `internal/api/grpc/interceptor/logging.go` |  unary/stream 日志与耗时。 |
| `api/openapi/plan.yaml` | 规划相关 path、schema、example；可驱动 swagger-ui。 |
| `api/openapi/poi.yaml` | POI 查询与管理路径定义。 |
| `api/openapi/swagger.yaml` | `openapi 3` 聚合多个模块，统一 servers 与安全方案。 |
| `api/proto/plan.proto` | `Plan`、`Adjust` RPC 与 message，与领域字段对应关系在注释说明。 |
| `api/proto/poi.proto` | POI 服务定义。 |
| `api/proto/common.proto` | 分页、错误详情、坐标、金额等共用类型。 |
| `configs/config.yaml` | 全服务默认：端口、日志、依赖地址占位。 |
| `configs/config.dev.yaml` | 本地覆盖，小数据量、debug 日志。 |
| `configs/config.staging.yaml` | 预发：接近生产的连接池与限流。 |
| `configs/config.prod.yaml` | 生产：脱敏示例，敏感项仍走环境变量。 |
| `configs/solver/genetic.yaml` | 种群规模、交叉率、变异率、精英数、最大代数。 |
| `configs/solver/antcolony.yaml` | 蚂蚁数、迭代数、信息素参数。 |
| `configs/solver/constraints.yaml` | 各约束类型开关、默认软硬、惩罚系数。 |
| `tests/benchmark/solver_bench_test.go` | 固定规模实例，测 GA/ACO/hybrid 单次耗时与分配次数。 |
| `tests/benchmark/planner_bench_test.go` | 端到端含 mock 仓储或小型真实数据，测 `engine.Plan` 延迟。 |

---

## 说明

- Sprint 5、8 体量大：可砍掉 Neo4j / gRPC / 部分 yaml，把周次顺延或合并。
- 单测文件（如 `tests/unit/domain`、`tests/unit/solver`）不逐文件列：开发对应包时同步补 `_test.go` 即可。
