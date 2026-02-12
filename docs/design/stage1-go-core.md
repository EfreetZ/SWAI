# Stage 1 — Go 核心能力: go-core

> 预计周期：3 周 | 语言：Golang | 难度：⭐⭐
>
> 目标：理解语言如何支撑高并发系统，详细代码在项目实现中完成

---

## 1. 项目目标

建立扎实的 Go 语言功底，不是"能写"，而是"知道为什么这样写"：

- **语法精通** — 理解每个特性的底层机制
- **并发模型** — 深入 GMP 调度，写出正确的并发代码
- **数据结构** — 手写实现核心数据结构，为后续基础设施项目打基础
- **网络基础** — TCP/HTTP 编程，为 mini-redis / mini-mysql 准备

---

## 2. 四大模块

### 2.1 语法基础

关键知识点：
- 类型系统、控制流、函数（多返回值/闭包/defer）
- 指针与值传递
- 结构体与方法（值接收者 vs 指针接收者）
- 接口（隐式实现、iface/eface 内部结构）
- 错误处理模式、泛型基础

### 2.2 并发编程

关键知识点：
- GMP 调度模型（G/M/P 关系、Work Stealing、抢占式调度）
- Channel（buffered/unbuffered、select、deadlock 场景）
- sync 包（Mutex/RWMutex/WaitGroup/Once/Pool/Map）
- Context（超时/取消/传值、取消传播链路）
- 并发模式（fan-in/fan-out/pipeline/worker pool）
- 数据竞争检测（`go test -race`）

### 2.3 数据结构与算法

关键实现（每个都要手写 + 测试 + 基准测试）：
- Go slice 底层结构与扩容机制
- Go map 底层（hmap → bmap → overflow → 渐进式扩容）
- 链表（单链表/双向链表）+ LRU Cache
- B+Tree（完整实现，为 Stage 4 mini-mysql 准备）
- SkipList（完整实现，为 Stage 5 mini-redis 准备）
- 一致性哈希（完整实现，为 Stage 5 mini-redis Cluster 准备）

### 2.4 网络编程基础

关键知识点：
- TCP/IP 协议栈、三次握手/四次挥手
- Go `net` 包 TCP Server/Client
- HTTP 协议 + Go `net/http`
- JSON 序列化

---

## 3. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 |
|---------|---------|---------|
| Goroutine 泄漏 | pprof goroutine profile | `go tool pprof http://localhost:6060/debug/pprof/goroutine` |
| 死锁 | dlv 调试器 | `dlv debug ./cmd/server` → `goroutines` |
| CPU 飙高 | pprof CPU 火焰图 | `go tool pprof -http=:8080 profile.pb.gz` |
| 内存暴涨 | pprof heap + 逃逸分析 | `go tool compile -m main.go` |
| 数据竞争 | race detector | `go test -race ./...` |

---

## 4. 面试考点（按频率排序）

### 必问级

- [ ] **GMP 调度模型**：G/M/P 各自职责？Work Stealing 怎么工作？Go 1.14 之后抢占式调度如何实现？
- [ ] **Goroutine 泄漏**：常见场景（channel 阻塞、无退出条件循环、context 未传递）？如何排查？
- [ ] **Channel 死锁**：哪些操作会 panic？哪些会阻塞？
- [ ] **Go map 并发安全**：为什么不安全？`sync.Map` 适用场景？

### 高频级

- [ ] **defer 执行顺序**：LIFO 栈、参数求值时机、与命名返回值的交互
- [ ] **值接收者 vs 指针接收者**：对 interface 实现的影响
- [ ] **interface 内部结构**：iface/eface 的区别、`interface{} == nil` 陷阱
- [ ] **slice 扩容策略**：Go 1.18+ 的变化（< 256 翻倍，>= 256 约 1.25 倍）
- [ ] **Go map 底层结构**：hmap → bmap → overflow、扩容条件（负载因子 6.5 / 溢出桶过多）

### 架构级

- [ ] **Context 取消传播**：WithCancel/WithTimeout 的实现机制
- [ ] **B+Tree vs 红黑树**：为什么数据库选 B+Tree？（磁盘友好、范围查询）
- [ ] **一致性哈希**：虚拟节点的作用？数据迁移量如何最小化？
- [ ] **TCP TIME_WAIT**：为什么存在？过多怎么处理？

---

## 5. 目录结构

```
projects/stage1-go-core/
├── basics/                    # 语法基础
│   ├── types/
│   ├── functions/
│   ├── structs/
│   ├── interfaces/
│   └── errors/
├── concurrency/               # 并发编程
│   ├── goroutine/
│   ├── channel/
│   ├── sync/
│   ├── context/
│   ├── patterns/             # 并发模式
│   └── race/                 # 数据竞争案例
├── datastructures/            # 数据结构
│   ├── slice/
│   ├── map/
│   ├── linkedlist/
│   ├── btree/
│   ├── skiplist/
│   ├── consistent_hash/
│   └── sort/
├── network/                   # 网络编程
│   ├── tcp/
│   └── http/
├── bench/                     # 基准测试汇总
├── go.mod
├── Makefile
└── README.md
```

---

## 6. 实现里程碑

### M1: 语法基础（第1周）
- [ ] 所有语法知识点代码 + 测试
- [ ] defer / interface / error 深度案例
- [ ] 面试考点代码验证

### M2: 并发编程（第2周）
- [ ] GMP 模型验证实验（GOMAXPROCS / schedtrace）
- [ ] Channel + sync 所有模式
- [ ] Goroutine 泄漏案例 + 修复
- [ ] 数据竞争检测全部通过 (`go test -race`)

### M3: 数据结构 + 网络（第3周）
- [ ] B+Tree 完整实现 + 测试
- [ ] SkipList 完整实现 + 测试
- [ ] 一致性哈希实现 + 测试
- [ ] TCP Echo Server + 简单 HTTP Server
- [ ] 基准测试报告

---

## 7. 毕业标准

- [ ] 能白板画出 GMP 调度模型并解释 Work Stealing
- [ ] 能识别和修复 goroutine 泄漏（给出 3 种场景 + 排查方式）
- [ ] 手写 LRU Cache / B+Tree / SkipList 通过所有测试
- [ ] 能解释 Go map 底层结构和扩容机制（画图级别）
- [ ] `go test -race` 零告警
- [ ] 所有面试考点能给出代码级别的解释
