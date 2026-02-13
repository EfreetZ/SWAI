# SWAI — Software Architect Infrastructure

## 后端架构师成长型开源项目（从语法基础到架构师）

> **核心语言：Golang** | Docker 部署 | GitHub 托管
>
> Build Systems. Break Systems. Fix Systems. **Become an Architect.**

---

## 项目核心目标

本项目的目标不是简单"做项目"，而是构建一条**可执行、可复现、可扩展**的成长路径，使学习者从后端基础逐步成长为具备真实系统设计与架构能力的工程师。

**核心原则：**

- **项目驱动学习** — 每个知识点必须落地为真实代码
- **问题驱动成长** — 主动构造生产级问题并完成完整闭环
- **工程化优先** — 所有内容遵循真实公司开发流程
- **容器化环境** — 全部基于 Docker 部署
- **GitHub 托管** — 保证可协作与可持续迭代
- **单语言主线** — Golang 贯穿全部，避免多语言认知切换
- **实现 : 源码 = 7 : 3** — 优先自己实现，辅以源码阅读

---

## 使用说明

- 每个 `[ ]` 可标记为 `[x]` 表示完成
- 可自由增删改任意条目，我会根据此计划表生成代码
- 每个 Stage 对应 `projects/` 下的独立项目目录
- 每个项目遵循**工程闭环**（见下方）

---

## 工程闭环模型（最重要）

**每个技术模块必须遵循统一的 7 步闭环：**

```
1. 构建系统 ──── 写出可运行的基础版本
       │
2. 制造问题 ──── 人为注入高并发、延迟、数据不一致等故障
       │
3. 定位问题 ──── 使用专业工具进行分析（pprof / trace / tcpdump）
       │
4. 解决问题 ──── 优化架构或代码
       │
5. 验证测试 ──── 压测 + 单元测试 + 混沌测试
       │
6. 监控系统 ──── 建立指标与告警（Prometheus / Grafana）
       │
7. 复盘沉淀 ──── 输出问题分析文档，更新排障手册
```

> 目标：让你具备 **"系统出故障 → 快速定位 → 稳定修复"** 的高级工程能力。

---

## 架构成长的三次跃迁

```
CRUD 工程师 ──► 组件工程师 ──► 系统工程师 ──► 分布式工程师 ──► 架构师
                 Stage 0-3       Stage 4-6        Stage 7-8       Stage 9-11
```

- **第一次跃迁（Stage 0-3）：** 建立工程基础，理解组件原理
- **第二次跃迁（Stage 4-6）：** 构建基础设施，而非业务系统（mini-mysql / mini-redis / mini-mq）
- **第三次跃迁（Stage 7-8）：** 处理多节点问题（RPC / Raft / 分布式协调）
- **第四次跃迁（Stage 9-11）：** 流量治理 + 可观测性 + 架构级系统设计

---

## 中间件五层掌控模型

> 每个中间件（MySQL / Redis / Kafka）都要经过这五层：

| 层级 | 内容 | 关键问题 |
|------|------|---------|
| 1. Production Usage | 像生产环境一样部署（Docker Cluster + 压测） | 系统极限在哪？ |
| 2. Architecture | 只关注：数据如何存？请求如何流动？为什么快？ | 避免源码崇拜 |
| 3. Failure & Debug | 固定排障路径：CPU/IO → 连接 → 慢查询/热Key → 火焰图 | 故障定位能力 |
| 4. Performance | 只优化最慢的 20%，建立性能直觉 | 瓶颈在哪？ |
| 5. Rewrite It | 自己实现 mini 版本，第一性原理学习 | 本质是什么？ |

---

# Stage 0 — 工程基础

> 预计周期：2 周 | 目标：建立专业工程习惯，而非"能跑就行"
>
> 设计文档：`docs/design/stage0-engineering.md`

- [ ] Golang 项目结构设计（`cmd/` + `internal/` + `pkg/`）
- [ ] Makefile / Taskfile
- [ ] Git 分支模型（main / develop / feature / hotfix）
- [ ] Commit 规范（Conventional Commits）
- [ ] Dockerfile 编写（多阶段构建）
- [ ] docker-compose 环境编排
- [ ] GitHub Actions CI（lint + test + race）
- [ ] 单元测试 + Benchmark 规范
- [ ] 结构化日志（slog / zap）

### 输出项目

👉 `projects/stage0-engineering-template/` — 可直接复用的企业级 Golang 脚手架

---

# Stage 1 — Golang 核心能力

> 预计周期：3 周 | 目标：理解语言如何支撑高并发系统
>
> 设计文档：`docs/design/stage1-go-core.md`

## 1.1 语法基础

- [ ] 类型系统（int/string/bool/byte/rune/struct/interface）
- [ ] 控制流（if/for/switch/select）
- [ ] 函数（多返回值、闭包、defer、panic/recover）
- [ ] 指针与值传递
- [ ] 结构体与方法（值接收者 vs 指针接收者）
- [ ] 接口（隐式实现、空接口、类型断言）
- [ ] 错误处理模式（error wrapping、sentinel error）
- [ ] 泛型基础（Go 1.18+）

## 1.2 并发编程

- [ ] Goroutine 调度模型（GMP）
- [ ] Channel（buffered/unbuffered、select）
- [ ] sync 包（Mutex、RWMutex、WaitGroup、Once、Pool、Map）
- [ ] Context（WithCancel/WithTimeout/WithDeadline）
- [ ] 并发模式（fan-in、fan-out、pipeline、worker pool）
- [ ] 数据竞争检测（`go test -race`）

## 1.3 数据结构与算法

- [ ] 切片底层结构与扩容机制
- [ ] Go map 底层（hmap → bmap → overflow → 渐进式扩容）
- [ ] 链表（单链表、双向链表）
- [ ] 栈 / 队列 / 堆
- [ ] 二叉搜索树 / AVL / 红黑树概念
- [ ] B+Tree（完整实现）
- [ ] 跳表 SkipList（完整实现）
- [ ] 一致性哈希（完整实现）
- [ ] 排序算法（快排、归并、堆排序）

## 1.4 网络编程基础

- [ ] TCP/IP 协议栈基础
- [ ] Go `net` 包（TCP Server/Client）
- [ ] HTTP 协议 + Go `net/http`
- [ ] JSON 序列化

### 🔥 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| Goroutine 泄漏 | `pprof goroutine profile` |
| 死锁 | `dlv` 调试器 + goroutine stack |
| CPU 飙高 | `pprof CPU profile` + 火焰图 |
| 内存暴涨 | `pprof heap` + `go tool compile -m`（逃逸分析） |
| 数据竞争 | `go test -race` |

### 🎯 面试考点

- [ ] GMP 调度模型详解（G/M/P 关系、Work Stealing、抢占式调度）
- [ ] `defer` 执行顺序与 `panic` 的关系
- [ ] 值接收者 vs 指针接收者的区别与 interface 实现影响
- [ ] `interface{}` 的内部结构（iface / eface）
- [ ] Goroutine 泄漏的常见原因与排查
- [ ] `sync.Map` vs `map+RWMutex` 的适用场景
- [ ] Go slice 扩容策略
- [ ] Go map 底层结构与扩容
- [ ] B+Tree vs 红黑树的工程取舍
- [ ] TCP 三次握手 / 四次挥手 / TIME_WAIT

### 输出项目

👉 `projects/stage1-go-core/` — 语法练习 + 并发实验 + 数据结构实现 + 基准测试

---

# Stage 2 — 基础组件自研

> 预计周期：3 周 | 目标：理解"中间件本质"——**让你真正理解 Redis / MQ 为什么存在**
>
> 设计文档：`docs/design/stage2-mini-components.md`

## 核心实现

- [ ] LRU Cache（HashMap + 双向链表）
- [ ] LFU Cache（HashMap + 频率桶）
- [ ] 延迟队列（时间轮 / 最小堆）
- [ ] Worker Pool（goroutine 池化复用）
- [ ] 熔断器（Closed → Open → HalfOpen 三态转换）
- [ ] 限流器（令牌桶 Token Bucket）
- [ ] 布隆过滤器（Bloom Filter）

### 🔥 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 缓存击穿（热点 key 失效） | singleflight + 分布式锁 |
| 缓存雪崩（大量 key 同时过期） | 随机 TTL + 多级缓存 |
| 缓存穿透（请求不存在的 key） | 布隆过滤器 + 空值缓存 |
| 队列堆积 | Consumer 监控 + 动态扩容 |

### 🎯 面试考点

- [ ] LRU vs LFU 的适用场景？
- [ ] 时间轮的实现原理？对比最小堆？
- [ ] Worker Pool 如何优雅关闭？
- [ ] 熔断器三态转换的条件？
- [ ] 令牌桶 vs 漏桶 vs 滑动窗口的区别？
- [ ] 布隆过滤器的误判率与空间关系？

### 输出项目

👉 `projects/stage2-mini-components/` — 每个组件独立模块 + 完整测试 + 基准测试

---

# Stage 3 — Web 系统演进

> 预计周期：3 周 | 目标：体验从单体到微服务的架构升级路径
>
> 设计文档：`docs/design/stage3-web-service.md`

## 架构升级路径

```
单体系统 → 模块化 → 服务拆分 → 微服务雏形
```

## 核心实现

- [ ] 用户系统（注册 / 登录 / 个人信息）
- [ ] JWT 鉴权
- [ ] RBAC 权限模型
- [ ] 请求日志中间件
- [ ] Panic Recovery 中间件
- [ ] 优雅关闭（Graceful Shutdown）
- [ ] 参数校验
- [ ] 统一错误响应（`{code, data, message}`）
- [ ] Swagger API 文档

### 🔥 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 慢请求 | tracing + pprof |
| DB 被打爆 | 连接池监控 + 慢 SQL 日志 |
| Goroutine 耗尽 | `pprof goroutine` + 超时控制 |
| 接口鉴权绕过 | 安全审计 + 单元测试 |

### 🎯 面试考点

- [ ] JWT Token 过期与刷新机制
- [ ] RBAC vs ABAC 权限模型
- [ ] Go HTTP Server 优雅关闭的实现
- [ ] 中间件链的执行顺序
- [ ] RESTful API 设计规范

### 🐳 Docker 部署

- [ ] docker-compose：App + MySQL + Redis
- [ ] 环境变量配置管理

### 输出项目

👉 `projects/stage3-web-service/` — 完整用户服务 + Docker 一键启动

---

# Stage 4 — 中间件深度：mini-mysql

> 预计周期：2 个月 | 设计文档：`docs/design/stage4-mini-mysql.md`
>
> 这是成长为架构师的**分水岭**，进入中间件深度路线。
>
> 高可用 & 线上故障：`docs/design/middleware-ha-troubleshooting.md` → MySQL 篇

## 核心实现

- [ ] Page Manager（固定大小页读写）
- [ ] B+Tree Index（插入/删除/范围查询）
- [ ] WAL（Write-Ahead Logging）
- [ ] Buffer Pool Manager（LRU 置换）
- [ ] 简化事务（BEGIN/COMMIT/ROLLBACK）
- [ ] 简化 SQL 解析器（Mini SQL Parser）
- [ ] TCP Server

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 慢 SQL | `EXPLAIN` 执行计划分析 |
| 索引失效 | 执行计划 type=ALL 检测 |
| 死锁 | `SHOW ENGINE INNODB STATUS` + 锁分析 |
| IO 瓶颈 | `iostat` / `iotop` |

### 🎯 面试考点

- [ ] B+Tree 为什么适合磁盘存储？
- [ ] WAL 如何保证 crash safety？
- [ ] Buffer Pool 的 LRU 和 MySQL 的 LRU 区别（young/old 分代）
- [ ] 聚簇索引 vs 非聚簇索引
- [ ] 事务 ACID 各自如何保证？
- [ ] redo log vs undo log vs binlog
- [ ] 分库分表策略

### � Docker 部署

- [ ] Dockerfile（多阶段构建）
- [ ] docker-compose.yml（服务 + 持久化卷）
- [ ] sysbench 容器化压测

### 输出项目

👉 `projects/stage4-mini-mysql/`

---

# Stage 5 — 中间件深度：mini-redis

> 预计周期：2 个月 | 设计文档：`docs/design/stage5-mini-redis.md`
>
> 高可用 & 线上故障：`docs/design/middleware-ha-troubleshooting.md` → Redis 篇

## 核心实现（版本递进）

- [ ] V0：TCP + RESP 协议 + 基础 KV
- [ ] V1：Pipeline + TTL（惰性删除 + 定期删除）+ 多数据结构
- [ ] V2：AOF + RDB Snapshot 持久化
- [ ] V3：Master-Slave 主从复制
- [ ] V4：Cluster + 一致性哈希分片

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 热 Key | `redis-cli --hotkeys`（需 LFU） |
| 大 Key | `redis-cli --bigkeys` / `MEMORY USAGE` |
| 缓存穿透 / 雪崩 | 布隆过滤器 + 随机 TTL + 多级缓存 |
| 连接数过多 | `CLIENT LIST` + 连接池优化 |

### 🎯 面试考点

- [ ] Redis 单线程为什么快？
- [ ] 缓存雪崩 / 击穿 / 穿透的区别与解决方案
- [ ] AOF vs RDB 优缺点
- [ ] 主从复制全量 vs 增量同步
- [ ] Redis Cluster 的 slot 机制
- [ ] 大 Key / 热 Key 的发现与治理
- [ ] 内存淘汰策略（LRU/LFU/random/ttl）
- [ ] 分布式锁（SETNX + 过期 + Lua）

### � Docker 部署

- [ ] docker-compose：3 主 3 从集群
- [ ] redis-benchmark 容器化压测

### 输出项目

👉 `projects/stage5-mini-redis/`

---

# Stage 6 — 中间件深度：mini-mq

> 预计周期：2 个月 | 设计文档：`docs/design/stage6-mini-mq.md`
>
> 高可用 & 线上故障：`docs/design/middleware-ha-troubleshooting.md` → Kafka 篇

## 核心实现

- [ ] Segment 文件 + 稀疏索引
- [ ] Topic + Partition 模型
- [ ] Producer 批量发送
- [ ] Consumer + Offset 管理
- [ ] Consumer Group + Rebalance
- [ ] 日志清理（retention）

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 消息积压 | Consumer Lag 监控 |
| 重复消费 | 幂等处理 + Offset 管理 |
| 顺序错乱 | 单 Partition 保序 + Key 路由 |
| Broker 磁盘满 | retention 策略 + 磁盘告警 |

### 🎯 面试考点

- [ ] Kafka 为什么这么快？（顺序写 + 零拷贝 + 批量 + 分区并行）
- [ ] Consumer Group Rebalance 的流程与问题
- [ ] 消息丢失场景分析（Producer/Broker/Consumer 三端）
- [ ] 如何保证消息顺序性？
- [ ] exactly-once 语义如何实现？
- [ ] ISR 机制

### � Docker 部署

- [ ] docker-compose：3 Broker 集群
- [ ] Producer / Consumer 容器化压测

### 输出项目

👉 `projects/stage6-mini-mq/`

---

# Stage 7 — RPC Framework: mini-rpc

> 预计周期：1.5 个月 | 设计文档：`docs/design/stage7-mini-rpc.md`

## 核心实现

- [ ] 自定义二进制协议
- [ ] 多序列化（JSON / Protobuf / Msgpack）
- [ ] 服务注册与反射调用
- [ ] 服务发现（Memory / etcd）
- [ ] 负载均衡（RoundRobin / WeightedRR / ConsistentHash / LeastConn）
- [ ] 超时控制 + 重试 + 熔断器
- [ ] 中间件链（Logging / Metrics / RateLimit / Auth）
- [ ] 连接池

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 连接泄漏 | goroutine / fd 数量监控 |
| 超时链路不清 | 逐段耗时埋点 + tracing |
| 重试风暴 | 指数退避 + 最大重试次数 + jitter |
| 服务雪崩 | 熔断器 + 降级策略 |

### 🎯 面试考点

- [ ] RPC vs HTTP 的区别？
- [ ] gRPC 为什么用 HTTP/2？
- [ ] 服务发现的推 vs 拉模型
- [ ] 负载均衡算法对比
- [ ] 熔断器三态转换
- [ ] 优雅关闭的实现
- [ ] 重试风暴如何避免？

### � Docker 部署

- [ ] docker-compose：3 Server + 注册中心
- [ ] 模拟节点故障场景

### 输出项目

👉 `projects/stage7-mini-rpc/`

---

# Stage 8 — 分布式系统: mini-raft & 协调

> 预计周期：1.5 个月 | 设计文档：`docs/design/stage8-distributed.md`
>
> 走到这里，能力已经超越多数高级工程师。

## 核心实现

- [ ] Raft 共识算法（Leader Election + Log Replication）
- [ ] 持久化 + Snapshot
- [ ] 上层 KV Service（线性一致性读写）
- [ ] 分布式锁
- [ ] 分布式 ID 生成器（雪花算法）
- [ ] 服务注册发现

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 网络分区 | Jepsen 风格一致性检验 |
| Leader 崩溃 | 自动选举验证 + 日志对比 |
| 脑裂 | Quorum 验证 + 多数派检查 |
| 时钟漂移 | 逻辑时钟（Term）追踪 |

### 🎯 面试考点

- [ ] CAP 定理的含义与取舍
- [ ] Raft vs Paxos 的区别
- [ ] Raft 如何保证 Safety？
- [ ] 脑裂如何避免？
- [ ] etcd 的使用场景
- [ ] 线性一致性 vs 最终一致性
- [ ] 分布式 ID 生成方案对比（雪花 / UUID / 号段）

### 🐳 Docker 部署

- [ ] docker-compose：3 / 5 节点集群
- [ ] 混沌测试脚本（kill 节点、网络分区）

### 输出项目

👉 `projects/stage8-distributed/`

---

# Stage 9 — API Gateway: mini-gateway

> 预计周期：1.5 个月 | 设计文档：`docs/design/stage9-mini-gateway.md`

## 核心实现

- [ ] 反向代理 + Radix Tree 路由
- [ ] 负载均衡 + 健康检查
- [ ] 限流（Token Bucket / Sliding Window）
- [ ] 熔断
- [ ] JWT / API Key 认证
- [ ] CORS 处理
- [ ] Filter Chain 插件机制
- [ ] Prometheus Metrics + Access Log
- [ ] 配置热加载

### � 构造问题 & 定位手段

| 构造问题 | 定位手段 |
|---------|---------|
| 延迟飙高 | 后端 vs 网关分段计时 |
| 限流误判 | 压测验证 + 限流日志 |
| 配置变更故障 | 配置回滚 + 灰度发布 |
| 后端全部挂掉 | 熔断 + 降级响应 |

### 🎯 面试考点

- [ ] 限流算法对比（令牌桶 vs 漏桶 vs 滑动窗口）
- [ ] 分布式限流如何实现？（Redis + Lua）
- [ ] 熔断 vs 降级 vs 限流
- [ ] 网关在微服务架构中的角色
- [ ] 灰度发布 / 蓝绿部署 / 金丝雀发布

### 🐳 Docker 部署

- [ ] docker-compose：Gateway + 3 后端 + Prometheus + Grafana
- [ ] wrk / k6 容器化压测

### 输出项目

� `projects/stage9-mini-gateway/`

---

# Stage 10 — 可观测性体系

> 预计周期：3 周 | 目标：架构师必备能力
>
> 设计文档：`docs/design/stage10-observability.md`

## 核心实现

- [ ] Prometheus 指标采集 + 自定义 Metrics
- [ ] Grafana Dashboard（RED 指标：Rate / Error / Duration）
- [ ] OpenTelemetry 链路追踪（分布式 tracing）
- [ ] Loki 日志聚合
- [ ] 告警规则（Alertmanager）
- [ ] 错误预算（Error Budget）概念实践

### 关键指标体系

```
RED 指标（面向服务）：
  Rate    — 请求速率（QPS）
  Error   — 错误率
  Duration — 延迟分布（P50 / P95 / P99）

USE 指标（面向资源）：
  Utilization — 使用率
  Saturation  — 饱和度
  Errors      — 错误数
```

### 🎯 面试考点

- [ ] Metrics / Logging / Tracing 三支柱的区别与协同
- [ ] Prometheus 的 pull 模型 vs push 模型
- [ ] 如何设计告警规则避免告警风暴？
- [ ] 链路追踪如何实现跨服务传播？（trace_id / span_id）
- [ ] SLI / SLO / SLA 的定义与关系

### 🐳 Docker 部署

- [ ] docker-compose：Prometheus + Grafana + Loki + OpenTelemetry Collector
- [ ] 预置 Dashboard 模板

### 输出项目

👉 `projects/stage10-observability/`

---

# Stage 11 — 终极实战：Mini 电商平台

> 预计周期：2 个月 | 目标：架构级综合项目
>
> 设计文档：`docs/design/stage11-mini-ecommerce.md`

## 微服务架构

```
                    ┌─────────────┐
                    │ API Gateway │  ← Stage 9
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
    │ User Service│ │Order Service│ │Product Svc  │
    └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
           │               │               │
    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
    │   MySQL     │ │   MySQL     │ │   Redis     │
    └─────────────┘ └──────┬──────┘ └─────────────┘
                           │
                    ┌──────▼──────┐
                    │   Kafka     │  ← 异步解耦
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │Inventory Svc│
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Payment Svc│
                    └─────────────┘
```

## 核心服务

- [ ] API Gateway（mini-gateway 复用）
- [ ] 用户服务（注册 / 登录 / JWT）
- [ ] 商品服务（CRUD + 缓存）
- [ ] 订单服务（下单 + 状态机）
- [ ] 库存服务（扣减 + 防超卖）
- [ ] 支付服务（模拟支付流程）

## 必须解决的架构问题

- [ ] 分布式事务（Saga / TCC / 最终一致性）
- [ ] 缓存一致性（Cache-Aside / 延迟双删）
- [ ] 流量洪峰（秒杀场景：限流 + 队列削峰 + 预扣库存）
- [ ] 服务降级（非核心服务故障时的兜底策略）
- [ ] 幂等设计（订单防重复提交）
- [ ] 分布式 ID（雪花算法 / 号段模式）

### 🎯 面试考点

- [ ] 如何设计一个秒杀系统？（完整链路分析）
- [ ] 分布式事务的常见方案与取舍
- [ ] 如何做容量评估？
- [ ] 高可用设计原则（冗余 / 隔离 / 限流 / 降级 / 超时 / 重试）
- [ ] 微服务拆分原则
- [ ] 服务间通信（同步 RPC vs 异步 MQ）
- [ ] 服务网格概念（Service Mesh / Sidecar）

### 🐳 Docker 部署

- [ ] docker-compose：全部服务 + 中间件 + 可观测性栈
- [ ] 一键启动整个电商平台
- [ ] k6 压测脚本模拟秒杀场景

### 输出项目

� `projects/stage11-mini-ecommerce/`

---

# 统一问题定位方法论

> 详细文档：`docs/design/debugging-methodology.md`

## 三层排障模型

遇到故障时，按固定路径逐层排查：

### 第一层：资源瓶颈

```
CPU → 内存 → 磁盘 IO → 网络
工具：top / htop / iostat / iotop / netstat / ss
```

### 第二层：系统瓶颈

```
Goroutine 数 → 锁竞争 → 队列堆积 → 连接池
工具：pprof / trace / dlv / runtime metrics
```

### 第三层：架构瓶颈

```
DB 是否单点？→ 缓存是否失效？→ 是否同步调用？→ 是否缺少降级？
方法：架构审查 → 链路追踪 → 容量评估
```

> **80% 的系统问题，本质都是架构问题。**

## 必备工具箱

| 工具 | 用途 | 对应阶段 |
|------|------|---------|
| `pprof` | CPU / Memory / Goroutine 分析 | Stage 1+ |
| `trace` | Go 运行时调度追踪 | Stage 1+ |
| `dlv` | Go 调试器 | Stage 1+ |
| `go test -race` | 数据竞争检测 | Stage 1+ |
| `tcpdump` / `wireshark` | 网络抓包 | Stage 4+ |
| `iostat` / `iotop` | 磁盘 IO 分析 | Stage 4+ |
| `wrk` / `k6` | HTTP 压测 | Stage 3+ |
| `docker stats` | 容器资源监控 | Stage 0+ |
| `Prometheus` + `Grafana` | 指标监控 | Stage 10 |
| `OpenTelemetry` | 链路追踪 | Stage 10 |

---

# 仓库目录结构

```
SWAI/
├── ROADMAP.md                          # ← 主路线图（你正在看的文件）
├── README.md                           # 项目介绍
├── Makefile                            # 顶层构建命令
├── .gitignore
├── .github/workflows/ci.yml            # GitHub Actions CI
│
├── docs/
│   ├── backend_architect_growth.md     # 核心蓝图（项目原始设计）
│   ├── ARCHITECT_TRAINING_OS.md        # 架构成长理念
│   └── design/                         # 各阶段详细设计文档
│       ├── stage0-engineering.md
│       ├── stage1-go-core.md
│       ├── stage2-mini-components.md
│       ├── stage3-web-service.md
│       ├── stage4-mini-mysql.md
│       ├── stage5-mini-redis.md
│       ├── stage6-mini-mq.md
│       ├── stage7-mini-rpc.md
│       ├── stage8-distributed.md
│       ├── stage9-mini-gateway.md
│       ├── stage10-observability.md
│       ├── stage11-mini-ecommerce.md
│       ├── debugging-methodology.md    # 排障方法论
│       └── middleware-ha-troubleshooting.md  # 中间件高可用 & 线上故障
│
├── projects/
│   ├── stage0-engineering-template/    # 工程脚手架
│   ├── stage1-go-core/                 # Go 核心能力
│   ├── stage2-mini-components/         # 基础组件自研
│   ├── stage3-web-service/             # Web 系统演进
│   ├── stage4-mini-mysql/              # 存储引擎
│   ├── stage5-mini-redis/              # 缓存系统
│   ├── stage6-mini-mq/                 # 消息队列
│   ├── stage7-mini-rpc/                # RPC 框架
│   ├── stage8-distributed/             # 分布式协调
│   ├── stage9-mini-gateway/            # API 网关
│   ├── stage10-observability/          # 可观测性
│   └── stage11-mini-ecommerce/         # 终极实战
│
├── docker/                             # 通用 Docker 模板
│   ├── go-app.Dockerfile
│   └── docker-compose.base.yml
│
└── scripts/
    ├── setup.sh                        # 环境初始化
    └── bench.sh                        # 通用压测脚本
```

---

# 执行时间线

> 每天 2~3 小时，约 12 个月完成

| 阶段 | 内容 | 周期 | 累计 |
|------|------|------|------|
| Stage 0 | 工程基础 | 2 周 | 0.5 月 |
| Stage 1 | Go 核心能力 | 3 周 | 1.25 月 |
| Stage 2 | 基础组件自研 | 3 周 | 2 月 |
| Stage 3 | Web 系统演进 | 3 周 | 2.75 月 |
| Stage 4 | mini-mysql | 2 月 | 4.75 月 |
| Stage 5 | mini-redis | 2 月 | 6.75 月 |
| Stage 6 | mini-mq | 2 月 | 8.75 月 |
| Stage 7 | mini-rpc | 1.5 月 | 10.25 月 |
| Stage 8 | 分布式 (Raft) | 1.5 月 | 11.75 月 |
| Stage 9 | API Gateway | 1.5 月 | 13.25 月 |
| Stage 10 | 可观测性 | 3 周 | 14 月 |
| Stage 11 | Mini 电商平台 | 2 月 | 16 月 |

> **前 12 个月（Stage 0-9）** 完成核心基础设施能力建设
>
> **后 4 个月（Stage 10-11）** 完成架构级综合项目

---

# 成长结果

完成后，你将具备：

- **中间件级理解** — 不只会用，而是知道为什么
- **分布式系统设计能力** — 能处理多节点问题
- **故障定位能力** — 从资源 → 系统 → 架构三层排查
- **架构思维** — 以数据流、控制流、故障路径看系统
- **高并发经验** — 真实压测数据支撑

衡量标准只有一个：

> **我是否越来越能驾驭复杂系统？**

👉 从"CRUD 工程师"跃迁为：**系统型工程师 / 架构师候选人**
