# Stage 4 — Storage Foundation: mini-mysql

> 预计周期：2 个月 | 语言：Golang | 难度：⭐⭐⭐
>
> 五层掌控：生产使用 → 架构原理 → 故障排查 → 性能优化 → 重写它

---

## 1. 项目目标

从零实现一个简化版关系型数据库引擎，掌握存储系统的核心原理：

- 理解 **磁盘 IO** 如何成为数据库的瓶颈
- 掌握 **B+Tree** 作为索引结构的工程实现
- 理解 **WAL（Write-Ahead Logging）** 保证数据持久化的机制
- 构建 **Buffer Pool** 减少磁盘访问
- 实现简单事务（BEGIN / COMMIT / ROLLBACK）

**不是目标：** 实现完整 SQL 解析器、优化器、完整 MVCC。聚焦存储引擎层。

---

## 2. 整体架构

```
┌─────────────────────────────────────────────┐
│                  Client                     │
│            (TCP Connection)                 │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              Protocol Layer                 │
│         (简化 MySQL 协议 / 自定义)           │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            SQL Parser (简化)                │
│     支持: SELECT / INSERT / CREATE TABLE    │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            Execution Engine                 │
│         (Plan → Operator → Result)          │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│            Storage Engine                   │
│  ┌─────────┐ ┌─────────┐ ┌──────────────┐  │
│  │ B+Tree  │ │   WAL   │ │ Buffer Pool  │  │
│  │  Index  │ │  Logger │ │   Manager    │  │
│  └─────────┘ └─────────┘ └──────────────┘  │
│  ┌─────────────────────────────────────┐    │
│  │         Page Manager                │    │
│  │     (磁盘页读写 / 空间管理)          │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
                   │
            ┌──────▼──────┐
            │  Disk Files │
            │  .db / .wal │
            └─────────────┘
```

---

## 3. 核心模块设计

### 3.1 Page Manager（页管理器）

**职责：** 管理磁盘上的固定大小页（4KB / 8KB / 16KB）。

```go
// page.go
const PageSize = 4096

type PageID uint32

type Page struct {
    ID       PageID
    Data     [PageSize]byte
    Dirty    bool
    PinCount int
}

type PageManager interface {
    ReadPage(id PageID) (*Page, error)
    WritePage(page *Page) error
    AllocatePage() (PageID, error)
    FreePage(id PageID) error
}
```

**关键设计决策：**
- 页大小选择 4KB，与操作系统页对齐
- 页内部格式：Header (24B) + Data + Slot Array (从尾部增长)
- 页类型：数据页、索引页、溢出页

### 3.2 B+Tree Index

**职责：** 提供有序数据访问能力，支持点查与范围查询。

```go
// btree.go
type BPlusTree struct {
    rootPageID PageID
    order      int
    pager      PageManager
}

type NodeType uint8

const (
    LeafNode     NodeType = 0
    InternalNode NodeType = 1
)

// 节点页内布局
type BTreeNode struct {
    Type     NodeType
    NumKeys  uint16
    Keys     [][]byte
    Values   [][]byte    // 叶子节点: 实际数据; 内部节点: 子页ID
    NextLeaf PageID      // 叶子节点链表指针
}

type BPlusTreeAPI interface {
    Insert(key, value []byte) error
    Delete(key []byte) error
    Search(key []byte) ([]byte, error)
    RangeScan(startKey, endKey []byte) (Iterator, error)
}
```

**实现要点：**
- 节点分裂（Split）与合并（Merge）
- 叶子节点通过 NextLeaf 形成链表，支持范围扫描
- Order 计算：`(PageSize - HeaderSize) / (KeySize + ValueSize + PointerSize)`

### 3.3 WAL（Write-Ahead Logging）

**职责：** 在数据页写入磁盘前，先将变更记录写入日志，保证 crash recovery。

```go
// wal.go
type LSN uint64 // Log Sequence Number

type LogRecord struct {
    LSN       LSN
    TxID      uint64
    Type      LogType    // INSERT / UPDATE / DELETE / BEGIN / COMMIT / ABORT
    PageID    PageID
    Offset    uint16
    OldValue  []byte     // UNDO
    NewValue  []byte     // REDO
}

type WAL interface {
    Append(record *LogRecord) (LSN, error)
    Flush() error
    ReadFrom(lsn LSN) ([]*LogRecord, error)
    Checkpoint() error
}
```

**恢复策略（ARIES 简化版）：**
1. **Redo Phase：** 从最近 Checkpoint 开始，重放所有已提交事务的日志
2. **Undo Phase：** 回滚所有未提交事务的操作

### 3.4 Buffer Pool Manager

**职责：** 缓存磁盘页到内存，减少 IO。

```go
// buffer_pool.go
type BufferPoolManager struct {
    pages    map[PageID]*Page
    replacer Replacer         // LRU / Clock
    freeList []FrameID
    disk     PageManager
    capacity int
}

type BufferPoolAPI interface {
    FetchPage(id PageID) (*Page, error)
    UnpinPage(id PageID, isDirty bool) error
    FlushPage(id PageID) error
    NewPage() (*Page, error)
    FlushAllPages() error
}

// 页面置换策略
type Replacer interface {
    Victim() (FrameID, bool)   // 选择被淘汰的页
    Pin(frameID FrameID)       // 页被访问，不可淘汰
    Unpin(frameID FrameID)     // 页可被淘汰
    Size() int
}
```

**置换算法：** 先实现 LRU，后续可扩展为 LRU-K 或 Clock。

### 3.5 Transaction Manager（简化事务）

```go
// transaction.go
type TxState uint8

const (
    TxActive    TxState = 0
    TxCommitted TxState = 1
    TxAborted   TxState = 2
)

type Transaction struct {
    TxID      uint64
    State     TxState
    WriteSets []WriteRecord
}

type TxManager interface {
    Begin() *Transaction
    Commit(tx *Transaction) error
    Abort(tx *Transaction) error
}
```

**简化范围：**
- 仅支持单表事务
- 使用页级锁（Page-level Lock）
- 不实现 MVCC（留作后续扩展）

---

## 4. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| 慢查询（全表扫描） | EXPLAIN 执行计划分析 | `EXPLAIN SELECT ...` 查看 type=ALL |
| 死锁 | 锁等待日志 + 超时检测 | `SHOW ENGINE INNODB STATUS` |
| Buffer Pool 命中率低 | pprof + IO 监控 | `iostat -xz 1` / pprof heap |
| WAL 写满磁盘 | 磁盘空间监控 | `df -h` / `du -sh *.wal` |
| B+Tree 分裂导致写入抖动 | 写入延迟 P99 监控 | pprof CPU 火焰图 |
| 并发事务冲突 | 锁等待队列长度 | 自定义 metrics + dlv 调试 |

---

## 5. 面试考点

### 必问级

- [ ] **B+Tree vs B-Tree vs Hash Index：** 为什么数据库选 B+Tree？（磁盘友好、范围查询、叶子链表）
- [ ] **WAL 原理：** 为什么先写日志再写数据页？crash 后如何恢复（Redo + Undo）？
- [ ] **Buffer Pool：** LRU 淘汰策略的问题？MySQL 为什么用 LRU-K？
- [ ] **MVCC：** 快照读 vs 当前读？undo log 的作用？

### 高频级

- [ ] **索引失效场景：** 最左前缀、函数包裹、类型转换、LIKE '%xxx'
- [ ] **聚簇索引 vs 非聚簇索引：** 回表是什么？覆盖索引如何避免回表？
- [ ] **事务隔离级别：** RC vs RR 的区别？幻读如何解决？
- [ ] **死锁检测：** 等待图（wait-for graph）如何工作？

---

## 6. 实现里程碑

### M1: Page Manager + Disk IO（第1周）

- [ ] 实现固定大小页的读写
- [ ] 实现页分配器（Free List / Bitmap）
- [ ] 单元测试：并发读写页

### M2: B+Tree（第2-3周）

- [ ] 实现内存版 B+Tree（Insert / Search / Delete）
- [ ] 实现 Split 和 Merge
- [ ] 将 B+Tree 持久化到磁盘页
- [ ] 实现 RangeScan + Iterator
- [ ] 基准测试：10w 条数据插入 & 查询

### M3: WAL（第4周）

- [ ] 实现日志追加写入
- [ ] 实现 LSN 管理
- [ ] 实现 Checkpoint
- [ ] 实现 Crash Recovery（Redo + Undo）
- [ ] 测试：写入后 kill 进程，重启验证数据完整性

### M4: Buffer Pool（第5周）

- [ ] 实现 LRU Replacer
- [ ] 实现 Buffer Pool Manager
- [ ] 集成到 B+Tree，所有页访问经过 Buffer Pool
- [ ] 基准测试：对比有无 Buffer Pool 的 IO 次数

### M5: Transaction + SQL 简化层（第6-7周）

- [ ] 实现 BEGIN / COMMIT / ROLLBACK
- [ ] 页级锁
- [ ] 简化 SQL 解析器（正则 / 手写递归下降）
- [ ] 支持：`CREATE TABLE`, `INSERT`, `SELECT ... WHERE`

### M6: TCP Server + 压测（第8周）

- [ ] TCP 监听 + 连接管理
- [ ] 支持简化 MySQL 协议或自定义文本协议
- [ ] 使用 sysbench / 自定义客户端进行压测
- [ ] 输出性能报告：QPS / 延迟 P99

---

## 5. 目录结构

```
mini-mysql/
├── cmd/
│   └── server/
│       └── main.go              # 入口
├── internal/
│   ├── storage/
│   │   ├── page.go              # 页定义 & Page Manager
│   │   ├── btree.go             # B+Tree 实现
│   │   ├── btree_node.go        # B+Tree 节点操作
│   │   └── iterator.go          # 范围扫描迭代器
│   ├── buffer/
│   │   ├── buffer_pool.go       # Buffer Pool Manager
│   │   ├── lru_replacer.go      # LRU 置换策略
│   │   └── clock_replacer.go    # Clock 置换策略（扩展）
│   ├── wal/
│   │   ├── log_record.go        # 日志记录定义
│   │   ├── wal.go               # WAL 核心实现
│   │   └── recovery.go          # Crash Recovery
│   ├── txn/
│   │   ├── transaction.go       # 事务定义
│   │   ├── tx_manager.go        # 事务管理器
│   │   └── lock_manager.go      # 锁管理
│   ├── parser/
│   │   ├── lexer.go             # 词法分析
│   │   ├── parser.go            # 语法分析
│   │   └── ast.go               # AST 节点定义
│   ├── executor/
│   │   ├── executor.go          # 执行引擎
│   │   └── operators.go         # 扫描 / 过滤算子
│   └── server/
│       ├── server.go            # TCP Server
│       └── protocol.go          # 协议编解码
├── pkg/
│   └── utils/
│       └── encoding.go          # 通用编码工具
├── test/
│   ├── btree_test.go
│   ├── buffer_pool_test.go
│   ├── wal_test.go
│   └── integration_test.go
├── bench/
│   └── storage_bench_test.go    # 基准测试
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 6. 性能目标

| 指标 | 目标值 |
|------|--------|
| 单条 INSERT 延迟 | < 1ms (Buffer Pool hit) |
| 10w INSERT 总时间 | < 10s |
| 点查 QPS（单连接） | > 5,000 |
| 范围扫描 1000 条 | < 50ms |
| Crash Recovery 恢复时间 | < 5s (100w 记录) |

---

## 7. 关键学习产出

- **IO 模型理解：** 为什么数据库以页为单位读写？mmap vs pread？
- **索引原理：** B+Tree 为什么比 B-Tree、Hash 更适合数据库？
- **持久化保证：** WAL 如何保证 ACID 中的 D？
- **内存管理：** Buffer Pool 与 OS Page Cache 的关系
- **性能分析能力：** 使用 pprof 分析 CPU / 内存 / IO 瓶颈

---

## 8. 毕业标准

- [ ] B+Tree 能正确处理 100w 条数据的 CRUD
- [ ] WAL 能在 crash 后正确恢复数据
- [ ] Buffer Pool 命中率在读密集场景下 > 90%
- [ ] 能通过 TCP 接受客户端 SQL 请求
- [ ] 完成压测报告，清楚系统极限
- [ ] 能清晰画出数据从客户端到磁盘的完整路径
