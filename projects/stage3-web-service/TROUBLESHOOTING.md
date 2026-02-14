# Stage 3 排障手册

## 1. 登录后访问用户接口返回 401

### 定位
1. 检查 `Authorization` 请求头是否为 `Bearer <access_token>`。
2. 确认 token 未过期，并且类型为 `access`。
3. 运行 `go test ./internal/middleware -v` 验证鉴权链路。

### 处理
- 刷新 token 后重试。
- 校验服务端 `JWT_SECRET` 与签发端一致。

## 2. 返回 403（权限不足）

### 定位
1. 检查当前用户角色（admin/editor/viewer）。
2. 检查 RBAC 权限映射是否包含目标 `resource/action`。

### 处理
- 调整角色权限配置。
- 对只允许本人访问的接口，确认 user_id 是否匹配。

## 3. 服务退出时请求丢失

### 定位
1. 检查是否通过 SIGTERM/SIGINT 触发优雅关闭。
2. 检查 `WaitForShutdown` 超时是否过短。

### 处理
- 适当增加优雅关闭超时。
- 在关闭前停止接收新流量。

## 4. 慢请求排查

### 定位
1. 对关键接口加请求耗时日志。
2. 分析 goroutine 和 heap 指标变化。
3. 运行 `go test ./test/performance -v` 对比优化前后吞吐。

### 处理
- 优化热点路径。
- 增加超时控制，避免下游阻塞放大。
