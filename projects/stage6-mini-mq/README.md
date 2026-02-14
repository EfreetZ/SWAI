# Stage 6 — mini-mq

> 对应设计文档：`docs/design/stage6-mini-mq.md`

## 已实现能力（精简版）

- Segment 文件 + 稀疏索引
- Topic + Partition 管理
- Producer 批量发送 + 分区策略
- Consumer 拉取 + Offset 管理
- Consumer Group + Rebalance（Range / RoundRobin）
- Broker TCP 协议处理
- 监控告警模块
- 单元测试、基准测试、性能/压力/混沌测试

## 运行

```bash
make test
make race
make bench
make run
```

默认 broker 地址：`127.0.0.1:19092`

## 文档

- 排障手册：`TROUBLESHOOTING.md`
