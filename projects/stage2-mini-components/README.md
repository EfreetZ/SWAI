# Stage 2 — 基础组件自研：mini-components

> 对应设计文档：`docs/design/stage2-mini-components.md`

## 已实现组件

- `lru`：LRU 缓存（基础版 + 并发版 + TTL 版）
- `lfu`：LFU 缓存
- `delayqueue`：最小堆延迟队列 + 时间轮
- `workerpool`：带 `context.Context` 的任务池
- `circuitbreaker`：三态熔断器
- `ratelimiter`：令牌桶 + 滑动窗口
- `bloomfilter`：布隆过滤器
- `monitoring`：组件指标建模
- `alerting`：阈值告警评估
- `test/performance`、`test/stress`、`test/chaos`：性能/压力/混沌测试

## 运行

```bash
make test
make race
make bench
make perf
make stress
make chaos
```

## 关键约束

- 所有阻塞操作均支持 `context.Context` 取消。
- 错误显式返回 `error`，不使用 `panic`。
- 每个组件至少包含单元测试和基准测试。
