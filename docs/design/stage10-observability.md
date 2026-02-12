# Stage 10 — 可观测性体系: observability-stack

> 预计周期：3 周 | 语言：Golang | 难度：⭐⭐⭐⭐
>
> 架构师必备能力：没有可观测性，一切优化都是盲人摸象

---

## 1. 项目目标

构建完整的可观测性体系，覆盖 Metrics / Logging / Tracing 三大支柱：

- **Metrics（指标）：** Prometheus 采集 + Grafana 可视化 + Alertmanager 告警
- **Logging（日志）：** 结构化日志 + Loki 聚合 + 日志查询
- **Tracing（链路追踪）：** OpenTelemetry 分布式追踪 + Jaeger 可视化
- **综合实践：** 将可观测性集成到之前所有项目中

---

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Go 应用（业务服务）                      │
│                                                          │
│  ┌────────────┐  ┌────────────┐  ┌─────────────────┐   │
│  │ Prometheus │  │   slog /   │  │ OpenTelemetry   │   │
│  │  Client    │  │   Zap      │  │   SDK           │   │
│  │ (metrics)  │  │ (logging)  │  │ (tracing)       │   │
│  └─────┬──────┘  └─────┬──────┘  └──────┬──────────┘   │
└────────┼───────────────┼────────────────┼───────────────┘
         │               │                │
    ┌────▼────┐    ┌─────▼─────┐    ┌─────▼──────────┐
    │Promethe-│    │  Promtail │    │  OTel          │
    │us Server│    │  / Loki   │    │  Collector     │
    │ (pull)  │    │  Driver   │    │                │
    └────┬────┘    └─────┬─────┘    └──────┬─────────┘
         │               │                 │
    ┌────▼────┐    ┌─────▼─────┐    ┌──────▼─────────┐
    │ Grafana │    │   Loki    │    │    Jaeger      │
    │(可视化)  │    │(日志存储)  │    │  (追踪可视化)   │
    └────┬────┘    └───────────┘    └────────────────┘
         │
    ┌────▼────────┐
    │Alertmanager │
    │  (告警)      │
    └─────────────┘
```

---

## 3. Metrics — Prometheus + Grafana

### 3.1 指标体系设计

```go
// metrics.go

// RED 指标（面向服务 — 每个 API 都要有）
// Rate:     请求速率 (QPS)
// Error:    错误率
// Duration: 延迟分布

// USE 指标（面向资源 — CPU / Memory / Disk / Network）
// Utilization: 使用率
// Saturation:  饱和度（队列长度）
// Errors:      错误数

type Metrics struct {
    // HTTP 请求指标
    HTTPRequestTotal    *prometheus.CounterVec   // 总请求数 (method, path, status)
    HTTPRequestDuration *prometheus.HistogramVec  // 请求延迟 (method, path)
    HTTPActiveRequests  prometheus.Gauge          // 当前活跃请求数

    // 业务指标
    BusinessOperations  *prometheus.CounterVec   // 业务操作计数 (operation, result)

    // 资源指标
    DBConnectionsActive prometheus.Gauge          // 数据库活跃连接数
    DBConnectionsIdle   prometheus.Gauge          // 空闲连接数
    CacheHitRate        *prometheus.GaugeVec      // 缓存命中率 (cache_name)
    GoroutineCount      prometheus.Gauge          // goroutine 数量
}

func NewMetrics(namespace string) *Metrics {
    m := &Metrics{
        HTTPRequestTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: namespace,
                Name:      "http_requests_total",
                Help:      "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        ),
        HTTPRequestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: namespace,
                Name:      "http_request_duration_seconds",
                Help:      "HTTP request duration in seconds",
                Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
            },
            []string{"method", "path"},
        ),
        // ... 其他指标
    }
    // 注册到 prometheus
    prometheus.MustRegister(m.HTTPRequestTotal, m.HTTPRequestDuration)
    return m
}
```

### 3.2 Metrics 中间件

```go
// metrics_middleware.go
func MetricsMiddleware(m *Metrics) gin.HandlerFunc {
    return func(c *gin.Context) {
        m.HTTPActiveRequests.Inc()
        defer m.HTTPActiveRequests.Dec()

        start := time.Now()
        c.Next()
        duration := time.Since(start).Seconds()

        status := strconv.Itoa(c.Writer.Status())
        path := c.FullPath() // 使用路由模板而非实际路径，避免基数爆炸

        m.HTTPRequestTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        m.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    }
}
```

### 3.3 Grafana Dashboard 设计

```
Dashboard: Service Overview
├── Row 1: 流量概览
│   ├── Panel: QPS (rate(http_requests_total[5m]))
│   ├── Panel: Error Rate (rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]))
│   └── Panel: P99 Latency (histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])))
│
├── Row 2: 资源状态
│   ├── Panel: Goroutine Count
│   ├── Panel: DB Connections (active / idle)
│   ├── Panel: Cache Hit Rate
│   └── Panel: Memory Usage
│
└── Row 3: 业务指标
    ├── Panel: 注册数
    ├── Panel: 登录数
    └── Panel: 错误类型分布
```

### 3.4 告警规则

```yaml
# alerting_rules.yml
groups:
  - name: service-alerts
    rules:
      # 错误率 > 5% 持续 5 分钟
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

      # P99 延迟 > 1s 持续 5 分钟
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning

      # Goroutine 数量异常增长（可能泄漏）
      - alert: GoroutineLeak
        expr: go_goroutines > 10000
        for: 10m
        labels:
          severity: warning
```

---

## 4. Logging — 结构化日志 + Loki

### 4.1 结构化日志规范

```go
// logger.go
// 使用 slog（Go 1.21+ 标准库）

func NewLogger(level string, format string) *slog.Logger {
    var handler slog.Handler
    opts := &slog.HandlerOptions{
        Level: parseLevel(level),
    }

    switch format {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, opts)
    default:
        handler = slog.NewTextHandler(os.Stdout, opts)
    }

    return slog.New(handler)
}

// 日志规范：
// 1. 必须包含 request_id（链路追踪关联）
// 2. 必须包含 trace_id / span_id（与 tracing 关联）
// 3. 错误日志必须包含 stack trace
// 4. 禁止打印敏感信息（密码、token）
// 5. 使用结构化字段，禁止字符串拼接

// 示例
logger.Info("user logged in",
    "request_id", requestID,
    "trace_id", traceID,
    "user_id", userID,
    "ip", clientIP,
    "duration_ms", duration,
)

logger.Error("database query failed",
    "request_id", requestID,
    "error", err,
    "query", "SELECT * FROM users WHERE id = ?",
    "args", []interface{}{userID},
)
```

### 4.2 Loki 日志查询

```
# LogQL 查询示例

# 查看某个 request_id 的所有日志
{app="user-service"} |= "request_id=abc123"

# 查看错误日志
{app="user-service", level="error"}

# 统计每分钟错误数
count_over_time({app="user-service", level="error"}[1m])

# 查看慢请求（> 1000ms）
{app="user-service"} | json | duration_ms > 1000
```

---

## 5. Tracing — OpenTelemetry + Jaeger

### 5.1 OpenTelemetry 集成

```go
// tracing.go
func InitTracer(serviceName, endpoint string) (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
        trace.WithSampler(trace.AlwaysSample()), // 开发环境全采样
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, nil
}
```

### 5.2 Tracing 中间件

```go
// tracing_middleware.go
func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)

    return func(c *gin.Context) {
        // 从请求头提取上游 trace context
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        // 创建 span
        ctx, span := tracer.Start(ctx, c.FullPath(),
            oteltrace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
            ),
        )
        defer span.End()

        // 将 trace_id 放入 context
        c.Request = c.Request.WithContext(ctx)
        traceID := span.SpanContext().TraceID().String()
        c.Set("trace_id", traceID)
        c.Header("X-Trace-ID", traceID)

        c.Next()

        // 记录响应状态
        span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
        if c.Writer.Status() >= 500 {
            span.SetStatus(codes.Error, "server error")
        }
    }
}
```

### 5.3 跨服务传播

```go
// propagation.go
// 发起 RPC / HTTP 调用时，将 trace context 注入到请求头

func InjectTraceContext(ctx context.Context, req *http.Request) {
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// 调用链示例：
// Gateway (span-1) → UserService (span-2) → MySQL (span-3)
//     │                   │                     │
//     └── trace_id: abc ──┴── trace_id: abc ────┘
//         span_id: 001       span_id: 002        span_id: 003
//         parent: none       parent: 001         parent: 002
```

---

## 6. 错误预算（SLI / SLO / SLA）

### 概念

```
SLI (Service Level Indicator) — 服务质量指标
    例：过去 5 分钟 P99 延迟 < 200ms 的请求比例

SLO (Service Level Objective) — 服务质量目标
    例：SLI >= 99.9%（每月允许 43 分钟不达标）

SLA (Service Level Agreement) — 服务等级协议
    例：SLO 未达标时的赔偿条款

Error Budget = 1 - SLO
    例：SLO = 99.9%，Error Budget = 0.1% = 每月 43 分钟
```

### 实践

```go
// error_budget.go
type ErrorBudget struct {
    SLO            float64       // 目标 SLO（如 0.999）
    WindowDuration time.Duration // 计算窗口（如 30 天）
    totalRequests  int64
    failedRequests int64
}

// 当前 SLI
func (eb *ErrorBudget) CurrentSLI() float64 {
    if eb.totalRequests == 0 {
        return 1.0
    }
    return 1.0 - float64(eb.failedRequests)/float64(eb.totalRequests)
}

// 剩余错误预算百分比
func (eb *ErrorBudget) RemainingBudget() float64 {
    budget := 1.0 - eb.SLO // 总预算
    consumed := 1.0 - eb.CurrentSLI()
    return (budget - consumed) / budget * 100
}

// 是否还有错误预算
func (eb *ErrorBudget) HasBudget() bool {
    return eb.CurrentSLI() >= eb.SLO
}
```

---

## 7. 目录结构

```
projects/stage10-observability/
├── cmd/
│   └── demo-service/
│       └── main.go                # 演示服务
├── internal/
│   ├── metrics/
│   │   ├── metrics.go             # Metrics 定义
│   │   ├── middleware.go          # Metrics 中间件
│   │   └── collector.go          # 自定义 Collector
│   ├── logging/
│   │   ├── logger.go             # 结构化日志
│   │   └── middleware.go         # 日志中间件
│   ├── tracing/
│   │   ├── tracer.go             # OpenTelemetry 初始化
│   │   ├── middleware.go         # Tracing 中间件
│   │   └── propagation.go       # 跨服务传播
│   └── slo/
│       ├── error_budget.go       # 错误预算
│       └── slo_test.go
├── configs/
│   ├── prometheus.yml            # Prometheus 配置
│   ├── alerting_rules.yml        # 告警规则
│   ├── grafana/
│   │   └── dashboards/
│   │       ├── service-overview.json
│   │       └── resource-usage.json
│   ├── loki-config.yml           # Loki 配置
│   └── otel-collector.yml        # OTel Collector 配置
├── docker-compose.yml            # 完整可观测性栈
├── test/
│   ├── metrics_test.go
│   ├── tracing_test.go
│   └── integration_test.go
├── go.mod
├── Makefile
└── README.md
```

---

## 8. Docker 部署

```yaml
# docker-compose.yml
version: "3.8"

services:
  # 演示服务
  demo-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317

  # Prometheus
  prometheus:
    image: prom/prometheus:v2.51.0
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./configs/alerting_rules.yml:/etc/prometheus/alerting_rules.yml

  # Grafana
  grafana:
    image: grafana/grafana:10.4.0
    ports:
      - "3000:3000"
    volumes:
      - ./configs/grafana/dashboards:/var/lib/grafana/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

  # Loki（日志）
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    volumes:
      - ./configs/loki-config.yml:/etc/loki/local-config.yaml

  # Jaeger（链路追踪）
  jaeger:
    image: jaegertracing/all-in-one:1.55
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # HTTP collector

  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector:0.96.0
    ports:
      - "4317:4317"   # gRPC receiver
      - "4318:4318"   # HTTP receiver
    volumes:
      - ./configs/otel-collector.yml:/etc/otelcol/config.yaml

  # Alertmanager
  alertmanager:
    image: prom/alertmanager:v0.27.0
    ports:
      - "9093:9093"
```

---

## 9. 实现里程碑

### M1: Metrics + Grafana（第1周）
- [ ] Prometheus Client 集成
- [ ] RED 指标实现
- [ ] Metrics 中间件
- [ ] Prometheus + Grafana Docker 部署
- [ ] 创建 Service Overview Dashboard
- [ ] 告警规则配置

### M2: Logging + Loki（第2周）
- [ ] 结构化日志规范实现
- [ ] 日志中间件（关联 request_id + trace_id）
- [ ] Loki 部署 + 日志采集
- [ ] LogQL 查询实践

### M3: Tracing + 综合（第3周）
- [ ] OpenTelemetry SDK 集成
- [ ] Tracing 中间件
- [ ] 跨服务 trace 传播
- [ ] Jaeger 部署 + 链路可视化
- [ ] 错误预算实践
- [ ] 将可观测性集成回 Stage 3-9 项目

---

## 10. 毕业标准

- [ ] Grafana Dashboard 展示 RED + USE 指标
- [ ] 告警规则能正确触发（模拟错误率飙高）
- [ ] 日志包含 request_id + trace_id，可关联查询
- [ ] 跨服务调用链在 Jaeger 中可视化
- [ ] 能解释 Metrics / Logging / Tracing 各自解决什么问题
- [ ] 能计算错误预算并判断是否可以发布新版本
