# Stage 1 — Go 核心能力（持续实现中）

> 对应设计文档：`docs/design/stage1-go-core.md`

## 当前实现范围

本次先落地 Stage 1 的基础骨架与第一批核心代码：

- `basics/types`：类型解析与泛型工具函数
- `basics/functions`：带 context 的通用重试函数
- `basics/structs`：结构体建模、校验与拷贝
- `basics/interfaces`：接口抽象 + 内存存储实现
- `basics/errors`：错误包装与判定
- `concurrency/patterns`：基于 `context.Context` 的 Worker Pool
- `datastructures/linkedlist`：泛型 LRU Cache
- `network/tcp`：TCP Echo Server
- `monitoring`：运行时指标采集
- `alerting`：基于阈值的告警评估
- `test/performance`、`test/stress`、`test/chaos`：性能/压力/混沌测试示例

## 目录结构

```text
projects/stage1-go-core/
├── basics/
│   ├── types/
│   ├── functions/
│   ├── structs/
│   ├── interfaces/
│   └── errors/
├── concurrency/patterns/
├── datastructures/linkedlist/
├── network/tcp/
├── monitoring/
├── alerting/
├── test/
│   ├── performance/
│   ├── stress/
│   └── chaos/
├── TROUBLESHOOTING.md
├── go.mod
└── Makefile
```

## 运行方式

```bash
make test
make race
make bench
make perf
make stress
make chaos
```

## 设计约束

- 并发相关代码必须显式使用 `context.Context` 控制超时与取消。
- 错误通过 `error` 返回，不使用 `panic`（测试除外）。
- 所有样例都提供单元测试，关键模块提供基准测试。
