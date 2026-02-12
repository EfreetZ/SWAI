# Stage 6 — Log-based System: mini-mq

> 预计周期：2 个月 | 语言：Golang | 难度：⭐⭐⭐⭐
>
> 五层掌控：生产使用 → 架构原理 → 故障排查 → 性能优化 → 重写它

---

## 1. 项目目标

从零实现一个 Kafka 风格的分布式消息队列，深入理解：

- **日志即数据：** 以 append-only log 作为核心抽象
- **顺序写的吞吐优势：** 为什么 Kafka 可以达到百万级 TPS
- **分区与并行消费：** Partition 模型如何实现水平扩展
- **消费偏移管理：** Consumer Group 如何保证 at-least-once / exactly-once
- **批量与零拷贝：** 高性能 IO 的工程实践

---

## 2. 整体架构

```
                    ┌──────────────┐
                    │   Producer   │
                    │   (batch)    │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │   Broker     │
                    │              │
                    │ ┌──────────┐ │
                    │ │  Topic   │ │
                    │ │ ┌──────┐ │ │
                    │ │ │ P0   │ │ │     P = Partition
                    │ │ │ P1   │ │ │
                    │ │ │ P2   │ │ │
                    │ │ └──────┘ │ │
                    │ └──────────┘ │
                    │              │
                    │ ┌──────────┐ │
                    │ │ Offset   │ │
                    │ │ Manager  │ │
                    │ └──────────┘ │
                    └──────┬───────┘
                           │
               ┌───────────┼───────────┐
               │           │           │
        ┌──────▼──┐ ┌──────▼──┐ ┌─────▼───┐
        │Consumer │ │Consumer │ │Consumer │
        │  (C0)   │ │  (C1)   │ │  (C2)   │
        └─────────┘ └─────────┘ └─────────┘
                Consumer Group
```

---

## 3. 核心概念模型

```go
// model.go
type Topic struct {
    Name       string
    Partitions []*Partition
    Config     TopicConfig
}

type TopicConfig struct {
    NumPartitions     int
    RetentionBytes    int64
    RetentionMs       int64
    SegmentBytes      int64         // 单个 Segment 文件大小上限
    FlushIntervalMs   int64
}

type Partition struct {
    TopicName    string
    ID           int
    Dir          string            // 数据目录
    ActiveSeg    *Segment          // 当前写入的 Segment
    Segments     []*Segment        // 所有 Segment（有序）
    mu           sync.RWMutex
}

type Message struct {
    Offset    int64
    Key       []byte
    Value     []byte
    Timestamp int64
    Headers   map[string]string
}
```

---

## 4. 存储引擎设计

### 4.1 Segment 文件

每个 Partition 由多个 Segment 文件组成，每个 Segment 包含：

```
partition-0/
├── 00000000000000000000.log      # 消息日志文件
├── 00000000000000000000.index    # 偏移量索引
├── 00000000000000000000.timeindex # 时间索引
├── 00000000000005242880.log
├── 00000000000005242880.index
└── 00000000000005242880.timeindex
```

文件名 = 该 Segment 第一条消息的 Offset。

### 4.2 Log 文件格式

```
┌─────────────────────────────────────────┐
│  Record 0                                │
│  ┌────────┬────────┬───────┬──────────┐ │
│  │ Length  │ CRC32  │ Key   │  Value   │ │
│  │ 4 bytes│ 4 bytes│ var   │  var     │ │
│  └────────┴────────┴───────┴──────────┘ │
├─────────────────────────────────────────┤
│  Record 1                                │
│  ...                                     │
├─────────────────────────────────────────┤
│  Record N                                │
└─────────────────────────────────────────┘
```

```go
// segment.go
type Segment struct {
    BaseOffset  int64
    NextOffset  int64
    LogFile     *os.File
    IndexFile   *os.File
    LogSize     int64
    MaxBytes    int64
    mu          sync.Mutex
}

type IndexEntry struct {
    RelativeOffset uint32    // 相对于 BaseOffset 的偏移
    Position       uint32    // 在 .log 文件中的物理位置
}

func (s *Segment) Append(msg *Message) (int64, error)
func (s *Segment) Read(offset int64) (*Message, error)
func (s *Segment) Close() error
func (s *Segment) IsFull() bool
```

### 4.3 稀疏索引

不为每条消息建索引，每隔 N 条（如 4KB 数据量）建一个索引条目：

```
Index File:
┌────────────────────┬────────────────────┐
│ Relative Offset: 0 │ Position: 0        │
│ Relative Offset: 50│ Position: 4096     │
│ Relative Offset:100│ Position: 8192     │
└────────────────────┴────────────────────┘

查找 Offset=75:
1. 二分查找 Index → 找到 [50, 4096]
2. 从 Position=4096 开始顺序扫描直到 Offset=75
```

```go
// index.go
type SparseIndex struct {
    entries  []IndexEntry
    file     *os.File
    mmap     []byte           // mmap 映射，加速查找
}

func (idx *SparseIndex) Lookup(offset int64) (position int64, err error)
func (idx *SparseIndex) Append(offset int64, position int64) error
```

---

## 5. Producer 设计

```go
// producer.go
type Producer struct {
    brokerAddr  string
    conn        net.Conn
    batchSize   int
    lingerMs    int           // 攒批等待时间
    buffer      []*Message
    mu          sync.Mutex
}

type ProducerConfig struct {
    BatchSize     int       // 攒够多少条发送
    LingerMs      int       // 最大等待毫秒
    Acks          int       // 0: 不等确认, 1: Leader确认, -1: 所有副本确认
    Compression   string    // none / snappy / gzip
    Partitioner   Partitioner
}

type Partitioner interface {
    Partition(key []byte, numPartitions int) int
}

// 默认分区策略
type RoundRobinPartitioner struct { counter int64 }
type KeyHashPartitioner struct{}

func (p *Producer) Send(topic string, key, value []byte) error
func (p *Producer) sendBatch(messages []*Message) error
func (p *Producer) Close() error
```

**批量发送流程：**
1. Producer 将消息放入本地 buffer
2. 当 buffer 达到 `BatchSize` 或等待超过 `LingerMs`
3. 将同一 Partition 的消息打包发送
4. 等待 Broker 确认

---

## 6. Consumer 设计

### 6.1 基础 Consumer

```go
// consumer.go
type Consumer struct {
    brokerAddr    string
    groupID       string
    topics        []string
    partitions    map[string][]int   // topic → assigned partitions
    offsets       map[string]int64   // partition → committed offset
    conn          net.Conn
}

type ConsumerConfig struct {
    GroupID          string
    AutoCommit       bool
    AutoCommitMs     int
    FetchMinBytes    int
    FetchMaxWaitMs   int
    MaxPollRecords   int
}

func (c *Consumer) Poll(timeout time.Duration) ([]*Message, error)
func (c *Consumer) Commit() error
func (c *Consumer) CommitOffset(topic string, partition int, offset int64) error
func (c *Consumer) Close() error
```

### 6.2 Consumer Group + Rebalance

```go
// consumer_group.go
type ConsumerGroup struct {
    GroupID      string
    Members      map[string]*ConsumerMember
    Generation   int
    Coordinator  *GroupCoordinator
}

type ConsumerMember struct {
    MemberID    string
    Partitions  []TopicPartition
    LastHeartbeat time.Time
}

type GroupCoordinator struct {
    groups map[string]*ConsumerGroup
}

// Rebalance 策略
type AssignmentStrategy interface {
    Assign(members []string, partitions []TopicPartition) map[string][]TopicPartition
}

// Range 分配：按分区范围均分
type RangeAssignor struct{}
// RoundRobin 分配：轮询分配
type RoundRobinAssignor struct{}
```

**Rebalance 触发条件：**
- Consumer 加入 / 离开 Group
- Consumer 心跳超时
- Topic 分区数变更

### 6.3 Offset 管理

```go
// offset.go
type OffsetManager struct {
    store   OffsetStore
}

type OffsetStore interface {
    Get(group, topic string, partition int) (int64, error)
    Set(group, topic string, partition int, offset int64) error
}

// V1: 基于文件的 Offset 存储
type FileOffsetStore struct {
    dir string
}

// V2: 基于内部 Topic（__consumer_offsets）的存储
type TopicOffsetStore struct {
    topic *Topic
}
```

---

## 7. 网络协议

```go
// protocol.go
type RequestType uint16

const (
    ProduceRequest      RequestType = 0
    FetchRequest        RequestType = 1
    OffsetCommitRequest RequestType = 2
    OffsetFetchRequest  RequestType = 3
    MetadataRequest     RequestType = 4
    JoinGroupRequest    RequestType = 5
    HeartbeatRequest    RequestType = 6
)

type Request struct {
    APIKey      RequestType
    CorrelationID int32
    ClientID    string
    Body        []byte
}

type Response struct {
    CorrelationID int32
    ErrorCode     int16
    Body          []byte
}

// 编码格式：Length-Prefixed Binary
// [4 bytes: total length] [2 bytes: API key] [4 bytes: correlation ID] [...body]
```

---

## 8. Broker 核心

```go
// broker.go
type Broker struct {
    ID          int
    Addr        string
    Topics      map[string]*Topic
    listener    net.Listener
    offsetMgr   *OffsetManager
    groupCoord  *GroupCoordinator
}

func (b *Broker) Start() error
func (b *Broker) handleConnection(conn net.Conn)
func (b *Broker) handleProduce(req *ProduceRequest) *ProduceResponse
func (b *Broker) handleFetch(req *FetchRequest) *FetchResponse
func (b *Broker) CreateTopic(name string, config TopicConfig) error
```

---

## 9. 高性能关键设计

### 9.1 顺序写
- 消息追加到 Segment 尾部，利用磁盘顺序写的高吞吐
- 典型 HDD 顺序写: ~100MB/s, 随机写: ~1MB/s

### 9.2 零拷贝（sendfile）
```go
// Consumer 拉取消息时，直接从磁盘文件发送到网络 socket
// 避免 Kernel → User → Kernel 的数据拷贝
func sendfileToConsumer(conn net.Conn, file *os.File, offset int64, count int64) error {
    // 使用 syscall.Sendfile
    dst := conn.(*net.TCPConn)
    // ...
}
```

### 9.3 mmap 索引
```go
// 将 .index 文件通过 mmap 映射到内存
// 实现 O(1) 内存访问 + 无需手动管理缓存
func mmapIndex(file *os.File) ([]byte, error) {
    info, _ := file.Stat()
    return syscall.Mmap(int(file.Fd()), 0, int(info.Size()),
        syscall.PROT_READ, syscall.MAP_SHARED)
}
```

### 9.4 批量处理
- Producer 攒批发送，减少网络 RTT
- Broker 批量写入，减少 fsync 次数
- Consumer 批量拉取，提高吞吐

---

## 10. 目录结构

```
mini-mq/
├── cmd/
│   ├── broker/
│   │   └── main.go              # Broker 启动入口
│   ├── producer/
│   │   └── main.go              # Producer CLI 示例
│   └── consumer/
│       └── main.go              # Consumer CLI 示例
├── internal/
│   ├── broker/
│   │   ├── broker.go            # Broker 核心逻辑
│   │   ├── topic.go             # Topic 管理
│   │   └── config.go            # 配置
│   ├── storage/
│   │   ├── partition.go         # Partition 管理
│   │   ├── segment.go           # Segment 读写
│   │   ├── index.go             # 稀疏索引
│   │   ├── timeindex.go         # 时间索引
│   │   └── log_cleaner.go       # 日志清理（retention）
│   ├── protocol/
│   │   ├── request.go           # 请求定义
│   │   ├── response.go          # 响应定义
│   │   ├── codec.go             # 编解码
│   │   └── errors.go            # 错误码
│   ├── producer/
│   │   ├── producer.go          # Producer 客户端
│   │   ├── partitioner.go       # 分区策略
│   │   └── batch.go             # 批量发送
│   ├── consumer/
│   │   ├── consumer.go          # Consumer 客户端
│   │   ├── group.go             # Consumer Group
│   │   ├── coordinator.go       # Group Coordinator
│   │   ├── rebalance.go         # Rebalance 策略
│   │   └── offset.go            # Offset 管理
│   └── server/
│       ├── server.go            # TCP Server
│       └── handler.go           # 请求路由
├── test/
│   ├── segment_test.go
│   ├── index_test.go
│   ├── producer_test.go
│   ├── consumer_test.go
│   └── integration_test.go
├── bench/
│   ├── write_bench_test.go      # 写入吞吐基准测试
│   └── read_bench_test.go       # 读取吞吐基准测试
├── go.mod
├── Makefile
└── README.md
```

---

## 11. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| Consumer Lag 飙高 | 消费偏移监控 | `kafka-consumer-groups.sh --describe` |
| 磁盘写满 | 磁盘空间 + retention 配置 | `df -h` / `du -sh partition-*` |
| Partition 数据倾斜 | 各 Partition 流量统计 | `kafka-log-dirs.sh --describe` |
| Producer 发送失败 | 错误日志 + ack 超时 | Producer metrics / 网络报文 |
| Rebalance 风暴 | Consumer Group 状态变化 | Consumer 日志 / heartbeat 监控 |
| 消息丢失 | offset 提交策略审查 | auto.commit vs manual commit 分析 |

---

## 12. 面试考点

### 必问级

- [ ] **Kafka 为什么快：** 顺序写 + mmap + 零拷贝 + 批量发送
- [ ] **消息语义：** at-most-once / at-least-once / exactly-once 的区别和实现
- [ ] **Partition 模型：** 为什么分区？如何保证分区内有序？
- [ ] **Consumer Group：** Rebalance 触发条件？分配策略（Range / RoundRobin）？

### 高频级

- [ ] **Segment + Index：** 稀疏索引如何工作？为什么不给每条消息建索引？
- [ ] **零拷贝：** sendfile 的原理？数据流路径（少了哪几次拷贝）？
- [ ] **ISR 机制：** 在 Kafka 中如何保证数据不丢失？
- [ ] **积压处理：** Consumer Lag 持续增长怎么办？

---

## 13. 实现里程碑

### M1: 单 Partition 读写（第1-2周）
- [ ] 实现 Segment 文件追加写入
- [ ] 实现消息编码 / 解码（Length + CRC + Key + Value）
- [ ] 实现稀疏索引 + 二分查找
- [ ] 单元测试：100w 条消息写入 & 读取

### M2: 多 Partition + Topic（第3周）
- [ ] 实现 Topic → Partition 映射
- [ ] Partition 路由（Key Hash / Round Robin）
- [ ] Segment Rolling（文件满时滚动创建新 Segment）
- [ ] 日志清理（基于时间 / 大小的 retention）

### M3: 网络层 + Producer（第4-5周）
- [ ] 实现 Broker TCP Server
- [ ] 实现请求 / 响应协议编解码
- [ ] 实现 Producer 客户端 + 批量发送
- [ ] 基准测试：单 Broker 写入 QPS

### M4: Consumer + Offset（第6周）
- [ ] 实现 Consumer 拉取消息
- [ ] 实现 Offset 提交与存储
- [ ] 支持 from-beginning / from-latest / from-offset

### M5: Consumer Group + Rebalance（第7周）
- [ ] 实现 Group Coordinator
- [ ] 实现 Join / Sync / Heartbeat 协议
- [ ] 实现 Range / RoundRobin 分区分配
- [ ] 测试：3 Consumer 消费 6 Partition

### M6: 性能优化 + 压测（第8周）
- [ ] mmap 索引
- [ ] sendfile 零拷贝（如平台支持）
- [ ] 批量 fsync 策略
- [ ] 端到端压测报告

---

## 12. 性能目标

| 指标 | 目标值 |
|------|--------|
| 单 Partition 写入 QPS | > 100,000 msg/s |
| 单 Partition 写入吞吐 | > 50 MB/s |
| 消费吞吐 | > 80 MB/s |
| 端到端延迟（P99） | < 10ms |
| 100w 消息写入时间 | < 10s |

---

## 13. 关键学习产出

- **日志系统思维：** 理解 "log is the database" 这一核心理念
- **顺序写优势：** 为什么 Kafka 能在廉价硬件上实现高吞吐
- **Consumer Group：** 如何实现消费端的水平扩展
- **Offset 语义：** at-most-once / at-least-once / exactly-once 的区别
- **零拷贝：** sendfile / mmap 在高吞吐场景的价值

---

## 14. 毕业标准

- [ ] 理解顺序写 vs 随机写的性能差异（能给出数据）
- [ ] 能估算给定消息大小下的理论吞吐上限
- [ ] Consumer Group 能正确处理 Rebalance
- [ ] 做过 batch 写入 / 消费的压测
- [ ] Segment 清理策略正确，磁盘不无限增长
- [ ] 能画出消息从 Producer 到 Consumer 的完整数据流路径
