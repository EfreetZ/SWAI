# Stage 8 — mini-raft & distributed coordination

实现内容（简化版）：
- Raft 核心状态机（选主/日志复制接口）
- 上层 KV（Put/Get/Delete）
- 分布式锁（TTL）
- 雪花 ID 生成器
- 服务注册发现（内存）

```bash
make test
make race
make bench
```
