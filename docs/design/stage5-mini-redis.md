# Stage 5 — Cache Engineering: mini-redis

> 预计周期：2 个月 | 语言：Golang | 难度：⭐⭐⭐
>
> 五层掌控：生产使用 → 架构原理 → 故障排查 → 性能优化 → 重写它

---

## 1. 项目目标

从零实现一个 Redis 兼容的内存 KV 存储系统，通过 5 个版本递进式构建，掌握：

- **高性能网络编程：** 单线程事件驱动模型（Event Loop）
- **内存数据结构工程化：** HashMap / SkipList / ZipList
- **持久化策略：** AOF / RDB Snapshot
- **复制与高可用：** 主从复制
- **分布式：** 一致性哈希 + 集群分片

---

## 2. 版本递进设计

```
V0 (TCP + KV)
    │
    ▼
V1 (Pipeline + TTL + 多数据结构)
    │
    ▼
V2 (AOF + RDB Snapshot)
    │
    ▼
V3 (Master-Slave Replication)
    │
    ▼
V4 (Cluster + Consistent Hash)
```

---

## 3. V0 — TCP + KV（基础版）

### 目标
建立最小可用的 Redis-like 服务器。

### 架构

```
┌──────────┐     TCP      ┌──────────────────┐
│  Client  │ ◄──────────► │   mini-redis      │
│  (RESP)  │              │                    │
└──────────┘              │  ┌──────────────┐  │
                          │  │  RESP Parser  │  │
                          │  └──────┬───────┘  │
                          │         │          │
                          │  ┌──────▼───────┐  │
                          │  │ Command       │  │
                          │  │ Dispatcher    │  │
                          │  └──────┬───────┘  │
                          │         │          │
                          │  ┌──────▼───────┐  │
                          │  │  In-Memory   │  │
                          │  │  HashMap     │  │
                          │  └─────────────┘  │
                          └──────────────────┘
```

### 核心组件

```go
// resp.go - RESP 协议解析
type RESPType byte

const (
    SimpleString RESPType = '+'
    Error        RESPType = '-'
    Integer      RESPType = ':'
    BulkString   RESPType = '$'
    Array        RESPType = '*'
)

type RESPValue struct {
    Type     RESPType
    Str      string
    Num      int64
    Array    []RESPValue
}

type RESPParser interface {
    Parse(reader *bufio.Reader) (*RESPValue, error)
    Serialize(value *RESPValue) []byte
}
```

```go
// db.go - 核心数据库
type DB struct {
    data map[string]*Entry
    mu   sync.RWMutex
}

type Entry struct {
    Data     interface{}
    ExpireAt time.Time
}

// 支持的命令
// SET, GET, DEL, EXISTS, KEYS, PING, ECHO
```

### 支持命令
`PING`, `SET`, `GET`, `DEL`, `EXISTS`, `KEYS`

---

## 4. V1 — Pipeline + TTL + 多数据结构

### 新增能力

**Pipeline：** 批量命令处理，减少 RTT。

```go
// pipeline.go
type Pipeline struct {
    commands []*Command
    conn     net.Conn
}

// 客户端一次发送多条命令，服务端批量处理后一次性返回
```

**TTL 机制：**

```go
// ttl.go
type TTLManager struct {
    expireHeap *MinHeap  // 按过期时间排序
    ticker     *time.Ticker
}

// 过期策略
// 1. 惰性删除：访问时检查
// 2. 定期删除：后台协程每 100ms 随机抽样检查
func (t *TTLManager) ActiveExpire() {
    // 每次随机抽 20 个 key，删除已过期的
    // 如果过期比例 > 25%，再抽一轮
}
```

**多数据结构：**

| 数据结构 | 对应 Redis 类型 | 实现 |
|----------|-----------------|------|
| HashMap | String | `map[string][]byte` |
| LinkedList | List | 双向链表 |
| HashSet | Set | `map[string]struct{}` |
| SkipList | Sorted Set | 自实现跳表 |
| HashMap | Hash | `map[string]map[string][]byte` |

```go
// skiplist.go - Sorted Set 底层
type SkipListLevel struct {
    Forward *SkipListNode
    Span    int
}

type SkipListNode struct {
    Key    string
    Score  float64
    Levels []*SkipListLevel
}

type SkipList struct {
    Head   *SkipListNode
    Tail   *SkipListNode
    Length int
    Level  int
}

func (sl *SkipList) Insert(key string, score float64) *SkipListNode
func (sl *SkipList) Delete(key string, score float64) bool
func (sl *SkipList) GetByRank(rank int) *SkipListNode
func (sl *SkipList) RangeByScore(min, max float64) []*SkipListNode
```

### 新增命令
`EXPIRE`, `TTL`, `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LRANGE`, `SADD`, `SMEMBERS`, `ZADD`, `ZRANGE`, `ZRANGEBYSCORE`, `HSET`, `HGET`, `HGETALL`

---

## 5. V2 — AOF + RDB Snapshot

### 持久化架构

```
┌─────────────────────────────────────────┐
│              mini-redis                  │
│                                          │
│  ┌────────────┐    ┌─────────────────┐  │
│  │  In-Memory │    │   AOF Writer    │  │
│  │   Engine   │───►│  (append-only)  │  │
│  └────────────┘    └────────┬────────┘  │
│        │                    │           │
│        │           ┌────────▼────────┐  │
│        │           │   .aof file     │  │
│        │           └─────────────────┘  │
│        │                                │
│        │           ┌─────────────────┐  │
│        └──────────►│  RDB Snapshot   │  │
│                    │  (fork + dump)  │  │
│                    └────────┬────────┘  │
│                             │           │
│                    ┌────────▼────────┐  │
│                    │   .rdb file     │  │
│                    └─────────────────┘  │
└─────────────────────────────────────────┘
```

### AOF 实现

```go
// aof.go
type AOFSyncPolicy int

const (
    AOFAlways    AOFSyncPolicy = 0  // 每条命令 fsync
    AOFEverySec  AOFSyncPolicy = 1  // 每秒 fsync
    AOFNo        AOFSyncPolicy = 2  // OS 决定
)

type AOF struct {
    file       *os.File
    syncPolicy AOFSyncPolicy
    buffer     *bufio.Writer
    mu         sync.Mutex
}

func (a *AOF) Append(command []string) error
func (a *AOF) Replay(db *DB) error         // 启动时回放
func (a *AOF) Rewrite(db *DB) error        // AOF 重写（压缩）
```

### RDB Snapshot

```go
// rdb.go
type RDBEncoder struct {
    writer io.Writer
}

type RDBDecoder struct {
    reader io.Reader
}

// RDB 文件格式（简化）:
// [MAGIC "REDIS0001"] [DB_SELECTOR] [KEY_VALUE_PAIRS...] [EOF] [CHECKSUM]

func (r *RDBEncoder) SaveSnapshot(db *DB) error   // 全量快照
func (r *RDBDecoder) LoadSnapshot() (*DB, error)   // 加载快照
```

**AOF Rewrite 流程：**
1. 开启新协程遍历当前内存数据
2. 将每个 key 的当前状态写为一条命令
3. 期间新写入的命令追加到重写缓冲区
4. 重写完成后，将缓冲区追加到新 AOF 文件
5. 原子替换旧 AOF 文件

---

## 6. V3 — Master-Slave Replication

### 复制架构

```
┌──────────┐   sync    ┌──────────┐
│  Master  │ ────────► │  Slave1  │
│          │           └──────────┘
│          │   sync    ┌──────────┐
│          │ ────────► │  Slave2  │
└──────────┘           └──────────┘
```

### 复制协议

```go
// replication.go
type ReplicationState int

const (
    ReplNone       ReplicationState = 0
    ReplConnecting ReplicationState = 1
    ReplSyncing    ReplicationState = 2  // 全量同步
    ReplOnline     ReplicationState = 3  // 增量同步
)

type ReplicationInfo struct {
    Role           string  // master / slave
    MasterHost     string
    MasterPort     int
    ReplOffset     int64   // 复制偏移量
    ReplBacklog    *RingBuffer  // 复制积压缓冲区
}
```

**同步流程：**
1. **全量同步：** Slave 连接 Master → Master 生成 RDB → 发送给 Slave → Slave 加载
2. **增量同步：** Master 将写命令实时传播给 Slave（通过 Replication Backlog）
3. **断线重连：** 通过 offset 判断是否可增量同步，否则全量

---

## 7. V4 — Cluster + Consistent Hash

### 集群架构

```
┌─────────────────────────────────────────────┐
│                 Client                       │
│           (Smart Client / Proxy)             │
└──────┬──────────┬──────────┬────────────────┘
       │          │          │
┌──────▼───┐ ┌───▼──────┐ ┌▼──────────┐
│  Node 0  │ │  Node 1  │ │  Node 2   │
│ Slot 0-  │ │ Slot     │ │ Slot      │
│   5460   │ │ 5461-    │ │ 10923-    │
│          │ │  10922   │ │  16383    │
└──────────┘ └──────────┘ └───────────┘
```

### 数据分片

```go
// cluster.go
const TotalSlots = 16384

type ClusterNode struct {
    ID        string
    Addr      string
    Slots     []SlotRange
    State     NodeState
}

type SlotRange struct {
    Start uint16
    End   uint16
}

// 一致性哈希 / Slot 映射
func KeyToSlot(key string) uint16 {
    return crc16(key) % TotalSlots
}

type ClusterState struct {
    Nodes     map[string]*ClusterNode
    SlotMap   [TotalSlots]*ClusterNode  // slot → node
    Epoch     uint64
}
```

### MOVED / ASK 重定向

```go
// 当 key 不在当前节点时
// -MOVED <slot> <host>:<port>     → 永久迁移
// -ASK <slot> <host>:<port>       → 临时迁移中

type RedirectError struct {
    Type string  // MOVED / ASK
    Slot uint16
    Addr string
}
```

### Gossip 协议（节点发现）

```go
// gossip.go
type GossipMessage struct {
    SenderID  string
    Type      GossipType  // PING / PONG / MEET / FAIL
    Nodes     []NodeInfo
    Epoch     uint64
}

// 每秒随机选择节点发送 PING
// 通过 PONG 交换节点状态信息
```

---

## 8. 目录结构

```
mini-redis/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── server/
│   │   ├── server.go            # TCP Server + Event Loop
│   │   └── client.go            # 客户端连接管理
│   ├── protocol/
│   │   ├── resp.go              # RESP 协议解析
│   │   └── resp_test.go
│   ├── db/
│   │   ├── db.go                # 核心数据库
│   │   ├── command.go           # 命令注册 & 分发
│   │   ├── string_cmd.go        # String 类型命令
│   │   ├── list_cmd.go          # List 类型命令
│   │   ├── set_cmd.go           # Set 类型命令
│   │   ├── zset_cmd.go          # Sorted Set 命令
│   │   └── hash_cmd.go          # Hash 命令
│   ├── ds/
│   │   ├── dict.go              # 渐进式 rehash HashMap
│   │   ├── skiplist.go          # 跳表
│   │   ├── linkedlist.go        # 双向链表
│   │   └── intset.go            # 整数集合
│   ├── ttl/
│   │   ├── ttl_manager.go       # 过期管理
│   │   └── heap.go              # 最小堆
│   ├── persist/
│   │   ├── aof.go               # AOF 持久化
│   │   ├── aof_rewrite.go       # AOF 重写
│   │   ├── rdb.go               # RDB 快照
│   │   └── rdb_format.go        # RDB 编解码
│   ├── replication/
│   │   ├── master.go            # Master 复制逻辑
│   │   ├── slave.go             # Slave 复制逻辑
│   │   └── backlog.go           # 复制积压缓冲区
│   └── cluster/
│       ├── cluster.go           # 集群状态管理
│       ├── slot.go              # Slot 分配
│       ├── gossip.go            # Gossip 协议
│       └── migrate.go           # Slot 迁移
├── test/
│   ├── v0_basic_test.go
│   ├── v1_pipeline_test.go
│   ├── v2_persist_test.go
│   ├── v3_repl_test.go
│   └── v4_cluster_test.go
├── bench/
│   └── redis_bench_test.go
├── go.mod
├── Makefile
└── README.md
```

---

## 9. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| 缓存雪崩（大量 key 同时过期） | TTL 分布监控 | `redis-cli --scan` + TTL 抽样 |
| 缓存击穿（查不存在的 key） | 请求日志 + 空值率统计 | 布隆过滤器 / 空值缓存 |
| 缓存穿透（热 key 过期） | QPS 突变监控 | `redis-cli --hotkeys` / 互斥锁重建 |
| 大 Key 阻塞 | SLOWLOG + 内存分析 | `redis-cli --bigkeys` / `MEMORY USAGE` |
| 内存泄漏（过期策略失效） | INFO memory 监控 | `INFO memory` + 定期内存采样 |
| 主从复制延迟 | offset 差异监控 | `INFO replication` / `replBacklogSize` |

---

## 10. 面试考点

### 必问级

- [ ] **Redis 为什么快：** 单线程 + IO 多路复用 + 纯内存 + 高效数据结构
- [ ] **缓存雪崩/击穿/穿透：** 区别是什么？各自解决方案？
- [ ] **持久化：** AOF vs RDB 优缺点？AOF rewrite 流程？
- [ ] **数据结构编码：** String 用 SDS、Sorted Set 用 SkipList + HashTable 的原因

### 高频级

- [ ] **主从复制：** 全量 + 增量同步流程？断线重连如何处理？
- [ ] **Cluster 分片：** 16384 slot + CRC16？MOVED/ASK 重定向？
- [ ] **内存淘汰策略：** allkeys-lru / volatile-lru / noeviction 区别
- [ ] **大 Key 问题：** 如何发现？如何处理（拆分/异步删除）？

---

## 11. 实现里程碑

### M1: V0 — TCP + KV（第1-2周）
- [ ] RESP 协议解析器
- [ ] TCP Server（goroutine-per-connection）
- [ ] HashMap 存储 + 基础命令
- [ ] 可用 `redis-cli` 连接测试

### M2: V1 — Pipeline + TTL + 数据结构（第3-4周）
- [ ] Pipeline 批处理
- [ ] TTL 惰性删除 + 定期删除
- [ ] SkipList（Sorted Set）
- [ ] 双向链表（List）
- [ ] 基准测试：SET/GET QPS

### M3: V2 — 持久化（第5周）
- [ ] AOF append + replay
- [ ] AOF rewrite
- [ ] RDB snapshot + load
- [ ] 测试：kill 后重启数据不丢失

### M4: V3 — 主从复制（第6-7周）
- [ ] SLAVEOF 命令
- [ ] 全量同步（RDB 传输）
- [ ] 增量同步（Backlog）
- [ ] 测试：Master 写入，Slave 可读

### M5: V4 — 集群（第8周）
- [ ] Slot 分配 + CRC16 哈希
- [ ] MOVED / ASK 重定向
- [ ] Gossip 节点发现
- [ ] 测试：3 节点集群读写

---

## 10. 性能目标

| 指标 | 目标值 |
|------|--------|
| SET QPS（单连接） | > 50,000 |
| GET QPS（单连接） | > 80,000 |
| Pipeline 1000 命令 | < 10ms |
| AOF Rewrite 100w key | < 5s |
| 主从同步延迟 | < 100ms |

---

## 11. 毕业标准

- [ ] 可用 `redis-cli` 正常交互
- [ ] 能解释 Redis 单线程为什么快
- [ ] TTL 过期策略正确，内存不泄漏
- [ ] AOF + RDB 双持久化可靠
- [ ] 能定位缓存雪崩 / 击穿 / 穿透问题
- [ ] 完成压测报告
- [ ] 3 节点集群可正常工作
