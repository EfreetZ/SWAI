# Stage 0 排障手册

## 常见问题

### 1. `make build` 编译失败

**现象：** `cannot find module providing package xxx`

**排查：**
```bash
# 检查 go.mod 是否完整
cat go.mod

# 重新下载依赖
go mod tidy

# 验证依赖
go mod verify
```

**根因：** 新增了 import 但未运行 `go mod tidy`

---

### 2. `make test` 失败

**现象：** 测试报错或 panic

**排查步骤：**
```bash
# 运行单个包的测试，查看详细输出
go test -v ./internal/config/

# 运行单个测试函数
go test -v -run TestLoad ./internal/config/

# 查看竞争检测
go test -race ./...
```

---

### 3. Docker 构建失败

**现象：** `docker build` 报错

**常见原因：**

| 错误 | 原因 | 解决 |
|------|------|------|
| `go mod download` 超时 | 网络问题 / GOPROXY | 设置 `GOPROXY=https://goproxy.cn,direct` |
| `COPY go.sum` 失败 | 缺少 go.sum | 本地运行 `go mod tidy` |
| 二进制无法运行 | CGO 依赖 | 确保 `CGO_ENABLED=0` |

**修复 GOPROXY：**
```dockerfile
# 在 Dockerfile 的 builder 阶段添加
ENV GOPROXY=https://goproxy.cn,direct
```

---

### 4. `docker-compose up` 服务启动顺序问题

**现象：** App 容器启动后立即退出，日志显示连接 MySQL/Redis 失败

**排查：**
```bash
# 查看容器日志
docker-compose logs app
docker-compose logs mysql

# 检查 healthcheck 状态
docker-compose ps
```

**根因：** MySQL 还未 ready，App 就开始连接

**解决：** docker-compose.yml 中 `depends_on` + `condition: service_healthy` 已处理。如果仍有问题，在应用代码中加入**重连机制**。

---

### 5. 端口被占用

**现象：** `bind: address already in use`

**排查：**
```bash
# 查看占用端口的进程
lsof -i :8080
# 或
ss -tlnp | grep 8080

# 杀掉进程
kill -9 <PID>
```

---

### 6. 配置加载失败

**现象：** `failed to load config: 读取配置文件失败`

**排查：**
```bash
# 确认配置文件存在
ls -la configs/config.yaml

# 确认工作目录正确
pwd

# 手动指定配置文件路径
./bin/server -config ./configs/config.yaml
```

**Docker 场景：**
```bash
# 进入容器确认文件是否被正确 COPY
docker exec -it <container> ls -la /app/configs/
```

---

### 7. 环境变量未生效

**排查：**
```bash
# 打印当前环境变量
env | grep APP_
env | grep DB_

# Docker 场景
docker exec -it <container> env | grep APP_
```

**注意：** 环境变量优先级高于 YAML。如果值不对，先检查是否有意外的环境变量。

---

## 性能排查

### pprof（Go 程序性能分析）

```go
// 在 main.go 中添加 pprof 端点（仅开发环境）
import _ "net/http/pprof"
go func() {
    http.ListenAndServe(":6060", nil)
}()
```

```bash
# CPU 分析（采样 30 秒）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 泄漏检测
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 火焰图（需安装 go-torch 或使用 pprof web 命令）
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30
```

### 基准测试分析

```bash
# 运行基准测试并输出内存分配
go test -bench=. -benchmem ./...

# 输出格式说明：
# BenchmarkXxx-8   1000000   1200 ns/op   256 B/op   3 allocs/op
#                  ↑执行次数  ↑每次耗时     ↑每次内存    ↑每次分配次数

# 对比两次基准测试（需安装 benchstat）
go test -bench=. -count=5 ./... > old.txt
# ... 修改代码 ...
go test -bench=. -count=5 ./... > new.txt
benchstat old.txt new.txt
```

---

## 监控检查

### 健康检查验证

```bash
# 基本检查
curl -s http://localhost:8080/health | jq .

# 预期输出：
# { "code": 0, "data": { "status": "ok", "version": "0.1.0" }, "message": "success" }

# Docker healthcheck 状态
docker inspect --format='{{.State.Health.Status}}' <container>
```

### 日志检查

```bash
# JSON 格式日志用 jq 过滤
docker-compose logs app | jq 'select(.level == "ERROR")'

# 按 request_id 追踪
docker-compose logs app | jq 'select(.request_id == "xxx")'
```
