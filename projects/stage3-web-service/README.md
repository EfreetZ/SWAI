# Stage 3 — Web 系统演进：user-service

> 对应设计文档：`docs/design/stage3-web-service.md`

## 已实现能力

- 用户注册/登录/刷新 Token
- JWT 认证中间件
- RBAC 权限校验中间件
- 用户接口与角色接口（统一 `{code,data,message}` 响应）
- RequestID / Logger / Recovery / CORS 中间件
- 优雅关闭
- 运行时监控与阈值告警模块
- 单元测试、基准测试、性能/压力/混沌测试

## 快速开始

```bash
make test
make race
make bench
make run
```

服务启动后：`http://localhost:8081/health`

## 关键 API

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/users`
- `GET|PUT|DELETE /api/v1/users/{id}`
- `GET /api/v1/roles`

## Docker

```bash
make docker-build
make docker-run
```

## 文档

- 排障手册：`TROUBLESHOOTING.md`
