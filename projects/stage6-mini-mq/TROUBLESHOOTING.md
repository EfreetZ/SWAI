# Stage 6 排障手册

## 1. Consumer Lag 持续增长

### 定位
1. 检查 group offset 与 log end offset 差值。
2. 观察消费者处理耗时与批量大小。

### 处理
- 增加 partition 数。
- 提升 consumer 并发或减少单批处理耗时。

## 2. Broker 磁盘增长过快

### 定位
1. 检查 retention 配置与 segment rolling 频率。
2. 观察 topic 写入速率。

### 处理
- 调小 retention 时间/大小。
- 定期触发 log cleaner。

## 3. Rebalance 频繁触发

### 定位
1. 检查心跳超时与消费者抖动。
2. 分析 group 变更日志。

### 处理
- 拉长心跳超时阈值。
- 使用稳定分配策略并减少频繁扩缩容。

## 4. 消息重复消费

### 定位
1. 检查 offset 提交时机。
2. 检查是否在处理失败后仍提交 offset。

### 处理
- 仅在处理成功后提交 offset。
- 对消费逻辑增加幂等键校验。
