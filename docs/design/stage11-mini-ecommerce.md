# Stage 11 — 终极实战: Mini 电商平台

> 预计周期：2 个月 | 语言：Golang | 难度：⭐⭐⭐⭐⭐
>
> 将前面所有能力整合为一个架构级系统

---

## 1. 项目目标

构建一个包含多个微服务的 Mini 电商平台，综合运用前 10 个 Stage 的所有技术：

- **微服务架构：** 多服务拆分 + 独立部署
- **分布式事务：** 下单 → 扣库存 → 支付的跨服务一致性
- **缓存策略：** 商品缓存 + 库存预热 + 一致性保障
- **流量治理：** 秒杀场景的限流 + 削峰 + 降级
- **可观测性：** 全链路 Metrics + Logging + Tracing

---

## 2. 微服务架构

```
                        ┌─────────────────┐
                        │    Client       │
                        └────────┬────────┘
                                 │ HTTP
                        ┌────────▼────────┐
                        │  API Gateway    │  ← mini-gateway (Stage 9)
                        │  (限流/熔断/认证) │
                        └────────┬────────┘
                                 │ RPC
              ┌──────────────────┼──────────────────┐
              │                  │                   │
     ┌────────▼──────┐  ┌───────▼───────┐  ┌───────▼───────┐
     │ User Service  │  │ Product Svc   │  │ Order Service │
     │ (注册/登录)    │  │ (商品CRUD)     │  │ (下单/状态机)  │
     └────────┬──────┘  └───────┬───────┘  └───────┬───────┘
              │                 │                    │
     ┌────────▼──────┐  ┌──────▼────────┐          │
     │    MySQL      │  │ MySQL + Redis │          │
     └───────────────┘  └───────────────┘          │
                                            ┌──────▼────────┐
                                            │   Kafka       │
                                            │ (异步解耦)     │
                                            └──────┬────────┘
                                     ┌─────────────┼─────────────┐
                                     │                           │
                            ┌────────▼──────┐          ┌────────▼──────┐
                            │Inventory Svc  │          │ Payment Svc   │
                            │(库存扣减)      │          │(模拟支付)      │
                            └────────┬──────┘          └───────────────┘
                                     │
                            ┌────────▼──────┐
                            │ MySQL + Redis │
                            └───────────────┘

横切关注点（所有服务共享）：
├── Service Registry (etcd)          ← Stage 8
├── RPC Framework (mini-rpc)         ← Stage 7
├── Observability (Prometheus/Jaeger) ← Stage 10
└── Config Center (etcd)             ← Stage 8
```

---

## 3. 服务划分

### 3.1 User Service（用户服务）

```go
// 复用 Stage 3 的用户服务
type UserService interface {
    Register(ctx context.Context, req *RegisterReq) (*User, error)
    Login(ctx context.Context, req *LoginReq) (*TokenPair, error)
    GetUser(ctx context.Context, userID int64) (*User, error)
}
```

### 3.2 Product Service（商品服务）

```go
type Product struct {
    ID          int64   `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       int64   `json:"price"`       // 分为单位，避免浮点精度
    Stock       int64   `json:"stock"`       // 冗余字段，实际库存在 Inventory
    CategoryID  int64   `json:"category_id"`
    Status      int     `json:"status"`      // 0:下架 1:上架
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ProductService interface {
    CreateProduct(ctx context.Context, req *CreateProductReq) (*Product, error)
    GetProduct(ctx context.Context, id int64) (*Product, error)
    ListProducts(ctx context.Context, req *ListReq) ([]*Product, int64, error)
    UpdateProduct(ctx context.Context, req *UpdateProductReq) error
}

// 缓存策略：
// 1. 商品详情 → Redis Cache (TTL 10min + 随机偏移防雪崩)
// 2. 热门商品列表 → Redis Sorted Set
// 3. Cache-Aside 模式：读缓存 → miss → 读 DB → 写缓存
```

### 3.3 Order Service（订单服务）

```go
type Order struct {
    ID         string    `json:"id"`          // 分布式 ID（雪花算法）
    UserID     int64     `json:"user_id"`
    Items      []OrderItem `json:"items"`
    TotalPrice int64     `json:"total_price"` // 分
    Status     OrderStatus `json:"status"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type OrderItem struct {
    ProductID int64 `json:"product_id"`
    Quantity  int   `json:"quantity"`
    Price     int64 `json:"price"`
}

type OrderStatus int

const (
    OrderPending   OrderStatus = 0  // 待支付
    OrderPaid      OrderStatus = 1  // 已支付
    OrderShipped   OrderStatus = 2  // 已发货
    OrderCompleted OrderStatus = 3  // 已完成
    OrderCancelled OrderStatus = 4  // 已取消
)

type OrderService interface {
    CreateOrder(ctx context.Context, req *CreateOrderReq) (*Order, error)
    GetOrder(ctx context.Context, orderID string) (*Order, error)
    PayOrder(ctx context.Context, orderID string) error
    CancelOrder(ctx context.Context, orderID string) error
    ListUserOrders(ctx context.Context, userID int64, page, size int) ([]*Order, int64, error)
}
```

### 3.4 Inventory Service（库存服务）

```go
type Inventory struct {
    ProductID  int64 `json:"product_id"`
    Stock      int64 `json:"stock"`       // 总库存
    Locked     int64 `json:"locked"`      // 锁定库存（已下单未支付）
    Available  int64 `json:"available"`   // 可用库存 = stock - locked
}

type InventoryService interface {
    // 预扣库存（下单时调用）
    Deduct(ctx context.Context, productID int64, quantity int) error

    // 确认扣减（支付成功后调用）
    Confirm(ctx context.Context, productID int64, quantity int) error

    // 回滚库存（支付超时/取消时调用）
    Rollback(ctx context.Context, productID int64, quantity int) error

    // 查询库存
    GetStock(ctx context.Context, productID int64) (*Inventory, error)
}

// 防超卖方案：
// 1. Redis 预扣（DECRBY + Lua 原子操作）
// 2. MySQL 乐观锁（UPDATE ... WHERE stock >= quantity）
// 3. 库存预热到 Redis，异步同步到 MySQL
```

### 3.5 Payment Service（支付服务）

```go
type Payment struct {
    ID        string        `json:"id"`
    OrderID   string        `json:"order_id"`
    Amount    int64         `json:"amount"`
    Status    PaymentStatus `json:"status"`
    Channel   string        `json:"channel"`    // mock / alipay / wechat
    CreatedAt time.Time
}

type PaymentService interface {
    // 创建支付单
    CreatePayment(ctx context.Context, orderID string, amount int64) (*Payment, error)

    // 模拟支付回调（实际场景由支付网关回调）
    HandleCallback(ctx context.Context, paymentID string, success bool) error
}
```

---

## 4. 核心架构问题解决方案

### 4.1 下单流程（分布式事务 — Saga 模式）

```
下单请求
    │
    ▼
Order Service: 创建订单（状态=Pending）
    │
    ▼ (RPC)
Inventory Service: 预扣库存（Deduct）
    │ 失败 → 回滚：取消订单
    ▼ 成功
Order Service: 更新订单状态 → 发送 Kafka 消息
    │
    ▼ (Kafka)
Payment Service: 创建支付单 → 等待用户支付
    │
    ├── 支付成功 → Kafka → Inventory: Confirm + Order: 更新已支付
    │
    └── 支付超时(30min) → Kafka → Inventory: Rollback + Order: 取消
```

```go
// saga.go — Saga 编排器
type SagaStep struct {
    Name     string
    Execute  func(ctx context.Context) error  // 正向操作
    Rollback func(ctx context.Context) error  // 补偿操作
}

type Saga struct {
    steps     []SagaStep
    completed []int  // 已完成的步骤索引
}

func (s *Saga) AddStep(step SagaStep) {
    s.steps = append(s.steps, step)
}

// 按序执行，失败则反向补偿
func (s *Saga) Execute(ctx context.Context) error {
    for i, step := range s.steps {
        if err := step.Execute(ctx); err != nil {
            // 反向补偿已完成的步骤
            for j := i - 1; j >= 0; j-- {
                if rollbackErr := s.steps[j].Rollback(ctx); rollbackErr != nil {
                    // 补偿失败，记录日志，人工介入
                    log.Error("saga rollback failed", "step", s.steps[j].Name, "err", rollbackErr)
                }
            }
            return fmt.Errorf("saga step '%s' failed: %w", step.Name, err)
        }
        s.completed = append(s.completed, i)
    }
    return nil
}
```

### 4.2 库存防超卖（Redis + Lua）

```go
// inventory_redis.go

// Lua 脚本：原子性检查并扣减库存
const deductStockScript = `
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil then
    return -1  -- key 不存在
end
local quantity = tonumber(ARGV[1])
if stock < quantity then
    return 0   -- 库存不足
end
redis.call('DECRBY', KEYS[1], quantity)
return 1       -- 扣减成功
`

func (r *InventoryRedis) Deduct(ctx context.Context, productID int64, quantity int) error {
    key := fmt.Sprintf("inventory:stock:%d", productID)
    result, err := r.client.Eval(ctx, deductStockScript, []string{key}, quantity).Int()
    if err != nil {
        return err
    }
    switch result {
    case 1:
        return nil
    case 0:
        return ErrInsufficientStock
    case -1:
        return ErrStockNotPreloaded
    }
    return ErrUnknown
}
```

### 4.3 秒杀场景

```
正常流量 ──► API Gateway (限流 1000 QPS)
                │
                ▼
         秒杀接口 ──► Redis 预扣库存 (Lua 原子操作)
                          │
                    ┌─────┼─────┐
                    │           │
                 库存不足     扣减成功
                 快速返回      │
                              ▼
                         Kafka 异步
                              │
                              ▼
                     Order Service (创建订单)
                              │
                              ▼
                     Payment (等待支付, 30min 超时)
```

**关键策略：**

```go
// seckill.go
type SeckillService struct {
    gateway       *RateLimiter        // 限流
    inventory     *InventoryRedis     // Redis 预扣
    orderProducer *kafka.Producer     // 异步下单
    bloomFilter   *BloomFilter        // 用户去重（一人一单）
}

func (s *SeckillService) Seckill(ctx context.Context, userID, productID int64) error {
    // 1. 布隆过滤器：一人一单检查
    if s.bloomFilter.Contains(seckillKey(userID, productID)) {
        return ErrAlreadyPurchased
    }

    // 2. Redis 原子扣库存
    if err := s.inventory.Deduct(ctx, productID, 1); err != nil {
        return err // 库存不足，快速返回
    }

    // 3. 标记用户已购买
    s.bloomFilter.Add(seckillKey(userID, productID))

    // 4. 异步发送到 Kafka（削峰）
    msg := &OrderMessage{UserID: userID, ProductID: productID, Quantity: 1}
    return s.orderProducer.Send(ctx, "seckill-orders", msg)
}
```

### 4.4 缓存一致性（延迟双删）

```go
// cache_consistency.go

// 更新商品信息时保证缓存一致性
func (s *ProductService) UpdateProduct(ctx context.Context, req *UpdateProductReq) error {
    // 1. 删除缓存
    s.cache.Delete(ctx, productCacheKey(req.ID))

    // 2. 更新数据库
    if err := s.repo.Update(ctx, req); err != nil {
        return err
    }

    // 3. 延迟双删（防止并发读写导致缓存脏数据）
    go func() {
        time.Sleep(500 * time.Millisecond)
        s.cache.Delete(context.Background(), productCacheKey(req.ID))
    }()

    return nil
}
```

### 4.5 幂等设计（防重复提交）

```go
// idempotent.go

// 基于 Redis 的幂等 Token 方案
// 1. 客户端先请求获取一个幂等 Token
// 2. 提交订单时携带 Token
// 3. 服务端用 SETNX 检查 Token 是否已使用

type IdempotentChecker struct {
    rdb *redis.Client
    ttl time.Duration
}

func (ic *IdempotentChecker) GenerateToken(ctx context.Context) (string, error) {
    token := uuid.New().String()
    key := "idempotent:" + token
    ic.rdb.Set(ctx, key, "1", ic.ttl)
    return token, nil
}

func (ic *IdempotentChecker) Check(ctx context.Context, token string) (bool, error) {
    key := "idempotent:" + token
    // SETNX 的逆操作：删除成功说明 Token 有效且未被使用
    result, err := ic.rdb.Del(ctx, key).Result()
    return result == 1, err
}
```

---

## 5. 目录结构

```
projects/stage11-mini-ecommerce/
├── api-gateway/                   # 复用 Stage 9
│   └── ...
├── services/
│   ├── user/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── product/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── cache/           # 商品缓存策略
│   │   │   └── model/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── order/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   │   ├── order_service.go
│   │   │   │   ├── saga.go      # Saga 编排器
│   │   │   │   └── seckill.go   # 秒杀逻辑
│   │   │   ├── repository/
│   │   │   ├── consumer/        # Kafka 消费者
│   │   │   └── model/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── inventory/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── cache/           # Redis 库存预扣
│   │   │   └── model/
│   │   ├── Dockerfile
│   │   └── go.mod
│   └── payment/
│       ├── cmd/main.go
│       ├── internal/
│       ├── Dockerfile
│       └── go.mod
├── pkg/
│   ├── rpc/                      # 复用 Stage 7 mini-rpc
│   ├── idgen/                    # 分布式 ID 生成器
│   ├── idempotent/               # 幂等工具
│   └── cache/                    # 缓存工具（Cache-Aside / 延迟双删）
├── configs/
│   ├── gateway.yaml
│   ├── prometheus.yml
│   └── otel-collector.yml
├── deployments/
│   ├── docker-compose.yml        # 一键启动所有服务
│   └── docker-compose.infra.yml  # 基础设施（MySQL/Redis/Kafka/etcd）
├── scripts/
│   ├── init-db.sql               # 数据库初始化
│   ├── seed-data.sh              # 种子数据
│   └── seckill-bench.js          # k6 秒杀压测脚本
├── test/
│   ├── e2e/                      # 端到端测试
│   └── chaos/                    # 混沌测试
├── Makefile
└── README.md
```

---

## 6. 实现里程碑

### M1: 基础服务搭建（第1-2周）
- [ ] 项目骨架 + 服务注册发现
- [ ] User Service（复用 Stage 3）
- [ ] Product Service（CRUD + Redis 缓存）
- [ ] Docker Compose 基础设施

### M2: 订单 + 库存（第3-4周）
- [ ] Order Service（创建 + 状态机）
- [ ] Inventory Service（MySQL + Redis 预扣）
- [ ] Saga 分布式事务（下单 → 扣库存）
- [ ] 分布式 ID 生成器

### M3: 支付 + 异步消息（第5周）
- [ ] Payment Service（模拟支付）
- [ ] Kafka 异步消息（下单 → 支付 → 库存确认）
- [ ] 支付超时自动取消（延迟队列）
- [ ] 幂等防重

### M4: 秒杀场景（第6周）
- [ ] Redis 库存预热
- [ ] Lua 原子扣减
- [ ] 布隆过滤器一人一单
- [ ] Kafka 削峰异步下单
- [ ] 限流策略配置

### M5: 可观测性 + 压测（第7周）
- [ ] 全服务集成 Prometheus + Jaeger
- [ ] Grafana 全链路 Dashboard
- [ ] k6 秒杀压测脚本
- [ ] 压测报告

### M6: 混沌测试 + 完善（第8周）
- [ ] 模拟服务宕机 → 验证降级
- [ ] 模拟 Redis 故障 → 验证降级到 DB
- [ ] 模拟 Kafka 积压 → 验证反压
- [ ] 端到端测试完善
- [ ] 文档输出

---

## 7. 毕业标准

- [ ] 一键 `docker-compose up` 启动整个电商平台
- [ ] 下单 → 扣库存 → 支付的完整流程正确
- [ ] 秒杀场景下库存不超卖
- [ ] 支付超时自动回滚库存
- [ ] 单服务宕机不影响其他服务
- [ ] 全链路追踪可在 Jaeger 中可视化
- [ ] 压测报告：秒杀 QPS / P99 延迟 / 成功率
- [ ] 能画出完整的数据流 + 控制流 + 故障路径
