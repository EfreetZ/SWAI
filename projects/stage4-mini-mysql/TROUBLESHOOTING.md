# Stage 4 排障手册

## 1. 写入后重启数据丢失

### 定位
1. 检查 `.wal` 文件是否持续增长。
2. 检查 COMMIT 是否落盘（`wal.Flush` 是否执行）。
3. 运行 `go test ./internal/wal -v` 验证回放逻辑。

### 处理
- 在事务提交路径强制 `Flush`。
- 进程退出前执行 `Checkpoint`。

## 2. 范围查询慢

### 定位
1. 检查是否命中范围扫描迭代器。
2. 对比 `bench/storage_bench_test.go` 基准结果。

### 处理
- 优化节点扇出（order）。
- 降低频繁页换入换出。

## 3. Buffer Pool 命中率低

### 定位
1. 观察 `monitoring.StorageMetrics.BufferHitRatio()`。
2. 检查热点页是否频繁被淘汰。

### 处理
- 提升 buffer capacity。
- 调整 LRU 策略（预留热点页）。

## 4. 并发事务冲突

### 定位
1. 检查锁等待时间与冲突 key。
2. 使用 `go test ./test/stress -v` 复现并发写冲突。

### 处理
- 缩短事务持锁时间。
- 将大事务拆分为小事务。
