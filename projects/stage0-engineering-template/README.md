# Stage 0 — 工程基础脚手架

> SWAI 项目 Stage 0 实现：企业级 Golang 项目模板

## 项目简介

本项目是一个**可直接复用的 Go 项目脚手架**，所有后续 Stage 的项目都基于此模板启动。

核心能力：
- 标准项目结构（`cmd/` + `internal/` + `pkg/`）
- 配置管理（YAML + 环境变量覆盖）
- 结构化日志（Go 标准库 `slog`）
- HTTP 中间件链（RequestID → Logging → Recovery）
- 优雅关闭（Graceful Shutdown）
- Docker 多阶段构建
- GitHub Actions CI

## 快速开始

### 本地运行

```bash
# 构建并运行
make run

# 测试健康检查
curl http://localhost:8080/health
```

### Docker 运行

```bash
# 构建镜像
make docker-build

# 启动全部服务（App + MySQL + Redis）
make docker-run

# 停止
make docker-down
```

## 测试命令

```bash
# 单元测试
make test

# 竞争检测（并发安全验证）
make race

# 基准测试
make bench

# 覆盖率报告
make cover
```

## 目录结构

```
stage0-engineering-template/
├── cmd/server/main.go           # 入口：仅组装依赖，不写业务逻辑
├── internal/                    # 私有代码（Go 编译器强制不可外部引用）
│   ├── config/                  # 配置加载（YAML + 环境变量）
│   ├── handler/                 # HTTP 请求处理器
│   ├── logger/                  # 结构化日志
│   ├── middleware/              # 中间件（RequestID / Logging / Recovery）
│   └── server/                  # HTTP Server + 优雅关闭
├── pkg/response/                # 标准 JSON 响应格式（可被外部引用）
├── configs/config.yaml          # 默认配置
├── test/                        # 集成测试
├── Dockerfile                   # 多阶段构建
├── docker-compose.yml           # 本地开发编排
├── Makefile                     # 构建自动化
└── .golangci.yml                # Lint 配置
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查（Docker/LB 探活） |

**响应格式：**

```json
{
  "code": 0,
  "data": { "status": "ok", "version": "0.1.0" },
  "message": "success"
}
```

## 工程规范

### 分层架构

```
Handler → Service → Repository
   ↑                    ↓
   └── 依赖方向单向向下 ──┘
```

### 日志规范

```go
// ✅ 正确：结构化字段
slog.Info("user created", "user_id", 42, "username", "alice")
slog.Error("db query failed", "error", err, "request_id", reqID)

// ❌ 错误：字符串拼接、敏感信息
log.Printf("user %d created", userID)
slog.Info("password=" + password)
```

### Commit 规范

```
feat(config): add env override support
fix(middleware): fix request_id missing in panic recovery
test(handler): add health check benchmark
```

## 排障指南

见 [TROUBLESHOOTING.md](./TROUBLESHOOTING.md)
