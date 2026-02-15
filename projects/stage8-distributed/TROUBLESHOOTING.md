# Stage 8 排障

## Leader 频繁切换
- 检查 election timeout 是否过小。
- 检查节点心跳是否稳定。

## 日志复制失败
- 检查 prevLogIndex/prevLogTerm 是否匹配。
- 检查 follower 日志截断逻辑。

## 锁误释放
- 检查锁 TTL 与续约策略。
- 检查 owner 校验是否严格。
