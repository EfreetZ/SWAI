# Stage 3 — Web 系统演进: user-service

> 预计周期：3 周 | 语言：Golang | 难度：⭐⭐⭐
>
> 体验从单体到微服务的架构升级路径

---

## 1. 项目目标

构建一个生产级用户服务系统，掌握：

- **Web 开发全流程：** 路由 → 中间件 → 业务逻辑 → 数据库 → 响应
- **认证鉴权：** JWT Token 签发与验证
- **权限控制：** RBAC 模型实现
- **工程规范：** 统一响应格式、参数校验、错误处理、优雅关闭
- **架构演进感知：** 理解单体为什么要拆分、如何拆分

---

## 2. 架构升级路径

```
V1: 单体应用
    所有逻辑在一个进程内
    │
V2: 模块化
    按 domain 拆分包（user / auth / permission）
    │
V3: 服务拆分准备
    数据库访问抽象为 Repository 接口
    业务逻辑抽象为 Service 接口
    → 为后续拆分为独立微服务做准备
```

---

## 3. 整体架构

```
┌──────────────────────────────────────────────┐
│                   Client                      │
└──────────────────┬───────────────────────────┘
                   │ HTTP
┌──────────────────▼───────────────────────────┐
│              Middleware Chain                  │
│  RequestID → Logger → Recovery → CORS        │
│  → RateLimit → Auth(JWT) → RBAC             │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│                 Router                        │
│  /api/v1/auth/register    → AuthHandler      │
│  /api/v1/auth/login       → AuthHandler      │
│  /api/v1/users/:id        → UserHandler      │
│  /api/v1/roles            → RoleHandler      │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│              Service Layer                    │
│  UserService / AuthService / RoleService     │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│            Repository Layer                   │
│  UserRepo / RoleRepo / PermissionRepo        │
└──────────────────┬───────────────────────────┘
                   │
           ┌───────┼───────┐
           │               │
    ┌──────▼──────┐ ┌──────▼──────┐
    │    MySQL    │ │    Redis    │
    │  (持久化)    │ │  (Session)  │
    └─────────────┘ └─────────────┘
```

---

## 4. 核心模块设计

### 4.1 统一响应格式

```go
// response.go
type Response struct {
    Code    int         `json:"code"`
    Data    interface{} `json:"data"`
    Message string      `json:"message"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, &Response{
        Code:    0,
        Data:    data,
        Message: "success",
    })
}

// 错误响应
func Error(c *gin.Context, httpCode int, bizCode int, msg string) {
    c.JSON(httpCode, &Response{
        Code:    bizCode,
        Data:    nil,
        Message: msg,
    })
}

// 业务错误码定义
const (
    ErrCodeSuccess       = 0
    ErrCodeBadRequest    = 10001
    ErrCodeUnauthorized  = 10002
    ErrCodeForbidden     = 10003
    ErrCodeNotFound      = 10004
    ErrCodeInternal      = 10005
    ErrCodeUserExists    = 20001
    ErrCodeWrongPassword = 20002
    ErrCodeTokenExpired  = 20003
)
```

### 4.2 JWT 认证

```go
// jwt.go
type Claims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

type JWTManager struct {
    secretKey     []byte
    accessExpiry  time.Duration
    refreshExpiry time.Duration
}

// 签发 Access Token + Refresh Token
func (m *JWTManager) GenerateTokenPair(userID int64, username, role string) (accessToken, refreshToken string, err error)

// 验证并解析 Token
func (m *JWTManager) ParseToken(tokenStr string) (*Claims, error)

// 刷新 Token（用 Refresh Token 换新的 Access Token）
func (m *JWTManager) RefreshToken(refreshToken string) (newAccessToken string, err error)
```

### 4.3 RBAC 权限模型

```go
// rbac.go

// 角色
type Role struct {
    ID          int64        `json:"id"`
    Name        string       `json:"name"`         // admin / editor / viewer
    Permissions []Permission `json:"permissions"`
}

// 权限
type Permission struct {
    ID       int64  `json:"id"`
    Resource string `json:"resource"`   // "user" / "order" / "product"
    Action   string `json:"action"`     // "create" / "read" / "update" / "delete"
}

// 用户-角色关系
type UserRole struct {
    UserID int64
    RoleID int64
}

// RBAC 检查器
type RBACChecker struct {
    roleRepo RoleRepository
}

// 检查用户是否有权限执行某个操作
func (r *RBACChecker) HasPermission(userID int64, resource, action string) (bool, error)
```

### 4.4 中间件

```go
// middleware.go

// 请求 ID（链路追踪基础）
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// 结构化日志
func LoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        logger.Info("request",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Writer.Status(),
            "duration_ms", time.Since(start).Milliseconds(),
            "request_id", c.GetString("request_id"),
            "client_ip", c.ClientIP(),
        )
    }
}

// Panic Recovery
func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("panic recovered",
                    "error", err,
                    "stack", string(debug.Stack()),
                    "request_id", c.GetString("request_id"),
                )
                Error(c, http.StatusInternalServerError, ErrCodeInternal, "internal server error")
                c.Abort()
            }
        }()
        c.Next()
    }
}

// JWT 认证中间件
func AuthMiddleware(jwtMgr *JWTManager) gin.HandlerFunc
// RBAC 权限中间件
func RBACMiddleware(checker *RBACChecker, resource, action string) gin.HandlerFunc
```

### 4.5 优雅关闭

```go
// graceful.go
func GracefulShutdown(server *http.Server, timeout time.Duration) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    // 1. 停止接收新请求
    if err := server.Shutdown(ctx); err != nil {
        log.Error("server forced to shutdown", "error", err)
    }

    // 2. 关闭数据库连接
    // 3. 关闭 Redis 连接
    // 4. 刷新日志缓冲

    log.Info("server exited")
}
```

---

## 5. API 设计

### 认证接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 注册 | 无 |
| POST | `/api/v1/auth/login` | 登录 | 无 |
| POST | `/api/v1/auth/refresh` | 刷新 Token | Refresh Token |
| POST | `/api/v1/auth/logout` | 登出 | Access Token |

### 用户接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/users` | 用户列表 | admin |
| GET | `/api/v1/users/:id` | 用户详情 | admin / 本人 |
| PUT | `/api/v1/users/:id` | 更新用户 | admin / 本人 |
| DELETE | `/api/v1/users/:id` | 删除用户 | admin |

### 角色接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/api/v1/roles` | 角色列表 | admin |
| POST | `/api/v1/roles` | 创建角色 | admin |
| PUT | `/api/v1/roles/:id` | 更新角色 | admin |
| POST | `/api/v1/users/:id/roles` | 分配角色 | admin |

---

## 6. 目录结构

```
projects/stage3-web-service/
├── cmd/
│   └── server/
│       └── main.go                # 入口
├── internal/
│   ├── config/
│   │   └── config.go              # 配置加载（yaml / env）
│   ├── handler/
│   │   ├── auth.go                # 认证 Handler
│   │   ├── user.go                # 用户 Handler
│   │   ├── role.go                # 角色 Handler
│   │   └── response.go            # 统一响应
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   ├── auth.go                # JWT 认证
│   │   └── rbac.go                # RBAC 权限
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   └── role_service.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── role_repo.go
│   │   └── mysql.go               # MySQL 连接管理
│   ├── model/
│   │   ├── user.go                # User 实体
│   │   ├── role.go                # Role + Permission 实体
│   │   └── errors.go              # 业务错误定义
│   ├── pkg/
│   │   ├── jwt.go                 # JWT 工具
│   │   ├── password.go            # bcrypt 密码哈希
│   │   └── validator.go           # 参数校验
│   └── server/
│       ├── router.go              # 路由注册
│       ├── server.go              # HTTP Server
│       └── graceful.go            # 优雅关闭
├── migrations/
│   └── 001_init.sql               # 数据库建表 SQL
├── configs/
│   └── config.yaml                # 配置文件
├── docker-compose.yml             # App + MySQL + Redis
├── Dockerfile
├── test/
│   ├── auth_test.go
│   ├── user_test.go
│   └── rbac_test.go
├── go.mod
├── Makefile
└── README.md
```

---

## 7. 实现里程碑

### M1: 基础框架 + 用户 CRUD（第1周）
- [ ] 项目结构搭建
- [ ] 配置加载（yaml + 环境变量）
- [ ] MySQL 连接 + 建表
- [ ] 用户注册 / 登录 / CRUD
- [ ] 统一响应格式
- [ ] 参数校验
- [ ] Docker 部署（App + MySQL）

### M2: JWT + RBAC + 中间件（第2周）
- [ ] JWT 签发 / 验证 / 刷新
- [ ] RBAC 角色权限模型
- [ ] 请求日志中间件
- [ ] Panic Recovery 中间件
- [ ] CORS 中间件
- [ ] RequestID 中间件

### M3: 完善 + 故障注入（第3周）
- [ ] 优雅关闭
- [ ] Swagger 文档
- [ ] 制造慢请求 → pprof 分析
- [ ] 制造 DB 连接耗尽 → 连接池调优
- [ ] 制造 Goroutine 泄漏 → 排查修复
- [ ] 完整集成测试

---

## 8. 毕业标准

- [ ] 完整的用户注册登录流程
- [ ] JWT Token 过期后能通过 Refresh Token 续期
- [ ] RBAC 能正确拦截无权限请求
- [ ] 优雅关闭不丢失正在处理的请求
- [ ] Docker 一键启动全部服务
- [ ] 能画出请求从进入到响应的完整链路
