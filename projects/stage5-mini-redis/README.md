# Stage 5 — mini-redis

> 对应设计文档：`docs/design/stage5-mini-redis.md`

## 已实现能力（V0~V4 精简版）

- V0：TCP + RESP + 基础 KV 命令
- V1：Pipeline、TTL（惰性+定期删除）、List/Set/ZSet/Hash
- V2：AOF append/replay + RDB snapshot/load
- V3：Master-Slave 复制骨架 + backlog
- V4：Cluster slot + CRC16 + MOVED 重定向模型
- 监控与告警模块
- 单元测试、基准测试、性能/压力/混沌测试

## 运行

```bash
make test
make race
make bench
make run
```

默认地址：`127.0.0.1:16379`

## 文档

- 排障手册：`TROUBLESHOOTING.md`
