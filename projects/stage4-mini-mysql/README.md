# Stage 4 — mini-mysql

> 对应设计文档：`docs/design/stage4-mini-mysql.md`

## 已实现模块

- Page Manager（固定页读写）
- B+Tree（简化实现，支持 Insert/Search/Delete/Range）
- WAL（追加日志 + Flush + 简化恢复）
- Buffer Pool（LRU 置换）
- 事务管理（BEGIN/COMMIT/ROLLBACK）
- Mini SQL Parser（CREATE TABLE / INSERT / SELECT / TX）
- Executor + TCP 文本协议服务
- 监控指标与告警评估
- 单元测试、基准测试、性能/压力/混沌测试

## 运行

```bash
make test
make race
make bench
make run
```

默认 TCP 地址：`127.0.0.1:13306`

## Docker

```bash
make docker-build
make docker-run
```

## 文档

- 排障手册：`TROUBLESHOOTING.md`
