# Stage 7 — RPC Framework: mini-rpc

> 预计周期：1.5 个月 | 语言：Golang | 难度：⭐⭐⭐⭐
>
> 进入分布式系统领域：网络是系统的边界

---

## 1. 项目目标

从零构建一个生产级 RPC 框架，深入理解：

- **网络是系统的边界：** 远程调用的不可靠性本质
- **服务发现：** 如何找到可用的服务实例
- **负载均衡：** 请求如何分配到多个实例
- **容错机制：** Timeout / Retry / Circuit Breaker
- **序列化协议：** 如何高效传输结构化数据

---

## 2. 整体架构

```
┌──────────────────────────────────────────────────────┐
│                      Client                           │
│                                                       │
│  ┌─────────┐  ┌──────────┐  ┌───────────────────┐   │
│  │  Proxy   │  │ Protocol │  │ Service Discovery │   │
│  │ (Stub)   │──│ Encoder  │──│   (Registry)      │   │
│  └─────────┘  └──────────┘  └───────────────────┘   │
│       │                              │                │
│  ┌────▼────────────────────┐   ┌─────▼─────┐        │
│  │    Load Balancer        │   │  Registry  │        │
│  │ (RR/WRR/Random/Hash)   │   │  Client    │        │
│  └────┬────────────────────┘   └───────────┘        │
│       │                                              │
│  ┌────▼────────────────────────────────────────┐    │
│  │    Transport Layer (TCP / HTTP2)             │    │
│  │  ┌────────┐ ┌─────────┐ ┌────────────────┐  │    │
│  │  │Timeout │ │  Retry  │ │Circuit Breaker │  │    │
│  │  └────────┘ └─────────┘ └────────────────┘  │    │
│  └──────────────────────────────────────────────┘    │
└───────────────────────┬──────────────────────────────┘
                        │ Network
┌───────────────────────▼──────────────────────────────┐
│                      Server                           │
│                                                       │
│  ┌──────────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ TCP Listener │  │ Protocol │  │   Service      │  │
│  │              │──│ Decoder  │──│   Handler      │  │
│  └──────────────┘  └──────────┘  └───────────────┘  │
│                                                       │
│  ┌────────────────────────────────────────────────┐  │
│  │    Middleware Chain                              │  │
│  │  (logging → metrics → auth → rateLimit → biz)  │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
                        │
                ┌───────▼───────┐
                │   Registry    │
                │  (etcd-like)  │
                └───────────────┘
```

---

## 3. 协议设计

### 3.1 自定义二进制协议

```
┌───────────────────────────────────────────────────────┐
│                    RPC Frame                           │
├──────────┬──────────┬──────────┬──────────────────────┤
│  Magic   │ Version  │  Type    │   Codec              │
│  2 bytes │ 1 byte   │ 1 byte  │   1 byte             │
├──────────┴──────────┴──────────┴──────────────────────┤
│  Request ID                                            │
│  8 bytes                                               │
├────────────────────────────────────────────────────────┤
│  Payload Length                                         │
│  4 bytes                                               │
├────────────────────────────────────────────────────────┤
│  Metadata Length + Metadata (key-value pairs)          │
│  variable                                              │
├────────────────────────────────────────────────────────┤
│  Payload                                               │
│  variable                                              │
└────────────────────────────────────────────────────────┘
```

```go
// protocol.go
const MagicNumber uint16 = 0x4D52 // "MR" = Mini RPC

type MessageType byte

const (
    Request  MessageType = 0x00
    Response MessageType = 0x01
    Heartbeat MessageType = 0x02
)

type CodecType byte

const (
    JSON     CodecType = 0x00
    Protobuf CodecType = 0x01
    Msgpack  CodecType = 0x02
)

type Header struct {
    Magic         uint16
    Version       byte
    Type          MessageType
    Codec         CodecType
    RequestID     uint64
    PayloadLength uint32
}

type RPCRequest struct {
    ServiceName string
    MethodName  string
    Args        []byte
    Metadata    map[string]string
}

type RPCResponse struct {
    RequestID uint64
    Error     string
    Data      []byte
    Metadata  map[string]string
}
```

### 3.2 序列化

```go
// codec.go
type Codec interface {
    Encode(v interface{}) ([]byte, error)
    Decode(data []byte, v interface{}) error
    Name() string
}

type JSONCodec struct{}
type ProtobufCodec struct{}
type MsgpackCodec struct{}

// Codec 注册表
var codecMap = map[CodecType]Codec{
    JSON:     &JSONCodec{},
    Protobuf: &ProtobufCodec{},
    Msgpack:  &MsgpackCodec{},
}
```

---

## 4. Server 端设计

### 4.1 服务注册

```go
// service.go
type Service struct {
    Name    string
    rcvr    reflect.Value
    typ     reflect.Type
    methods map[string]*MethodType
}

type MethodType struct {
    method    reflect.Method
    ArgType   reflect.Type
    ReplyType reflect.Type
}

// 通过反射自动注册服务方法
// 方法签名约定: func (s *XXXService) Method(args *Args, reply *Reply) error
func (s *Service) Register(rcvr interface{}) error {
    // 遍历方法，提取满足约定的方法
}

func (s *Service) Call(methodName string, args, reply interface{}) error {
    // 反射调用
}
```

### 4.2 Server 核心

```go
// server.go
type Server struct {
    services    map[string]*Service
    listener    net.Listener
    middlewares []Middleware
    registry    Registry
    addr        string
}

func (s *Server) Register(rcvr interface{}) error
func (s *Server) Use(mw ...Middleware)
func (s *Server) Start(addr string) error
func (s *Server) Stop() error

func (s *Server) handleConnection(conn net.Conn) {
    // 1. 读取 Header
    // 2. 解码 Request
    // 3. 查找 Service + Method
    // 4. 执行中间件链
    // 5. 调用 Service.Call
    // 6. 编码 Response
    // 7. 写回
}
```

### 4.3 中间件

```go
// middleware.go
type HandlerFunc func(ctx *Context) error

type Middleware func(next HandlerFunc) HandlerFunc

type Context struct {
    Request   *RPCRequest
    Response  *RPCResponse
    Service   string
    Method    string
    Metadata  map[string]string
    StartTime time.Time
}

// 内置中间件
func LoggingMiddleware() Middleware     // 请求日志
func MetricsMiddleware() Middleware     // 耗时统计
func RecoveryMiddleware() Middleware    // panic 恢复
func RateLimitMiddleware(qps int) Middleware  // 限流
func AuthMiddleware(fn AuthFunc) Middleware   // 认证
```

---

## 5. Client 端设计

### 5.1 核心 Client

```go
// client.go
type Client struct {
    conn       net.Conn
    codec      Codec
    reqID      uint64
    pending    map[uint64]*Call   // 等待响应的请求
    mu         sync.Mutex
    closing    bool
}

type Call struct {
    ServiceMethod string
    Args          interface{}
    Reply         interface{}
    Error         error
    Done          chan *Call
}

func (c *Client) Call(service, method string, args, reply interface{}) error {
    // 同步调用
    call := c.Go(service, method, args, reply, make(chan *Call, 1))
    <-call.Done
    return call.Error
}

func (c *Client) Go(service, method string, args, reply interface{}, done chan *Call) *Call {
    // 异步调用
}
```

### 5.2 连接池

```go
// pool.go
type ConnPool struct {
    factory   func() (net.Conn, error)
    conns     chan net.Conn
    maxIdle   int
    maxActive int
    active    int32
    mu        sync.Mutex
}

func (p *ConnPool) Get() (net.Conn, error)
func (p *ConnPool) Put(conn net.Conn) error
func (p *ConnPool) Close() error
```

---

## 6. 服务发现

### 6.1 Registry 接口

```go
// registry.go
type ServiceInstance struct {
    ID       string
    Name     string
    Addr     string
    Port     int
    Metadata map[string]string
    Weight   int
}

type Registry interface {
    Register(instance *ServiceInstance) error
    Deregister(instance *ServiceInstance) error
    Discover(serviceName string) ([]*ServiceInstance, error)
    Watch(serviceName string) (<-chan []*ServiceInstance, error)
}
```

### 6.2 实现

```go
// 内存注册中心（开发 / 测试用）
type MemoryRegistry struct {
    services map[string][]*ServiceInstance
    watchers map[string][]chan []*ServiceInstance
    mu       sync.RWMutex
}

// etcd 注册中心（生产用，Phase 5 集成）
type EtcdRegistry struct {
    client     *clientv3.Client
    leaseID    clientv3.LeaseID
    leaseTTL   int64
}

// 服务端启动时注册，关闭时注销
// 客户端通过 Watch 监听变化，更新本地实例列表
```

---

## 7. 负载均衡

```go
// balancer.go
type Balancer interface {
    Pick(instances []*ServiceInstance) (*ServiceInstance, error)
}

// Round Robin
type RoundRobinBalancer struct {
    counter uint64
}

// Weighted Round Robin
type WeightedRRBalancer struct {
    weights map[string]int
    current map[string]int
}

// Random
type RandomBalancer struct{}

// Consistent Hash（适合有状态服务）
type ConsistentHashBalancer struct {
    ring     *ConsistentHashRing
    replicas int
}

// Least Connections
type LeastConnBalancer struct {
    conns map[string]int64
    mu    sync.RWMutex
}
```

---

## 8. 容错机制

### 8.1 Timeout

```go
// timeout.go
type TimeoutConfig struct {
    ConnectTimeout time.Duration  // 连接超时
    ReadTimeout    time.Duration  // 读超时
    WriteTimeout   time.Duration  // 写超时
    CallTimeout    time.Duration  // 整体调用超时
}

func CallWithTimeout(ctx context.Context, fn func() error, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    done := make(chan error, 1)
    go func() { done <- fn() }()
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ErrTimeout
    }
}
```

### 8.2 Retry

```go
// retry.go
type RetryPolicy struct {
    MaxRetries    int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
    RetryOn       func(err error) bool  // 哪些错误重试
}

func Retry(ctx context.Context, policy *RetryPolicy, fn func() error) error {
    var lastErr error
    delay := policy.InitialDelay
    for i := 0; i <= policy.MaxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
            if !policy.RetryOn(err) {
                return err
            }
        }
        time.Sleep(delay)
        delay = time.Duration(float64(delay) * policy.BackoffFactor)
        if delay > policy.MaxDelay {
            delay = policy.MaxDelay
        }
    }
    return lastErr
}
```

### 8.3 Circuit Breaker（熔断器）

```go
// circuit_breaker.go
type CircuitState int

const (
    StateClosed    CircuitState = 0  // 正常
    StateOpen      CircuitState = 1  // 熔断，快速失败
    StateHalfOpen  CircuitState = 2  // 半开，尝试恢复
)

type CircuitBreaker struct {
    state          CircuitState
    failureCount   int64
    successCount   int64
    threshold      int64         // 触发熔断的失败次数
    timeout        time.Duration // Open → HalfOpen 的等待时间
    lastFailTime   time.Time
    mu             sync.Mutex
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if !cb.AllowRequest() {
        return ErrCircuitOpen
    }
    err := fn()
    if err != nil {
        cb.RecordFailure()
    } else {
        cb.RecordSuccess()
    }
    return err
}

func (cb *CircuitBreaker) AllowRequest() bool
func (cb *CircuitBreaker) RecordFailure()
func (cb *CircuitBreaker) RecordSuccess()
```

---

## 9. 目录结构

```
mini-rpc/
├── cmd/
│   ├── server/
│   │   └── main.go              # 示例 Server
│   └── client/
│       └── main.go              # 示例 Client
├── internal/
│   ├── protocol/
│   │   ├── header.go            # 协议头
│   │   ├── message.go           # 请求 / 响应
│   │   └── codec.go             # 编解码器
│   ├── server/
│   │   ├── server.go            # RPC Server
│   │   ├── service.go           # 服务注册 + 反射调用
│   │   └── middleware.go        # 中间件
│   ├── client/
│   │   ├── client.go            # RPC Client
│   │   ├── pool.go              # 连接池
│   │   └── proxy.go             # 代理 / Stub 生成
│   ├── registry/
│   │   ├── registry.go          # Registry 接口
│   │   ├── memory.go            # 内存注册中心
│   │   └── etcd.go              # etcd 注册中心
│   ├── balancer/
│   │   ├── balancer.go          # Balancer 接口
│   │   ├── round_robin.go       # Round Robin
│   │   ├── weighted_rr.go       # Weighted Round Robin
│   │   ├── random.go            # Random
│   │   ├── consistent_hash.go   # 一致性哈希
│   │   └── least_conn.go        # 最少连接
│   ├── resilience/
│   │   ├── timeout.go           # 超时控制
│   │   ├── retry.go             # 重试策略
│   │   └── circuit_breaker.go   # 熔断器
│   └── transport/
│       ├── tcp.go               # TCP 传输
│       └── connection.go        # 连接管理
├── examples/
│   ├── helloworld/              # Hello World 示例
│   └── benchmark/               # 性能测试示例
├── test/
│   ├── server_test.go
│   ├── client_test.go
│   ├── balancer_test.go
│   ├── circuit_breaker_test.go
│   └── integration_test.go
├── go.mod
├── Makefile
└── README.md
```

---

## 10. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| 连接泄漏（连接池耗尽） | 连接池 metrics + goroutine 数 | pprof goroutine / `ss -tnp` |
| 网络超时级联 | 链路追踪 + 超时日志 | tracing / `tcpdump` |
| 负载不均衡 | 各实例 QPS 分布 | Prometheus + Grafana |
| 熔断器误触发 | 熔断状态监控 | 自定义 metrics (circuit_state) |
| 序列化性能瓶颈 | CPU 火焰图 | pprof CPU profile |
| 服务发现延迟 | Watch 事件时间戳 | etcd 日志 / 服务列表快照对比 |

---

## 11. 面试考点

### 必问级

- [ ] **RPC vs HTTP：** 为什么微服务间用 RPC 而不用 HTTP？（性能、协议、流式传输）
- [ ] **服务发现：** 注册中心挂了怎么办？健康检查如何工作？
- [ ] **负载均衡：** Round Robin vs Weighted RR vs 一致性哈希的适用场景
- [ ] **熔断器：** Closed → Open → HalfOpen 三态转换逻辑？与降级的区别？

### 高频级

- [ ] **重试策略：** 指数退避 + 最大重试次数？哪些错误可重试（幂等问题）？
- [ ] **连接池：** 为什么需要？最大连接数 / 空闲连接数如何配置？
- [ ] **反射 vs 代码生成：** 各自优缺点？gRPC 用的哪种？
- [ ] **协议设计：** 为什么用 Length-Prefixed？粘包拆包怎么处理？

---

## 12. 实现里程碑

### M1: 协议 + 基础通信（第1-2周）
- [ ] 定义二进制协议格式
- [ ] 实现 JSON / Msgpack 编解码
- [ ] 实现 TCP Server / Client 基础通信
- [ ] 测试：客户端调用服务端方法并获得响应

### M2: 服务注册 + 反射调用（第3周）
- [ ] 通过反射自动发现服务方法
- [ ] 实现 Service.Register + Service.Call
- [ ] 支持同步 / 异步调用
- [ ] 连接池

### M3: 服务发现 + 负载均衡（第4周）
- [ ] Memory Registry
- [ ] Round Robin / Weighted RR / Random 负载均衡
- [ ] 一致性哈希
- [ ] Watch 机制（服务实例变化通知）

### M4: 容错机制（第5周）
- [ ] Timeout（连接 / 读写 / 调用）
- [ ] Retry + Exponential Backoff
- [ ] Circuit Breaker（Closed → Open → HalfOpen）
- [ ] 测试：模拟网络延迟 / 服务宕机场景

### M5: 中间件 + 完善（第6周）
- [ ] 中间件链
- [ ] Logging / Metrics / Recovery 中间件
- [ ] RateLimit 中间件
- [ ] 端到端压测报告

---

## 11. 性能目标

| 指标 | 目标值 |
|------|--------|
| 单连接 QPS | > 10,000 |
| P99 延迟（本地） | < 1ms |
| P99 延迟（跨机器） | < 5ms |
| 连接池复用率 | > 95% |
| 序列化 1KB 消息 | < 50μs |

---

## 12. 关键学习产出

- **网络不可靠性：** 理解为什么 RPC ≠ 本地调用
- **服务发现本质：** 注册 / 发现 / 健康检查的完整链路
- **负载均衡策略：** 不同算法适用的场景
- **容错思维：** Timeout → Retry → Circuit Breaker 的递进防御
- **反射与代码生成：** 框架如何做到用户无感知

---

## 13. 毕业标准

- [ ] 能通过简单注解注册服务并被客户端发现
- [ ] 3 实例下负载均衡分配均匀
- [ ] 模拟一个实例宕机后，自动切换到其他实例
- [ ] Circuit Breaker 在持续失败后正确熔断
- [ ] 完成压测报告
- [ ] 能清晰描述一次 RPC 调用的完整生命周期
