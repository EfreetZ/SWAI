# Stage 0 — 工程基础: engineering-template

> 预计周期：2 周 | 语言：Golang | 难度：⭐
>
> 目标：建立专业工程习惯，而非"能跑就行"

---

## 1. 项目目标

构建一个**可直接复用的企业级 Golang 项目脚手架**，所有后续 Stage 的项目都基于此模板启动：

- **标准项目结构** — 遵循 Go 社区最佳实践
- **构建自动化** — Makefile 一键构建 / 测试 / lint
- **容器化** — Dockerfile 多阶段构建 + docker-compose 编排
- **CI/CD** — GitHub Actions 自动化流水线
- **测试规范** — 单元测试 + 基准测试 + 竞争检测
- **日志规范** — 结构化日志（slog）

---

## 2. 标准项目结构

```
project-name/
├── cmd/
│   └── server/
│       └── main.go              # 入口文件，仅负责组装和启动
├── internal/                     # 私有代码（不可被外部引用）
│   ├── config/
│   │   └── config.go            # 配置加载（yaml + 环境变量）
│   ├── handler/                  # HTTP/RPC Handler
│   ├── service/                  # 业务逻辑层
│   ├── repository/               # 数据访问层
│   ├── model/                    # 数据模型定义
│   ├── middleware/                # 中间件
│   └── server/
│       └── server.go            # 服务启动与关闭
├── pkg/                          # 可被外部引用的公共包
│   └── utils/
├── api/                          # API 定义（proto / swagger）
├── configs/
│   └── config.yaml              # 默认配置文件
├── migrations/                   # 数据库迁移文件
├── scripts/                      # 工具脚本
├── test/                         # 集成测试
├── .github/
│   └── workflows/
│       └── ci.yml               # GitHub Actions CI
├── Dockerfile                    # 多阶段构建
├── docker-compose.yml            # 本地开发编排
├── .golangci.yml                 # golangci-lint 配置
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 关键原则

- `cmd/` — 每个可执行文件一个子目录，`main.go` 只做组装（DI），不写业务逻辑
- `internal/` — Go 编译器强制私有，外部包无法 import
- `pkg/` — 真正可复用的公共库才放这里，不要滥用
- 分层：`handler → service → repository`，依赖方向单向向下

---

## 3. Makefile

```makefile
.PHONY: build run test lint fmt vet race bench clean docker-build docker-run

APP_NAME := server
BUILD_DIR := bin

# 构建
build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

# 测试
test:
	go test ./... -v -count=1

race:
	go test -race ./... -v

bench:
	go test -bench=. -benchmem ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码质量
lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# Docker
docker-build:
	docker build -t $(APP_NAME):latest .

docker-run:
	docker-compose up -d

docker-down:
	docker-compose down

# 清理
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
```

---

## 4. Dockerfile（多阶段构建）

```dockerfile
# === 构建阶段 ===
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# === 运行阶段 ===
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/server .
COPY configs/ ./configs/

EXPOSE 8080
ENTRYPOINT ["./server"]
```

**为什么用多阶段构建：**
- 构建镜像包含完整 Go 工具链（~1GB）
- 运行镜像只包含二进制文件（~10-20MB）
- 减少攻击面，生产更安全

---

## 5. docker-compose

```yaml
version: "3.8"

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=root
      - MYSQL_DATABASE=app
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 10

volumes:
  mysql_data:
```

---

## 6. GitHub Actions CI

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: golangci/golangci-lint-action@v4

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out
```

---

## 7. 配置管理

```go
// internal/config/config.go

type Config struct {
    App   AppConfig   `yaml:"app"`
    DB    DBConfig    `yaml:"db"`
    Redis RedisConfig `yaml:"redis"`
    Log   LogConfig   `yaml:"log"`
}

type AppConfig struct {
    Name string `yaml:"name" env:"APP_NAME"`
    Port int    `yaml:"port" env:"APP_PORT"`
    Env  string `yaml:"env"  env:"APP_ENV"`
}

type DBConfig struct {
    Host     string `yaml:"host"     env:"DB_HOST"`
    Port     int    `yaml:"port"     env:"DB_PORT"`
    User     string `yaml:"user"     env:"DB_USER"`
    Password string `yaml:"password" env:"DB_PASSWORD"`
    Database string `yaml:"database" env:"DB_NAME"`
}

type RedisConfig struct {
    Host string `yaml:"host" env:"REDIS_HOST"`
    Port int    `yaml:"port" env:"REDIS_PORT"`
}

type LogConfig struct {
    Level  string `yaml:"level"  env:"LOG_LEVEL"`
    Format string `yaml:"format" env:"LOG_FORMAT"` // json / text
}

// 加载顺序：yaml 文件 → 环境变量覆盖
func Load(path string) (*Config, error) {
    cfg := &Config{}
    // 1. 读取 yaml
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    // 2. 环境变量覆盖（使用 env tag 反射读取）
    overrideFromEnv(cfg)
    return cfg, nil
}
```

---

## 8. 结构化日志

```go
// internal/logger/logger.go
// 使用 Go 1.21+ 标准库 slog

func New(level, format string) *slog.Logger {
    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: parseLevel(level)}

    switch format {
    case "json":
        handler = slog.NewJSONHandler(os.Stdout, opts)
    default:
        handler = slog.NewTextHandler(os.Stdout, opts)
    }

    logger := slog.New(handler)
    slog.SetDefault(logger) // 设为全局默认
    return logger
}

// 日志规范：
// 1. 使用结构化字段，禁止 fmt.Sprintf 拼接
// 2. 必须携带 request_id（方便链路追踪）
// 3. 错误日志必须携带 error 字段
// 4. 禁止打印密码、token 等敏感信息

// 正确示例
slog.Info("user created", "user_id", 42, "username", "alice")
slog.Error("db query failed", "error", err, "query", "SELECT ...")

// 错误示例 ❌
log.Printf("user %d created", userID)        // 非结构化
slog.Info("password=" + password)              // 敏感信息
```

---

## 9. Git 规范

### 分支策略

```
main ──────────────────────────────── 生产分支
  │
  ├── develop ─────────────────────── 开发分支
  │     │
  │     ├── feature/add-user-api ──── 功能分支
  │     ├── feature/add-rbac ──────── 功能分支
  │     │
  │     └── hotfix/fix-login-bug ──── 紧急修复
  │
  └── release/v1.0.0 ─────────────── 发布分支
```

### Commit 规范（Conventional Commits）

```
<type>(<scope>): <description>

type:
  feat     — 新功能
  fix      — 修复 Bug
  docs     — 文档变更
  style    — 代码格式（不影响逻辑）
  refactor — 重构
  perf     — 性能优化
  test     — 测试
  chore    — 构建 / 工具变更

示例:
  feat(user): add login API
  fix(auth): fix JWT token expiration check
  docs: update README with quick start guide
  test(cache): add LRU cache benchmark
```

---

## 10. golangci-lint 配置

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck       # 检查未处理的 error
    - govet          # go vet 检查
    - staticcheck    # 静态分析
    - unused         # 未使用的代码
    - gosimple       # 简化建议
    - ineffassign    # 无效赋值
    - misspell       # 拼写检查
    - gofmt          # 格式检查
    - goimports      # import 排序

linters-settings:
  errcheck:
    check-type-assertions: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

---

## 11. 实现里程碑

### M1: 项目结构 + Makefile（第1-2天）
- [ ] 创建标准目录结构
- [ ] 编写 Makefile（build / test / lint / fmt）
- [ ] 编写 .golangci.yml
- [ ] 编写 .gitignore

### M2: 配置 + 日志（第3-4天）
- [ ] 配置加载（yaml + 环境变量）
- [ ] 结构化日志（slog）
- [ ] 健康检查端点（GET /health）

### M3: Docker + CI（第5-7天）
- [ ] Dockerfile 多阶段构建
- [ ] docker-compose（App + MySQL + Redis）
- [ ] GitHub Actions CI（lint + test + race）
- [ ] README 编写

### M4: Git 规范（第8-10天）
- [ ] .gitignore 完善
- [ ] Commit 规范文档
- [ ] 分支策略文档
- [ ] 创建第一个 PR 实践完整流程

---

## 12. 毕业标准

- [ ] `make build` 一键构建二进制
- [ ] `make test` 所有测试通过
- [ ] `make lint` 无告警
- [ ] `docker-compose up` 一键启动全部服务
- [ ] GitHub Actions CI 绿色通过
- [ ] 项目 README 包含：简介、快速开始、目录结构说明
- [ ] 能解释为什么用 `internal/` 而不是所有代码放 `pkg/`
