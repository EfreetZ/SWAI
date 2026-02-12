# Stage 2 — 基础组件自研: mini-components

> 预计周期：3 周 | 语言：Golang | 难度：⭐⭐⭐
>
> **让你真正理解 Redis / MQ 为什么存在**

---

## 1. 项目目标

在接触中间件之前，先自己实现它们的核心组件，建立第一性原理理解：

- **LRU / LFU Cache** → 理解 Redis 内存淘汰策略的本质
- **延迟队列** → 理解 MQ 延迟消息的核心机制
- **Worker Pool** → 理解 goroutine 管理与资源控制
- **熔断器 / 限流器** → 理解微服务容错的基石
- **布隆过滤器** → 理解缓存穿透的防御手段

---

## 2. 组件一：LRU Cache

### 核心思想

最近最少使用淘汰：当缓存满时，移除**最久没有被访问**的数据。

### 数据结构

```
HashMap + 双向链表 = O(1) 查找 + O(1) 淘汰

┌──────────┐
│ HashMap  │──key──►┌──────┐◄──►┌──────┐◄──►┌──────┐
│          │        │ Node │    │ Node │    │ Node │
└──────────┘        │(head)│    │      │    │(tail)│
                    └──────┘    └──────┘    └──────┘
                     最近使用              最久未使用
```

### 接口设计

```go
// lru.go
type LRUCache struct {
    capacity int
    cache    map[string]*Node
    head     *Node  // 哨兵头节点（最近使用端）
    tail     *Node  // 哨兵尾节点（最久未使用端）
    mu       sync.RWMutex
}

type Node struct {
    Key   string
    Value interface{}
    Prev  *Node
    Next  *Node
}

func NewLRUCache(capacity int) *LRUCache

// Get：命中则移到头部，返回值；未命中返回 nil, false
func (c *LRUCache) Get(key string) (interface{}, bool)

// Put：插入/更新，超容量则淘汰尾部
func (c *LRUCache) Put(key string, value interface{})

// Len：当前缓存条目数
func (c *LRUCache) Len() int

// 内部方法
func (c *LRUCache) moveToHead(node *Node)    // 标记为最近使用
func (c *LRUCache) removeNode(node *Node)     // 从链表移除
func (c *LRUCache) addToHead(node *Node)      // 添加到头部
func (c *LRUCache) removeTail() *Node         // 淘汰最久未使用
```

### 面试扩展

- **线程安全版本：** 加 `sync.RWMutex`，Get 用读锁，Put 用写锁
- **带 TTL 版本：** Node 增加 `ExpireAt` 字段，Get 时惰性检查

---

## 3. 组件二：LFU Cache

### 核心思想

最不经常使用淘汰：当缓存满时，移除**访问频率最低**的数据。频率相同时，淘汰最早的。

### 数据结构

```
HashMap + 频率桶（每个桶是一个双向链表）

频率桶:
 freq=1: [Node_C] ←→ [Node_D]
 freq=2: [Node_B]
 freq=5: [Node_A]
 minFreq = 1（指向最低频率桶，O(1) 淘汰）
```

### 接口设计

```go
// lfu.go
type LFUCache struct {
    capacity int
    minFreq  int
    cache    map[string]*LFUNode
    freqMap  map[int]*DoublyLinkedList  // freq → 该频率的节点链表
    mu       sync.RWMutex
}

type LFUNode struct {
    Key   string
    Value interface{}
    Freq  int
    Prev  *LFUNode
    Next  *LFUNode
}

func NewLFUCache(capacity int) *LFUCache
func (c *LFUCache) Get(key string) (interface{}, bool)
func (c *LFUCache) Put(key string, value interface{})

// 内部：访问时频率+1，从旧桶移到新桶
func (c *LFUCache) increaseFreq(node *LFUNode)
```

### LRU vs LFU 对比

| 特性 | LRU | LFU |
|------|-----|-----|
| 淘汰依据 | 最久未访问 | 访问频率最低 |
| 适用场景 | 访问模式有时间局部性 | 有明确的冷热数据 |
| 问题 | 偶尔的全量扫描会污染缓存 | 新数据频率低容易被误淘汰 |
| Redis 对应 | `volatile-lru` / `allkeys-lru` | `volatile-lfu` / `allkeys-lfu` |

---

## 4. 组件三：延迟队列

### 方案对比

| 方案 | 时间复杂度 | 精度 | 适用场景 |
|------|-----------|------|---------|
| 最小堆 | 入队 O(logN)，出队 O(logN) | 精确 | 任务量不大 |
| 时间轮 | 入队 O(1)，出队 O(1) | 槽位粒度 | 海量定时任务 |

### 最小堆实现

```go
// delay_queue_heap.go
type DelayTask struct {
    ID       string
    ExecuteAt time.Time
    Callback  func()
}

type DelayQueue struct {
    heap   []*DelayTask       // 最小堆（按 ExecuteAt 排序）
    index  map[string]int     // taskID → 堆中位置
    mu     sync.Mutex
    notify chan struct{}       // 新任务到来时通知调度协程
}

func NewDelayQueue() *DelayQueue
func (dq *DelayQueue) Add(task *DelayTask) error
func (dq *DelayQueue) Cancel(taskID string) error
func (dq *DelayQueue) Start(ctx context.Context)  // 启动调度协程
```

### 时间轮实现

```go
// timing_wheel.go

// 分层时间轮（类似时钟：秒针 → 分针 → 时针）
//
// 单层时间轮：
// ┌───┬───┬───┬───┬───┬───┬───┬───┐
// │ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │ 6 │ 7 │  ← 8 个槽位
// └───┴───┴───┴───┴───┴───┴───┴───┘
//           ▲
//           │ 当前指针（tick）
//
// 每个 tick（如 100ms）指针前进一格
// 槽位上挂载到期任务链表

type TimingWheel struct {
    tickDuration time.Duration  // 每个 tick 的时长
    wheelSize    int            // 槽位数
    slots        []*list.List   // 每个槽位的任务链表
    currentPos   int            // 当前指针位置
    ticker       *time.Ticker
    overflow     *TimingWheel   // 上层时间轮（处理更大延迟）
    mu           sync.Mutex
}

func NewTimingWheel(tickDuration time.Duration, wheelSize int) *TimingWheel
func (tw *TimingWheel) Add(delay time.Duration, callback func()) *Timer
func (tw *TimingWheel) Start(ctx context.Context)
func (tw *TimingWheel) Stop()
```

---

## 5. 组件四：Worker Pool

### 设计目标

控制 goroutine 数量，避免无限创建导致 OOM。

```go
// worker_pool.go
type Task func()

type WorkerPool struct {
    maxWorkers int
    taskQueue  chan Task
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

func NewWorkerPool(maxWorkers, queueSize int) *WorkerPool

// Submit：提交任务，队列满时阻塞或返回错误
func (p *WorkerPool) Submit(task Task) error

// SubmitWithTimeout：带超时的提交
func (p *WorkerPool) SubmitWithTimeout(task Task, timeout time.Duration) error

// Start：启动 worker 协程
func (p *WorkerPool) Start()

// Shutdown：优雅关闭（等待所有任务完成）
func (p *WorkerPool) Shutdown()

// Running：当前活跃的 worker 数
func (p *WorkerPool) Running() int
```

### 优雅关闭流程

```
1. 调用 cancel() 通知所有 worker 停止接收新任务
2. 关闭 taskQueue channel
3. 每个 worker 处理完当前任务后退出
4. wg.Wait() 等待所有 worker 结束
```

---

## 6. 组件五：熔断器

### 三态转换

```
         连续失败 >= threshold
  Closed ─────────────────────► Open
    ▲                              │
    │ 连续成功 >= threshold        │ 超时后
    │                              ▼
    └──────────────────────── HalfOpen
                              (允许少量请求试探)
```

```go
// circuit_breaker.go
type State int

const (
    Closed   State = 0  // 正常
    Open     State = 1  // 熔断
    HalfOpen State = 2  // 半开试探
)

type CircuitBreaker struct {
    state          State
    failureCount   int64
    successCount   int64
    failThreshold  int64         // 失败多少次触发熔断
    successThreshold int64       // 半开状态成功多少次恢复
    timeout        time.Duration // Open → HalfOpen 等待时间
    lastFailTime   time.Time
    halfOpenMax    int64         // 半开状态允许通过的最大请求数
    halfOpenCurr   int64
    mu             sync.Mutex
}

func NewCircuitBreaker(failThreshold, successThreshold int64, timeout time.Duration) *CircuitBreaker

// Execute：执行函数，自动记录成功/失败，触发状态转换
func (cb *CircuitBreaker) Execute(fn func() error) error

func (cb *CircuitBreaker) State() State
func (cb *CircuitBreaker) Reset()
```

---

## 7. 组件六：限流器（令牌桶）

```go
// token_bucket.go

// 令牌桶算法：
// - 以固定速率往桶中放令牌
// - 请求到来时取一个令牌，取到则放行，取不到则拒绝
// - 桶有容量上限（burst），允许突发流量

type TokenBucket struct {
    rate       float64     // 每秒产生的令牌数
    burst      int         // 桶容量（最大突发）
    tokens     float64     // 当前令牌数
    lastTime   time.Time   // 上次计算时间
    mu         sync.Mutex
}

func NewTokenBucket(rate float64, burst int) *TokenBucket

// Allow：是否允许通过（消耗一个令牌）
func (tb *TokenBucket) Allow() bool

// AllowN：是否允许 N 个请求通过
func (tb *TokenBucket) AllowN(n int) bool

// Wait：阻塞等待直到有令牌可用
func (tb *TokenBucket) Wait(ctx context.Context) error
```

---

## 8. 组件七：布隆过滤器

```go
// bloom_filter.go

// 布隆过滤器：
// - 用 k 个哈希函数将元素映射到位数组的 k 个位置
// - 查询时检查所有 k 个位置是否都为 1
// - 可能误判（false positive），但不会漏判（no false negative）
//
// 适用场景：缓存穿透防御、去重

type BloomFilter struct {
    bits    []byte       // 位数组
    size    uint64       // 位数组大小
    hashNum uint64       // 哈希函数个数
}

func NewBloomFilter(expectedItems uint64, falsePositiveRate float64) *BloomFilter

// Add：添加元素
func (bf *BloomFilter) Add(data []byte)

// Contains：查询元素是否可能存在
func (bf *BloomFilter) Contains(data []byte) bool

// EstimateFalsePositiveRate：估算当前误判率
func (bf *BloomFilter) EstimateFalsePositiveRate() float64

// 最优参数计算：
// bits = -n * ln(p) / (ln2)^2
// hashNum = (bits / n) * ln2
```

---

## 9. 目录结构

```
projects/stage2-mini-components/
├── lru/
│   ├── lru.go
│   ├── lru_concurrent.go     # 线程安全版本
│   ├── lru_ttl.go            # 带 TTL 版本
│   └── lru_test.go
├── lfu/
│   ├── lfu.go
│   └── lfu_test.go
├── delayqueue/
│   ├── heap_queue.go         # 最小堆版本
│   ├── timing_wheel.go       # 时间轮版本
│   └── delay_test.go
├── workerpool/
│   ├── pool.go
│   └── pool_test.go
├── circuitbreaker/
│   ├── breaker.go
│   └── breaker_test.go
├── ratelimiter/
│   ├── token_bucket.go
│   ├── sliding_window.go     # 滑动窗口版本
│   └── limiter_test.go
├── bloomfilter/
│   ├── bloom.go
│   └── bloom_test.go
├── bench/
│   └── bench_test.go         # 所有组件基准测试
├── go.mod
├── Makefile
└── README.md
```

---

## 10. 实现里程碑

### M1: 缓存组件（第1周）
- [ ] LRU Cache 实现 + 测试
- [ ] LRU 线程安全版本 + 带 TTL 版本
- [ ] LFU Cache 实现 + 测试
- [ ] 基准测试：LRU vs LFU 在不同访问模式下的命中率

### M2: 延迟队列 + Worker Pool（第2周）
- [ ] 最小堆延迟队列
- [ ] 时间轮延迟队列
- [ ] Worker Pool + 优雅关闭
- [ ] 基准测试：时间轮 vs 最小堆在大量任务下的表现

### M3: 容错组件 + 布隆过滤器（第3周）
- [ ] 熔断器（三态转换 + 自动恢复）
- [ ] 令牌桶限流器
- [ ] 滑动窗口限流器
- [ ] 布隆过滤器
- [ ] 集成测试：模拟缓存击穿 → 用布隆过滤器 + 熔断器防御

---

## 11. 毕业标准

- [ ] 所有组件通过测试 + 基准测试
- [ ] 能解释 LRU/LFU 各自的适用场景
- [ ] 能解释时间轮为什么比最小堆更适合海量定时任务
- [ ] 熔断器能正确在失败时熔断、恢复时放行
- [ ] 限流器在超 QPS 时准确拒绝
- [ ] 布隆过滤器误判率在理论范围内
