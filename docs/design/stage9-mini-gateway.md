# Stage 9 — API Gateway: mini-gateway

> 预计周期：1.5 个月 | 语言：Golang | 难度：⭐⭐⭐⭐
>
> 进入架构师领域：网关是系统的流量入口，所有架构决策在此汇聚

---

## 1. 项目目标

从零构建一个高性能 API 网关，掌握流量治理的核心能力：

- **反向代理：** 请求路由 + 转发
- **限流：** 保护后端服务不被洪峰打垮
- **熔断：** 快速失败，避免级联故障
- **认证鉴权：** JWT / API Key 验证
- **可观测性：** Metrics / Tracing / Logging

**本质理解：** 网关是系统的流量入口，所有架构决策在此汇聚——路由、安全、流控、可观测。

---

## 2. 整体架构

```
┌─────────────────────────────────────────────────────┐
│                     Clients                          │
│              (HTTP / gRPC / WebSocket)                │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                  mini-gateway                        │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │              Filter Chain (Pipeline)            │  │
│  │                                                 │  │
│  │  ┌─────┐ ┌──────┐ ┌─────┐ ┌──────┐ ┌───────┐ │  │
│  │  │Rate │→│Auth  │→│CORS │→│Retry │→│Logging│ │  │
│  │  │Limit│ │      │ │     │ │      │ │       │ │  │
│  │  └─────┘ └──────┘ └─────┘ └──────┘ └───────┘ │  │
│  └──────────────────┬─────────────────────────────┘  │
│                     │                                │
│  ┌──────────────────▼─────────────────────────────┐  │
│  │              Router                             │  │
│  │  /api/v1/users → user-service                  │  │
│  │  /api/v1/orders → order-service                │  │
│  │  /api/v1/products → product-service            │  │
│  └──────────────────┬─────────────────────────────┘  │
│                     │                                │
│  ┌──────────────────▼─────────────────────────────┐  │
│  │          Load Balancer                          │  │
│  │     (RoundRobin / WeightedRR / Random)         │  │
│  └──────────────────┬─────────────────────────────┘  │
│                     │                                │
│  ┌──────────────────▼─────────────────────────────┐  │
│  │          Connection Pool                        │  │
│  │     (HTTP Client / gRPC Client)                │  │
│  └──────────────────┬─────────────────────────────┘  │
└──────────────────────┼──────────────────────────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
    ┌────▼────┐  ┌─────▼────┐  ┌────▼─────┐
    │ Service │  │ Service  │  │ Service  │
    │    A    │  │    B     │  │    C     │
    └─────────┘  └──────────┘  └──────────┘
```

---

## 3. 配置驱动设计

### 3.1 路由配置

```yaml
# config.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

routes:
  - name: user-service
    prefix: /api/v1/users
    strip_prefix: true
    targets:
      - addr: http://127.0.0.1:8081
        weight: 3
      - addr: http://127.0.0.1:8082
        weight: 1
    filters:
      - rate_limit:
          qps: 1000
          burst: 2000
      - auth:
          type: jwt
          secret: ${JWT_SECRET}
      - circuit_breaker:
          threshold: 5
          timeout: 30s
      - retry:
          max_retries: 2
          retry_on: [502, 503, 504]

  - name: order-service
    prefix: /api/v1/orders
    methods: [GET, POST]
    targets:
      - addr: http://127.0.0.1:8083
    filters:
      - rate_limit:
          qps: 500

global_filters:
  - cors:
      allow_origins: ["*"]
      allow_methods: [GET, POST, PUT, DELETE]
  - logging:
      format: json
  - metrics:
      enabled: true
```

### 3.2 配置模型

```go
// config.go
type GatewayConfig struct {
    Server  ServerConfig  `yaml:"server"`
    Routes  []RouteConfig `yaml:"routes"`
    Global  GlobalConfig  `yaml:"global_filters"`
}

type ServerConfig struct {
    Port         int           `yaml:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
}

type RouteConfig struct {
    Name        string         `yaml:"name"`
    Prefix      string         `yaml:"prefix"`
    StripPrefix bool           `yaml:"strip_prefix"`
    Methods     []string       `yaml:"methods"`
    Targets     []TargetConfig `yaml:"targets"`
    Filters     []FilterConfig `yaml:"filters"`
}

type TargetConfig struct {
    Addr   string `yaml:"addr"`
    Weight int    `yaml:"weight"`
}

type FilterConfig map[string]interface{}
```

---

## 4. 核心模块设计

### 4.1 Router（路由引擎）

```go
// router.go
type Router struct {
    routes []*Route
    tree   *RadixTree  // 前缀树，高效路由匹配
}

type Route struct {
    Name        string
    Prefix      string
    Methods     map[string]bool
    StripPrefix bool
    Targets     []*Target
    Balancer    Balancer
    FilterChain *FilterChain
}

type Target struct {
    Addr   string
    Weight int
    Alive  bool
    mu     sync.RWMutex
}

// Radix Tree 实现高效前缀匹配
type RadixTree struct {
    root *radixNode
}

type radixNode struct {
    prefix   string
    route    *Route
    children []*radixNode
}

func (r *Router) Match(method, path string) (*Route, map[string]string)
func (r *Router) AddRoute(route *Route) error
func (r *Router) RemoveRoute(name string) error
```

### 4.2 Filter Chain（过滤器链 / 插件系统）

```go
// filter.go
type Filter interface {
    Name() string
    Order() int
    Handle(ctx *RequestContext, next FilterFunc) error
}

type FilterFunc func(ctx *RequestContext) error

type FilterChain struct {
    filters []Filter
}

func (fc *FilterChain) Execute(ctx *RequestContext) error {
    index := 0
    var next FilterFunc
    next = func(ctx *RequestContext) error {
        if index >= len(fc.filters) {
            return nil // 到达链尾
        }
        f := fc.filters[index]
        index++
        return f.Handle(ctx, next)
    }
    return next(ctx)
}

type RequestContext struct {
    Request      *http.Request
    Response     http.ResponseWriter
    Route        *Route
    Target       *Target
    StartTime    time.Time
    Metadata     map[string]interface{}
    StatusCode   int
    ResponseBody []byte
    Aborted      bool
}
```

### 4.3 Rate Limiter（限流器）

```go
// rate_limiter.go

// === Token Bucket（令牌桶）===
type TokenBucket struct {
    rate       float64     // 每秒产生令牌数
    burst      int         // 桶容量
    tokens     float64     // 当前令牌数
    lastTime   time.Time
    mu         sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(tb.lastTime).Seconds()
    tb.tokens += elapsed * tb.rate
    if tb.tokens > float64(tb.burst) {
        tb.tokens = float64(tb.burst)
    }
    tb.lastTime = now

    if tb.tokens >= 1 {
        tb.tokens--
        return true
    }
    return false
}

// === Sliding Window（滑动窗口）===
type SlidingWindowLimiter struct {
    windowSize time.Duration
    limit      int
    windows    map[int64]int  // timestamp_bucket → count
    mu         sync.Mutex
}

func (sw *SlidingWindowLimiter) Allow() bool

// === Rate Limiter Filter ===
type RateLimitFilter struct {
    limiters map[string]*TokenBucket  // per-route limiter
    global   *TokenBucket             // global limiter
}

func (f *RateLimitFilter) Handle(ctx *RequestContext, next FilterFunc) error {
    if !f.limiters[ctx.Route.Name].Allow() {
        ctx.StatusCode = http.StatusTooManyRequests
        ctx.Aborted = true
        return ErrRateLimited
    }
    return next(ctx)
}
```

### 4.4 Circuit Breaker（熔断器）

```go
// circuit_breaker_filter.go
type CircuitBreakerFilter struct {
    breakers map[string]*CircuitBreaker  // per-route
}

type CircuitBreaker struct {
    state          CircuitState
    failCount      int64
    successCount   int64
    threshold      int64
    timeout        time.Duration
    lastFailTime   time.Time
    halfOpenMax    int           // 半开状态允许的最大请求数
    halfOpenCount  int64
    mu             sync.Mutex
}

type CircuitState int

const (
    Closed   CircuitState = 0  // 正常通行
    Open     CircuitState = 1  // 熔断，快速失败
    HalfOpen CircuitState = 2  // 试探性放行
)

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case Closed:
        return true
    case Open:
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.state = HalfOpen
            cb.halfOpenCount = 0
            return true
        }
        return false
    case HalfOpen:
        return atomic.AddInt64(&cb.halfOpenCount, 1) <= int64(cb.halfOpenMax)
    }
    return false
}

func (cb *CircuitBreaker) RecordSuccess()
func (cb *CircuitBreaker) RecordFailure()
```

### 4.5 Auth Filter（认证）

```go
// auth.go
type AuthFilter struct {
    Type   string // "jwt" / "apikey"
    Config AuthConfig
}

type AuthConfig struct {
    JWTSecret    string
    APIKeys      map[string]bool
    ExcludePaths []string
}

func (f *AuthFilter) Handle(ctx *RequestContext, next FilterFunc) error {
    // 跳过排除路径
    for _, p := range f.Config.ExcludePaths {
        if strings.HasPrefix(ctx.Request.URL.Path, p) {
            return next(ctx)
        }
    }

    switch f.Type {
    case "jwt":
        return f.validateJWT(ctx, next)
    case "apikey":
        return f.validateAPIKey(ctx, next)
    }
    return next(ctx)
}

func (f *AuthFilter) validateJWT(ctx *RequestContext, next FilterFunc) error {
    token := ctx.Request.Header.Get("Authorization")
    // Bearer token 解析
    // JWT 签名验证
    // 提取 claims 放入 ctx.Metadata
    return next(ctx)
}
```

### 4.6 反向代理 + 负载均衡

```go
// proxy.go
type ReverseProxy struct {
    transport http.RoundTripper
    pool      *ConnectionPool
}

func (p *ReverseProxy) Forward(ctx *RequestContext) error {
    target := ctx.Route.Balancer.Pick(ctx.Route.Targets)
    if target == nil {
        return ErrNoAvailableTarget
    }
    ctx.Target = target

    // 构造转发请求
    targetURL := target.Addr + ctx.Request.URL.Path
    if ctx.Route.StripPrefix {
        targetURL = target.Addr + strings.TrimPrefix(ctx.Request.URL.Path, ctx.Route.Prefix)
    }

    proxyReq, _ := http.NewRequest(ctx.Request.Method, targetURL, ctx.Request.Body)
    copyHeaders(proxyReq.Header, ctx.Request.Header)

    // 添加网关头
    proxyReq.Header.Set("X-Forwarded-For", ctx.Request.RemoteAddr)
    proxyReq.Header.Set("X-Request-ID", ctx.Metadata["request_id"].(string))

    resp, err := p.transport.RoundTrip(proxyReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 写回响应
    copyHeaders(ctx.Response.Header(), resp.Header)
    ctx.StatusCode = resp.StatusCode
    ctx.Response.WriteHeader(resp.StatusCode)
    io.Copy(ctx.Response, resp.Body)

    return nil
}

// balancer.go
type Balancer interface {
    Pick(targets []*Target) *Target
}

type RoundRobinBalancer struct{ counter uint64 }
type WeightedRRBalancer struct{}
type RandomBalancer struct{}
```

---

## 5. 可观测性

### 5.1 Metrics

```go
// metrics.go
type Metrics struct {
    RequestTotal    *Counter    // 总请求数 (by route, method, status)
    RequestDuration *Histogram  // 请求延迟 (by route)
    ActiveRequests  *Gauge      // 当前活跃请求数
    TargetHealth    *Gauge      // 后端健康状态
    CircuitState    *Gauge      // 熔断器状态
    RateLimitHits   *Counter    // 限流触发次数
}

type MetricsFilter struct {
    metrics *Metrics
}

func (f *MetricsFilter) Handle(ctx *RequestContext, next FilterFunc) error {
    f.metrics.ActiveRequests.Inc()
    defer f.metrics.ActiveRequests.Dec()

    start := time.Now()
    err := next(ctx)
    duration := time.Since(start)

    f.metrics.RequestTotal.Inc(ctx.Route.Name, ctx.Request.Method, ctx.StatusCode)
    f.metrics.RequestDuration.Observe(ctx.Route.Name, duration.Seconds())

    return err
}

// 暴露 /metrics 端点（Prometheus 格式）
func (m *Metrics) Handler() http.Handler
```

### 5.2 Access Log

```go
// logging.go
type AccessLog struct {
    Timestamp  time.Time `json:"timestamp"`
    Method     string    `json:"method"`
    Path       string    `json:"path"`
    StatusCode int       `json:"status_code"`
    Duration   float64   `json:"duration_ms"`
    ClientIP   string    `json:"client_ip"`
    Target     string    `json:"target"`
    RequestID  string    `json:"request_id"`
    UserAgent  string    `json:"user_agent"`
    Error      string    `json:"error,omitempty"`
}

type LoggingFilter struct {
    logger *slog.Logger
    format string  // "json" / "text"
}
```

### 5.3 Health Check

```go
// health.go
type HealthChecker struct {
    interval time.Duration
    timeout  time.Duration
    targets  []*Target
}

func (hc *HealthChecker) Start() {
    ticker := time.NewTicker(hc.interval)
    for range ticker.C {
        for _, target := range hc.targets {
            go hc.check(target)
        }
    }
}

func (hc *HealthChecker) check(target *Target) {
    ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", target.Addr+"/health", nil)
    resp, err := http.DefaultClient.Do(req)

    target.mu.Lock()
    defer target.mu.Unlock()
    target.Alive = err == nil && resp.StatusCode == 200
}
```

---

## 6. 热加载配置

```go
// hot_reload.go
type ConfigWatcher struct {
    configPath string
    gateway    *Gateway
    watcher    *fsnotify.Watcher
}

func (cw *ConfigWatcher) Watch() {
    for {
        select {
        case event := <-cw.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                newConfig, err := LoadConfig(cw.configPath)
                if err != nil {
                    log.Error("config reload failed", "err", err)
                    continue
                }
                cw.gateway.Reload(newConfig)
                log.Info("config reloaded successfully")
            }
        }
    }
}
```

---

## 7. 目录结构

```
mini-gateway/
├── cmd/
│   └── gateway/
│       └── main.go              # 入口
├── internal/
│   ├── gateway/
│   │   ├── gateway.go           # Gateway 核心
│   │   └── config.go            # 配置定义 & 加载
│   ├── router/
│   │   ├── router.go            # 路由引擎
│   │   └── radix_tree.go        # 前缀树
│   ├── filter/
│   │   ├── filter.go            # Filter 接口 & Chain
│   │   ├── rate_limit.go        # 限流
│   │   ├── circuit_breaker.go   # 熔断
│   │   ├── auth.go              # 认证
│   │   ├── cors.go              # CORS
│   │   ├── retry.go             # 重试
│   │   ├── logging.go           # 日志
│   │   ├── metrics.go           # 指标
│   │   └── request_id.go        # 请求 ID 注入
│   ├── proxy/
│   │   ├── reverse_proxy.go     # 反向代理
│   │   └── conn_pool.go         # 连接池
│   ├── balancer/
│   │   ├── balancer.go          # Balancer 接口
│   │   ├── round_robin.go
│   │   ├── weighted_rr.go
│   │   └── random.go
│   ├── health/
│   │   └── checker.go           # 健康检查
│   └── reload/
│       └── hot_reload.go        # 热加载
├── configs/
│   ├── gateway.yaml             # 默认配置
│   └── gateway_dev.yaml         # 开发配置
├── test/
│   ├── router_test.go
│   ├── rate_limit_test.go
│   ├── circuit_breaker_test.go
│   ├── auth_test.go
│   ├── proxy_test.go
│   └── integration_test.go
├── bench/
│   └── gateway_bench_test.go
├── go.mod
├── Makefile
└── README.md
```

---

## 8. 构造问题 & 定位手段

| 构造问题 | 定位手段 | 排查命令 / 工具 |
|---------|---------|----------------|
| 流量洪峰打满后端 | 限流 metrics + QPS 监控 | Prometheus rate_limit_hits 指标 |
| 级联故障（后端崩溃蔓延） | 熔断器状态 + 链路追踪 | circuit_state metrics / tracing |
| 配置热加载失败 | 配置文件校验日志 | fsnotify 事件日志 |
| JWT 验证性能瓶颈 | CPU 火焰图 | pprof CPU profile |
| 负载不均衡 | 后端实例 QPS 分布 | Prometheus target 指标 |
| 健康检查误摘除 | Health Check 日志 + 目标状态 | target_health metrics |

---

## 9. 面试考点

### 必问级

- [ ] **网关的作用：** 为什么需要网关？和 Nginx 的区别？
- [ ] **限流算法：** 令牌桶 vs 漏桶 vs 滑动窗口的适用场景
- [ ] **熔断降级：** 什么时候熔断？什么时候降级？区别是什么？
- [ ] **配置热加载：** 如何实现不停机更新路由规则？

### 高频级

- [ ] **Filter Chain 模式：** 设计原理？与中间件模式的关系？
- [ ] **Radix Tree 路由：** 为什么用前缀树而不用 HashMap？
- [ ] **反向代理：** X-Forwarded-For / X-Real-IP 的作用？
- [ ] **健康检查：** 主动探活 vs 被动检测的区别？

---

## 10. 实现里程碑

### M1: 反向代理 + 路由（第1-2周）
- [ ] 配置文件解析
- [ ] Radix Tree 路由匹配
- [ ] HTTP 反向代理转发
- [ ] Strip Prefix / 请求头复制
- [ ] 测试：配置路由，成功转发到后端

### M2: 负载均衡 + 健康检查（第3周）
- [ ] Round Robin / Weighted RR / Random
- [ ] 后端 Health Check（定时探活）
- [ ] 不健康节点自动摘除
- [ ] 测试：后端实例宕机后自动切换

### M3: 限流 + 熔断（第4周）
- [ ] Token Bucket 限流器
- [ ] Sliding Window 限流器
- [ ] Circuit Breaker（三态转换）
- [ ] 测试：超 QPS 返回 429；后端持续失败触发熔断

### M4: 认证 + CORS + Filter Chain（第5周）
- [ ] JWT 验证
- [ ] API Key 验证
- [ ] CORS 处理
- [ ] Filter Chain 组装 & 执行
- [ ] 测试：无 token 返回 401

### M5: 可观测性（第6周）
- [ ] Prometheus Metrics 暴露
- [ ] JSON Access Log
- [ ] Request ID 链路追踪
- [ ] /metrics + /health 端点

### M6: 热加载 + 压测（第7周）
- [ ] 配置文件热加载
- [ ] 使用 wrk / k6 压测
- [ ] 输出性能报告
- [ ] 编写使用文档

---

## 9. 性能目标

| 指标 | 目标值 |
|------|--------|
| 代理转发 QPS（单核） | > 20,000 |
| 代理附加延迟（P99） | < 2ms |
| 限流判断耗时 | < 1μs |
| 熔断判断耗时 | < 1μs |
| 配置热加载时间 | < 100ms |
| 并发连接数 | > 10,000 |

---

## 10. 关键学习产出

- **流量治理全局观：** 路由 → 限流 → 熔断 → 认证 → 转发，完整链路
- **限流算法对比：** Token Bucket vs Sliding Window vs Leaky Bucket 的适用场景
- **熔断思维：** 为什么快速失败优于慢失败
- **可观测性三支柱：** Metrics / Logging / Tracing 的协同
- **插件化架构：** Filter Chain 模式实现功能可插拔

---

## 11. 毕业标准

- [ ] 能正确路由多个后端服务
- [ ] 限流在超 QPS 时返回 429
- [ ] 熔断在后端持续失败时触发，恢复后自动放行
- [ ] JWT 认证能正确拦截未授权请求
- [ ] Metrics 端点可被 Prometheus 采集
- [ ] 配置热加载不中断服务
- [ ] 完成压测报告，清楚网关性能极限
- [ ] 能画出请求从进入网关到到达后端的完整流转路径
