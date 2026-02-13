# SWAI — Software Architect Infrastructure

> 后端架构师成长型开源项目：从语法基础到系统架构师

Build Systems. Break Systems. Fix Systems. **Become an Architect.**

---

## 这是什么？

SWAI 是一套**工程化、可执行、可递进**的架构师成长训练系统。通过真实基础设施项目驱动学习，而非碎片化技术堆砌。

**核心语言：Golang** | 全部 Docker 部署 | GitHub 托管

**核心方法论：**

- **项目驱动** — 每个知识点落地为真实代码
- **问题驱动** — 主动构造故障 → 定位 → 修复 → 监控 → 预防
- **面试导向** — 覆盖高频面试考点，每个知识点有代码级验证
- **工程闭环** — 7 步闭环：构建 → 制造问题 → 定位 → 解决 → 测试 → 监控 → 复盘

---

## 学习路径

```
Stage 0:  工程基础（项目结构 / Docker / CI / 日志）
    ▼
Stage 1:  Go 核心能力（语法 / 并发GMP / 数据结构 / 网络）
    ▼
Stage 2:  基础组件自研（LRU / 延迟队列 / 熔断器 / 限流器）
    ▼     ← 理解中间件为什么存在
Stage 3:  Web 系统演进（用户服务 / JWT / RBAC / 优雅关闭）
    ▼     ← 分水岭：进入中间件深度路线
Stage 4:  mini-mysql（B+Tree / WAL / Buffer Pool / 事务）
    ▼
Stage 5:  mini-redis（RESP / TTL / AOF / 主从 / Cluster）
    ▼
Stage 6:  mini-mq（Segment / Partition / Consumer Group）
    ▼     ← 进入分布式系统
Stage 7:  mini-rpc（协议 / 服务发现 / 负载均衡 / 熔断）
    ▼
Stage 8:  mini-raft（Leader Election / Log Replication / 分布式锁）
    ▼     ← 进入架构师领域
Stage 9:  mini-gateway（路由 / 限流 / 熔断 / 认证 / 热加载）
    ▼
Stage 10: 可观测性（Prometheus / Grafana / OpenTelemetry / Loki）
    ▼
Stage 11: 终极实战 — Mini 电商平台（微服务 / 秒杀 / 分布式事务）
```

---

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/<your-username>/SWAI.git
cd SWAI

# 查看完整路线图
cat ROADMAP.md

# 运行测试
make test

# 运行基准测试
make bench

# 启动某个项目的 Docker 环境
make docker-up PROJECT=stage5-mini-redis
```

---

## 目录结构

```
SWAI/
├── ROADMAP.md                          # 主路线图（计划表，可增删改）
├── README.md                           # 项目介绍
├── Makefile                            # 顶层构建命令
│
├── docs/
│   ├── backend_architect_growth.md     # 核心蓝图
│   ├── ARCHITECT_TRAINING_OS.md        # 架构成长理念
│   └── design/                         # 各阶段详细设计文档
│
├── projects/                           # 各阶段项目代码
│   ├── stage0-engineering-template/
│   ├── stage1-go-core/
│   ├── stage2-mini-components/
│   ├── stage3-web-service/
│   ├── stage4-mini-mysql/
│   ├── stage5-mini-redis/
│   ├── stage6-mini-mq/
│   ├── stage7-mini-rpc/
│   ├── stage8-distributed/
│   ├── stage9-mini-gateway/
│   ├── stage10-observability/
│   └── stage11-mini-ecommerce/
│
├── docker/                             # 通用 Docker 模板
├── scripts/                            # 工具脚本
└── .github/workflows/                  # CI/CD
```

---

## 设计文档

| 阶段 | 设计文档 | 说明 |
|------|---------|------|
| Stage 0 | [工程基础](docs/design/stage0-engineering.md) | 项目结构 / Docker / CI |
| Stage 1 | [Go 核心](docs/design/stage1-go-core.md) | 语法 / 并发 / 数据结构 |
| Stage 2 | [组件自研](docs/design/stage2-mini-components.md) | LRU / 时间轮 / 熔断 / 限流 |
| Stage 3 | [Web 服务](docs/design/stage3-web-service.md) | 用户系统 / JWT / RBAC |
| Stage 4 | [mini-mysql](docs/design/stage4-mini-mysql.md) | B+Tree / WAL / Buffer Pool |
| Stage 5 | [mini-redis](docs/design/stage5-mini-redis.md) | 5 版本递进 |
| Stage 6 | [mini-mq](docs/design/stage6-mini-mq.md) | Segment / Partition / Consumer Group |
| Stage 7 | [mini-rpc](docs/design/stage7-mini-rpc.md) | 服务发现 / 负载均衡 / 熔断 |
| Stage 8 | [mini-raft](docs/design/stage8-distributed.md) | 选主 / 日志复制 / 分布式锁 |
| Stage 9 | [mini-gateway](docs/design/stage9-mini-gateway.md) | 限流 / 熔断 / 认证 / 可观测 |
| Stage 10 | [可观测性](docs/design/stage10-observability.md) | Prometheus / Grafana / Jaeger |
| Stage 11 | [电商平台](docs/design/stage11-mini-ecommerce.md) | 微服务 / 秒杀 / 分布式事务 |
| 通用 | [排障方法论](docs/design/debugging-methodology.md) | 三层模型 / 工具箱 / 模板 |
| 通用 | [中间件高可用 & 线上故障](docs/design/middleware-ha-troubleshooting.md) | MySQL/Redis/Kafka 集群部署 + 真实故障场景 |

---

## 执行时间线

> 每天 2~3 小时

| 阶段 | 内容 | 周期 | 累计 |
|------|------|------|------|
| Stage 0-1 | 工程基础 + Go 核心 | 5 周 | 1.25 月 |
| Stage 2-3 | 组件自研 + Web 系统 | 6 周 | 2.75 月 |
| Stage 4 | mini-mysql | 2 月 | 4.75 月 |
| Stage 5 | mini-redis | 2 月 | 6.75 月 |
| Stage 6 | mini-mq | 2 月 | 8.75 月 |
| Stage 7-8 | mini-rpc + Raft | 3 月 | 11.75 月 |
| Stage 9-10 | Gateway + 可观测性 | 2 月 | 13.75 月 |
| Stage 11 | Mini 电商平台 | 2 月 | 15.75 月 |

---

## 成长跃迁

```
CRUD 工程师 ──► 组件工程师 ──► 系统工程师 ──► 分布式工程师 ──► 架构师
                 Stage 0-3       Stage 4-6        Stage 7-8       Stage 9-11
```

衡量标准只有一个：

> **我是否越来越能驾驭复杂系统？**

当你开始用以下视角看系统时，说明你已迈入架构师视角：
- **数据流** — 数据如何存储、流动、转换
- **控制流** — 请求如何路由、调度、处理
- **故障路径** — 系统如何失败、恢复、自愈

---

## License

MIT
