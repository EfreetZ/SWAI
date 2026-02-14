# Stage 1 排障手册

## 1. Worker Pool 任务堆积

### 现象
- 提交任务变慢，调用方超时。

### 定位方式
1. 观察任务提交错误是否为 `context deadline exceeded`。
2. 检查 worker 数量与队列长度配置。
3. 使用 `go test -run TestWorkerPoolSubmitAndClose -v ./concurrency/patterns` 做快速回归。

### 解决建议
- 提高 worker 数量。
- 将慢任务拆分并确保支持 `ctx.Done()`。
- 对调用方增加重试退避策略。

## 2. LRU 命中率下降

### 现象
- 缓存频繁淘汰，后端负载升高。

### 定位方式
1. 检查容量是否过小。
2. 通过 `test/performance` 验证当前容量下的吞吐表现。

### 解决建议
- 根据热点分布调整容量。
- 结合业务按 key 前缀分片缓存。

## 3. TCP Echo 服务连接异常

### 现象
- 客户端偶发连接重置或读写超时。

### 定位方式
1. 检查是否提前取消了 `context`。
2. 使用 `nc` 手工验证：`echo "hello" | nc 127.0.0.1 <port>`。
3. 运行 `go test ./network/tcp -v` 验证回归。

### 解决建议
- 调整客户端超时。
- 优化服务端读写超时策略。

## 4. 告警误报

### 现象
- goroutine 或 heap 告警频繁触发。

### 定位方式
1. 检查阈值是否与当前环境匹配。
2. 对比 `monitoring.CollectRuntimeMetrics` 的实时结果。

### 解决建议
- 分环境设置阈值（开发/测试/生产）。
- 增加采样窗口，避免瞬时抖动触发告警。
