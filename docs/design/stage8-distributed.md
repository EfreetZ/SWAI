# Stage 8 — Distributed Coordination: mini-raft

> 预计周期：1.5 个月 | 语言：Golang | 难度：⭐⭐⭐⭐⭐
>
> 分布式系统核心：共识是一切分布式系统的基石

---

## 1. 项目目标

实现一个简化版 Raft 共识算法及其上层 KV 存储，深入理解：

- **CAP 定理：** 一致性、可用性、分区容忍三者不可兼得
- **共识问题：** 多节点如何就同一个值达成一致
- **Leader Election：** 选主机制与任期（Term）概念
- **Log Replication：** 日志如何在集群中安全复制
- **Safety：** 为什么 Raft 能保证线性一致性

**最终交付物：** 一个 3~5 节点的分布式 KV 存储，支持 Put / Get / Delete，具备容错能力。

---

## 2. 整体架构

```
┌───────────────────────────────────────────┐
│               Client (CLI / HTTP)          │
└──────────────────┬────────────────────────┘
                   │
      ┌────────────▼────────────┐
      │       KV Service        │
      │   (线性化读写接口)       │
      └────────────┬────────────┘
                   │
      ┌────────────▼────────────┐
      │      Raft Consensus     │
      │                         │
      │  ┌──────┐ ┌──────────┐ │
      │  │Leader│ │ Election │ │
      │  │ Log  │ │  Timer   │ │
      │  └──────┘ └──────────┘ │
      │  ┌──────────────────┐  │
      │  │ Log Replication  │  │
      │  └──────────────────┘  │
      │  ┌──────────────────┐  │
      │  │   State Machine  │  │
      │  │  (KV Store Apply)│  │
      │  └──────────────────┘  │
      └────────────┬────────────┘
                   │ RPC
         ┌─────────┼─────────┐
         │         │         │
    ┌────▼──┐ ┌───▼───┐ ┌──▼─────┐
    │Node 1 │ │Node 2 │ │Node 3  │
    │(Leader)│ │(Follwr)│ │(Follwr)│
    └───────┘ └───────┘ └────────┘
```

---

## 3. Raft 核心数据结构

### 3.1 节点状态

```go
// raft.go
type NodeState int

const (
    Follower  NodeState = 0
    Candidate NodeState = 1
    Leader    NodeState = 2
)

type RaftNode struct {
    // 持久化状态（写入磁盘，crash 后恢复）
    currentTerm int64       // 当前任期
    votedFor    string      // 本任期投票给谁
    log         []LogEntry  // 日志条目

    // 易失状态（所有节点）
    commitIndex int64       // 已提交的最高日志索引
    lastApplied int64       // 已应用到状态机的最高索引

    // 易失状态（仅 Leader）
    nextIndex   map[string]int64  // 每个 Follower 的下一个发送索引
    matchIndex  map[string]int64  // 每个 Follower 已匹配的最高索引

    // 运行时
    id          string
    state       NodeState
    leader      string
    peers       []string
    mu          sync.Mutex

    // 通道
    applyCh     chan ApplyMsg
    electionTimer  *time.Timer
    heartbeatTimer *time.Timer
}

type LogEntry struct {
    Index   int64
    Term    int64
    Command []byte  // 序列化的命令
}

type ApplyMsg struct {
    CommandValid bool
    Command      []byte
    CommandIndex int64
    CommandTerm  int64
}
```

### 3.2 RPC 定义

```go
// rpc.go

// === RequestVote RPC ===
type RequestVoteArgs struct {
    Term         int64   // 候选人任期
    CandidateID  string  // 候选人 ID
    LastLogIndex int64   // 候选人最后日志索引
    LastLogTerm  int64   // 候选人最后日志任期
}

type RequestVoteReply struct {
    Term        int64   // 当前任期（用于候选人更新自己）
    VoteGranted bool    // 是否投票
}

// === AppendEntries RPC ===
type AppendEntriesArgs struct {
    Term         int64       // Leader 任期
    LeaderID     string      // Leader ID
    PrevLogIndex int64       // 新日志前一条的索引
    PrevLogTerm  int64       // 新日志前一条的任期
    Entries      []LogEntry  // 要追加的日志（心跳时为空）
    LeaderCommit int64       // Leader 的 commitIndex
}

type AppendEntriesReply struct {
    Term    int64   // 当前任期
    Success bool    // 是否追加成功

    // 优化：快速回退
    ConflictTerm  int64  // 冲突日志的任期
    ConflictIndex int64  // 该任期的第一条日志索引
}

// === InstallSnapshot RPC ===
type InstallSnapshotArgs struct {
    Term              int64
    LeaderID          string
    LastIncludedIndex int64
    LastIncludedTerm  int64
    Data              []byte  // 快照数据
}

type InstallSnapshotReply struct {
    Term int64
}
```

---

## 4. Leader Election（选主）

### 流程

```
Follower ──(选举超时)──► Candidate ──(获得多数票)──► Leader
    ▲                       │                         │
    │                       │(发现更高 Term)           │
    │                       ▼                         │
    └───────────────── Follower ◄─────────────────────┘
                        (发现更高 Term)
```

### 实现

```go
// election.go
func (rn *RaftNode) startElection() {
    rn.mu.Lock()
    rn.currentTerm++
    rn.state = Candidate
    rn.votedFor = rn.id
    rn.persist()  // 持久化 currentTerm 和 votedFor

    term := rn.currentTerm
    lastLogIndex := rn.lastLogIndex()
    lastLogTerm := rn.lastLogTerm()
    rn.mu.Unlock()

    votes := int32(1) // 投给自己

    for _, peer := range rn.peers {
        go func(peer string) {
            args := &RequestVoteArgs{
                Term:         term,
                CandidateID:  rn.id,
                LastLogIndex: lastLogIndex,
                LastLogTerm:  lastLogTerm,
            }
            reply := &RequestVoteReply{}

            if rn.sendRequestVote(peer, args, reply) {
                rn.mu.Lock()
                defer rn.mu.Unlock()

                if reply.Term > rn.currentTerm {
                    rn.becomeFollower(reply.Term)
                    return
                }

                if reply.VoteGranted && rn.state == Candidate && rn.currentTerm == term {
                    if atomic.AddInt32(&votes, 1) > int32(len(rn.peers)/2) {
                        rn.becomeLeader()
                    }
                }
            }
        }(peer)
    }
}

// 选举超时：150ms ~ 300ms 随机
func (rn *RaftNode) resetElectionTimer() {
    timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
    rn.electionTimer.Reset(timeout)
}
```

### 投票规则

```go
func (rn *RaftNode) handleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
    rn.mu.Lock()
    defer rn.mu.Unlock()

    reply.Term = rn.currentTerm
    reply.VoteGranted = false

    // 1. 任期过旧，拒绝
    if args.Term < rn.currentTerm {
        return
    }

    // 2. 发现更高任期，转为 Follower
    if args.Term > rn.currentTerm {
        rn.becomeFollower(args.Term)
    }

    // 3. 检查是否已投票
    if rn.votedFor != "" && rn.votedFor != args.CandidateID {
        return
    }

    // 4. 日志至少和自己一样新（Election Restriction）
    if args.LastLogTerm > rn.lastLogTerm() ||
       (args.LastLogTerm == rn.lastLogTerm() && args.LastLogIndex >= rn.lastLogIndex()) {
        rn.votedFor = args.CandidateID
        rn.persist()
        reply.VoteGranted = true
        rn.resetElectionTimer()
    }
}
```

---

## 5. Log Replication（日志复制）

### Leader 发送日志

```go
// replication.go
func (rn *RaftNode) broadcastAppendEntries() {
    for _, peer := range rn.peers {
        go rn.sendAppendEntriesToPeer(peer)
    }
}

func (rn *RaftNode) sendAppendEntriesToPeer(peer string) {
    rn.mu.Lock()
    nextIdx := rn.nextIndex[peer]
    prevLogIndex := nextIdx - 1
    prevLogTerm := rn.logTermAt(prevLogIndex)

    // 如果需要的日志已被快照截断，发送快照
    if prevLogIndex < rn.lastSnapshotIndex {
        rn.mu.Unlock()
        rn.sendInstallSnapshot(peer)
        return
    }

    entries := rn.logSlice(nextIdx)
    args := &AppendEntriesArgs{
        Term:         rn.currentTerm,
        LeaderID:     rn.id,
        PrevLogIndex: prevLogIndex,
        PrevLogTerm:  prevLogTerm,
        Entries:      entries,
        LeaderCommit: rn.commitIndex,
    }
    rn.mu.Unlock()

    reply := &AppendEntriesReply{}
    if rn.sendAppendEntries(peer, args, reply) {
        rn.handleAppendEntriesReply(peer, args, reply)
    }
}

func (rn *RaftNode) handleAppendEntriesReply(peer string, args *AppendEntriesArgs, reply *AppendEntriesReply) {
    rn.mu.Lock()
    defer rn.mu.Unlock()

    if reply.Term > rn.currentTerm {
        rn.becomeFollower(reply.Term)
        return
    }

    if reply.Success {
        rn.nextIndex[peer] = args.PrevLogIndex + int64(len(args.Entries)) + 1
        rn.matchIndex[peer] = rn.nextIndex[peer] - 1
        rn.advanceCommitIndex()
    } else {
        // 快速回退优化
        if reply.ConflictTerm > 0 {
            rn.nextIndex[peer] = reply.ConflictIndex
        } else {
            rn.nextIndex[peer]--
        }
    }
}
```

### Commit 推进

```go
func (rn *RaftNode) advanceCommitIndex() {
    // 找到大多数节点已匹配的最高索引
    matches := make([]int64, 0, len(rn.peers)+1)
    matches = append(matches, rn.lastLogIndex())
    for _, idx := range rn.matchIndex {
        matches = append(matches, idx)
    }
    sort.Slice(matches, func(i, j int) bool { return matches[i] > matches[j] })

    majority := matches[len(matches)/2]

    // 只提交当前任期的日志（Raft Safety）
    if majority > rn.commitIndex && rn.logTermAt(majority) == rn.currentTerm {
        rn.commitIndex = majority
        rn.applyLogs()
    }
}
```

---

## 6. 持久化 & 快照

```go
// persist.go
type PersistState struct {
    CurrentTerm int64
    VotedFor    string
    Log         []LogEntry
}

func (rn *RaftNode) persist() {
    state := &PersistState{
        CurrentTerm: rn.currentTerm,
        VotedFor:    rn.votedFor,
        Log:         rn.log,
    }
    data := encode(state) // gob / protobuf
    rn.persister.SaveState(data)
}

func (rn *RaftNode) readPersist(data []byte) {
    state := decode(data)
    rn.currentTerm = state.CurrentTerm
    rn.votedFor = state.VotedFor
    rn.log = state.Log
}

// snapshot.go
type Snapshot struct {
    LastIncludedIndex int64
    LastIncludedTerm  int64
    Data              []byte  // 状态机快照
}

func (rn *RaftNode) TakeSnapshot(index int64, data []byte) {
    // 截断 index 之前的日志
    // 保存快照到磁盘
}
```

---

## 7. 上层 KV Service

```go
// kv_server.go
type KVServer struct {
    raft      *RaftNode
    store     map[string]string
    applyCh   chan ApplyMsg
    notifyCh  map[int64]chan *CommandResult
    lastApply map[string]int64  // 去重：clientID → 最后请求序号
    mu        sync.RWMutex
}

type Op struct {
    Type      string  // "Put" / "Get" / "Delete"
    Key       string
    Value     string
    ClientID  string
    SeqNum    int64
}

type CommandResult struct {
    Value string
    Err   string
}

func (kv *KVServer) Put(key, value string) error {
    op := &Op{Type: "Put", Key: key, Value: value}
    return kv.propose(op)
}

func (kv *KVServer) Get(key string) (string, error) {
    op := &Op{Type: "Get", Key: key}
    return kv.proposeAndWait(op)
}

func (kv *KVServer) propose(op *Op) error {
    // 1. 将 Op 提交给 Raft
    index, term, isLeader := kv.raft.Start(encode(op))
    if !isLeader {
        return ErrNotLeader
    }
    // 2. 等待 Raft apply
    // 3. 返回结果
}

// 后台协程：监听 applyCh，将已提交的命令应用到状态机
func (kv *KVServer) applyLoop() {
    for msg := range kv.applyCh {
        op := decode(msg.Command)
        // 去重检查
        // 应用到 store
        // 通知等待的客户端
    }
}
```

### HTTP 接口

```go
// http.go
func (kv *KVServer) ServeHTTP() {
    http.HandleFunc("/put", kv.handlePut)    // PUT /put?key=x&value=y
    http.HandleFunc("/get", kv.handleGet)    // GET /get?key=x
    http.HandleFunc("/delete", kv.handleDel) // DELETE /delete?key=x
    http.HandleFunc("/status", kv.handleStatus) // 集群状态
}
```

---

## 8. 目录结构

```
mini-raft/
├── cmd/
│   └── server/
│       └── main.go              # 节点启动入口
├── internal/
│   ├── raft/
│   │   ├── raft.go              # Raft 核心状态机
│   │   ├── election.go          # Leader Election
│   │   ├── replication.go       # Log Replication
│   │   ├── persist.go           # 持久化
│   │   ├── snapshot.go          # 快照
│   │   ├── rpc.go               # RPC 定义
│   │   └── util.go              # 工具函数
│   ├── kvstore/
│   │   ├── kv_server.go         # KV Service
│   │   ├── state_machine.go     # 状态机
│   │   └── client.go            # KV Client
│   ├── transport/
│   │   ├── transport.go         # 网络传输接口
│   │   ├── grpc_transport.go    # gRPC 实现
│   │   └── http_transport.go    # HTTP 实现
│   ├── storage/
│   │   ├── persister.go         # 持久化接口
│   │   └── file_persister.go    # 文件持久化
│   └── server/
│       ├── http_server.go       # HTTP API
│       └── config.go            # 配置
├── test/
│   ├── raft_election_test.go    # 选主测试
│   ├── raft_replication_test.go # 日志复制测试
│   ├── raft_persist_test.go     # 持久化测试
│   ├── raft_snapshot_test.go    # 快照测试
│   ├── kv_test.go               # KV 功能测试
│   ├── linearizability_test.go  # 线性一致性测试
│   └── chaos_test.go            # 混沌测试
├── scripts/
│   ├── start_cluster.sh         # 启动 3 节点集群
│   └── chaos.sh                 # 混沌测试脚本
├── go.mod
├── Makefile
└── README.md
```

---

## 9. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| 网络分区（脑裂） | 节点状态监控 + 日志对比 | 各节点 term/state 日志 |
| Leader 拍打（频繁换主） | 选举超时配置 + 心跳监控 | election timeout 分布分析 |
| 日志复制延迟 | matchIndex 差异监控 | Leader nextIndex/matchIndex metrics |
| 日志无限增长 | 快照触发监控 | 日志文件大小 / snapshot 频率 |
| 线性一致性违反 | Jepsen 风格检测 | 操作日志录制 + 线性化检查 |
| Crash 后恢复失败 | 持久化数据完整性 | 重启后数据对比 / CRC 校验 |

---

## 10. 面试考点

### 必问级

- [ ] **CAP 定理：** Raft 在 CAP 中选择了什么？网络分区时如何表现？
- [ ] **Leader Election：** 选举超时为什么随机化？投票规则（Election Restriction）？
- [ ] **Log Replication：** 日志匹配属性（Log Matching Property）是什么？
- [ ] **Quorum：** 为什么多数派能保证安全性？

### 高频级

- [ ] **Raft vs Paxos：** 主要区别？Raft 为什么更容易实现？
- [ ] **快照：** 为什么需要快照？InstallSnapshot RPC 流程？
- [ ] **读一致性：** Leader Read / ReadIndex / Lease Read 的区别
- [ ] **etcd 应用：** 分布式锁 / 服务发现 / 配置中心的实现原理

---

## 11. 实现里程碑

### M1: 选主（第1-2周）
- [ ] 节点状态机（Follower / Candidate / Leader）
- [ ] RequestVote RPC
- [ ] 选举超时 + 随机化
- [ ] 测试：3 节点能选出 Leader；kill Leader 后重新选举

### M2: 日志复制（第3-4周）
- [ ] AppendEntries RPC（心跳 + 日志）
- [ ] Log Matching Property
- [ ] Commit 推进 + Apply
- [ ] 测试：Leader 写入后 Follower 日志一致

### M3: 持久化 + 恢复（第5周）
- [ ] 持久化 currentTerm / votedFor / log
- [ ] Crash 后从磁盘恢复
- [ ] 测试：kill 节点再重启，数据不丢失

### M4: 快照（第6周）
- [ ] TakeSnapshot + InstallSnapshot RPC
- [ ] 日志截断
- [ ] 测试：日志不会无限增长

### M5: KV Service + HTTP API（第7周）
- [ ] KV 上层服务
- [ ] 命令去重（Exactly-Once 语义）
- [ ] HTTP 接口
- [ ] 测试：通过 curl 读写 KV

### M6: 混沌测试（第8周）
- [ ] 网络分区模拟
- [ ] 随机 kill 节点
- [ ] 线性一致性检验
- [ ] 性能压测

---

## 10. 测试策略

### 确定性测试

```go
// 模拟网络，控制消息传递
type SimulatedNetwork struct {
    nodes     map[string]*RaftNode
    partitions map[string]map[string]bool  // 网络分区
    delay     time.Duration
}

func (n *SimulatedNetwork) Partition(nodes ...string)   // 制造网络分区
func (n *SimulatedNetwork) Heal()                        // 恢复网络
func (n *SimulatedNetwork) Disconnect(node string)       // 断开节点
func (n *SimulatedNetwork) Reconnect(node string)        // 重连节点
```

### 线性一致性验证

使用 Jepsen 风格的检测：记录所有操作的调用时间和返回时间，验证是否存在合法的线性化顺序。

---

## 11. 性能目标

| 指标 | 目标值 |
|------|--------|
| 写入 QPS（3 节点） | > 5,000 |
| 读取 QPS（Leader Read） | > 10,000 |
| Leader 选举时间 | < 3s |
| 日志复制延迟（P99） | < 10ms |
| 快照恢复时间（1GB 数据） | < 30s |

---

## 12. 关键学习产出

- **CAP 直觉：** 网络分区时 Raft 如何选择一致性而牺牲可用性
- **共识本质：** 多数派（Quorum）为什么能保证安全性
- **Term 的意义：** 逻辑时钟如何解决分布式时序问题
- **Safety vs Liveness：** Raft 保证 Safety，通过随机超时保证 Liveness
- **工程复杂度：** 论文与实现之间的巨大鸿沟

---

## 13. 毕业标准

- [ ] 3 节点集群能正确选主、复制日志
- [ ] Kill 少数节点后集群仍可用
- [ ] Kill 多数节点后集群不可用（正确行为）
- [ ] 网络分区恢复后数据一致
- [ ] 能通过线性一致性检验
- [ ] 能清晰解释 Raft 论文中的每个 Safety 属性
