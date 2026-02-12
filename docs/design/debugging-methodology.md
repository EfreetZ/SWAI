# 🔧 排障方法论 — Debugging & Troubleshooting Guide

> 贯穿所有阶段的通用排障能力建设

---

## 1. 三层排障模型（核心框架）

遇到故障时，按固定路径**逐层排查**，从外到内逐步缩小范围：

```
┌─────────────────────────────────────────────────────────┐
│ 第一层：资源瓶颈                                          │
│                                                          │
│   CPU → 内存 → 磁盘 IO → 网络                            │
│   工具：top / htop / iostat / iotop / netstat / ss       │
│                                                          │
│   判断：硬件资源是否已经到达极限？                           │
├─────────────────────────────────────────────────────────┤
│ 第二层：系统瓶颈                                          │
│                                                          │
│   Goroutine 数 → 锁竞争 → 队列堆积 → 连接池              │
│   工具：pprof / trace / dlv / runtime metrics            │
│                                                          │
│   判断：程序内部是否有逻辑瓶颈？                            │
├─────────────────────────────────────────────────────────┤
│ 第三层：架构瓶颈                                          │
│                                                          │
│   DB 是否单点？→ 缓存是否失效？→ 是否同步调用？             │
│   → 是否缺少降级？→ 流量是否合理分配？                      │
│   方法：架构审查 → 链路追踪 → 容量评估                      │
│                                                          │
│   判断：系统架构本身是否有设计缺陷？                         │
└─────────────────────────────────────────────────────────┘
```

> **80% 的系统问题，本质都是架构问题。**

---

## 2. 排障五步法（执行流程）

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 1.看监控  │ →  │ 2.查日志  │ →  │ 3.定瓶颈  │ →  │ 4.验假设  │ →  │ 5.复盘   │
│          │    │          │    │          │    │          │    │          │
│ CPU/Mem  │    │ 错误日志  │    │ pprof    │    │ 修改代码  │    │ 根因分析 │
│ IO/Net   │    │ 访问日志  │    │ trace    │    │ 灰度发布  │    │ 补全告警 │
│ Goroutine│    │ 慢日志    │    │ 火焰图    │    │ 观察指标  │    │ 写预案   │
└──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘
```

### 核心原则

- **先看全局，再看局部：** 先看系统级指标，再深入单个组件
- **先排除，再确认：** 通过二分法缩小范围
- **改一处，测一次：** 不要同时改多个地方
- **不猜测，看数据：** 所有判断基于监控和日志数据

---

## 3. Go 程序排障工具箱

### 3.1 pprof — 性能剖析

```go
// 在程序中开启 pprof
import _ "net/http/pprof"

func main() {
    // 开启 pprof HTTP 端点
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    // ... 业务代码
}
```

**常用命令：**

```bash
# CPU 分析（采样 30 秒）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析（当前堆内存）
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 分析（查看所有 goroutine 栈）
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 阻塞分析（哪里在等锁/channel）
go tool pprof http://localhost:6060/debug/pprof/block

# Mutex 竞争分析
go tool pprof http://localhost:6060/debug/pprof/mutex

# 生成火焰图（需要安装 graphviz）
go tool pprof -http=:8080 profile.pb.gz
```

**pprof 交互命令：**

```
(pprof) top 20        # 前 20 个 CPU 消耗函数
(pprof) list funcName # 查看函数逐行耗时
(pprof) web           # 在浏览器打开调用图
(pprof) svg           # 生成 SVG 图
```

### 3.2 trace — 运行时追踪

```bash
# 采集 5 秒 trace 数据
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# 打开 trace 可视化
go tool trace trace.out
```

**trace 能看到：**
- Goroutine 调度时序
- 网络/系统调用阻塞
- GC 暂停时间
- Goroutine 创建与销毁

### 3.3 dlv — Go 调试器

```bash
# 启动调试
dlv debug ./cmd/server

# 附加到运行中的进程
dlv attach <pid>

# 常用命令
(dlv) break main.go:42     # 设置断点
(dlv) continue              # 继续执行
(dlv) next                  # 下一行
(dlv) step                  # 进入函数
(dlv) print varName         # 打印变量
(dlv) goroutines            # 查看所有 goroutine
(dlv) goroutine <id>        # 切换到指定 goroutine
(dlv) stack                 # 查看调用栈
```

### 3.4 数据竞争检测

```bash
# 编译时开启竞争检测
go test -race ./...
go run -race main.go
go build -race -o server ./cmd/server
```

---

## 4. 系统级排障

### 4.1 CPU 排障

```bash
# 查看 CPU 使用率
top -p <pid>
htop

# Go 程序 CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 常见 CPU 问题：
# 1. 死循环 / 忙等待 → pprof 火焰图定位热点函数
# 2. GC 压力大 → trace 查看 GC 暂停时间
# 3. 序列化开销 → 换更高效的编解码（JSON → Protobuf）
```

### 4.2 内存排障

```bash
# 查看进程内存
ps aux | grep <process>
cat /proc/<pid>/status | grep -i vm

# Go 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 内存泄漏排查步骤：
# 1. 定时采集 heap profile
# 2. 用 pprof diff 对比两个时间点的 heap
go tool pprof -base heap1.pb.gz heap2.pb.gz

# 常见内存问题：
# 1. Goroutine 泄漏（每个 goroutine 最少 2KB 栈）
# 2. 大切片不释放（切片引用导致底层数组无法 GC）
# 3. sync.Pool 误用
# 4. string/[]byte 转换产生大量临时对象
```

### 4.3 IO 排障

```bash
# 磁盘 IO
iostat -xz 1       # 每秒刷新，显示设备利用率
iotop               # 按进程查看 IO

# 关键指标：
# - %util：磁盘利用率（> 80% 说明 IO 瓶颈）
# - await：IO 平均等待时间（> 10ms 偏高）
# - r/s, w/s：每秒读写次数

# 文件描述符
ls -la /proc/<pid>/fd | wc -l    # 查看 fd 数量
ulimit -n                         # 查看 fd 上限
```

### 4.4 网络排障

```bash
# 连接状态
ss -tnp | grep <port>       # TCP 连接状态
netstat -an | grep TIME_WAIT | wc -l   # TIME_WAIT 数量

# 抓包
tcpdump -i any port 8080 -w capture.pcap
# 然后用 Wireshark 分析

# DNS 解析
dig example.com
nslookup example.com

# 连通性
curl -v http://target:port/health
telnet target port

# 常见网络问题：
# 1. TIME_WAIT 过多 → 调整 tcp_tw_reuse / 连接池
# 2. 连接被拒绝 → 检查端口监听 / 防火墙
# 3. 连接超时 → 网络分区 / 路由问题
# 4. RST 重置 → 服务端崩溃 / 半开连接
```

---

## 5. 中间件排障

### 5.1 MySQL 排障路径

```
性能问题 → 1.查慢查询日志(slow_query_log)
         → 2.EXPLAIN 分析执行计划
         → 3.检查索引使用(type: ALL → 全表扫描)
         → 4.查看连接数(SHOW PROCESSLIST)
         → 5.检查锁等待(SHOW ENGINE INNODB STATUS)

关键命令:
  SHOW VARIABLES LIKE 'slow_query%';
  SHOW STATUS LIKE 'Threads%';
  SHOW ENGINE INNODB STATUS\G
  SELECT * FROM information_schema.INNODB_LOCK_WAITS;
```

### 5.2 Redis 排障路径

```
性能问题 → 1.INFO memory（内存使用）
         → 2.SLOWLOG GET 20（慢查询）
         → 3.redis-cli --bigkeys（大Key扫描）
         → 4.CLIENT LIST（连接数）
         → 5.INFO stats（命中率）

关键命令:
  INFO memory                    # 内存使用详情
  INFO stats                     # keyspace_hits / misses → 命中率
  SLOWLOG GET 20                 # 最近 20 条慢查询
  CLIENT LIST                    # 所有客户端连接
  DEBUG OBJECT <key>             # 查看 key 的编码和大小
  MEMORY USAGE <key>             # 精确查看 key 占用内存
  redis-cli --bigkeys            # 扫描大 Key
  redis-cli --hotkeys            # 扫描热 Key（需 LFU）
  redis-cli --latency            # 延迟检测
```

### 5.3 Kafka 排障路径

```
消费延迟 → 1.查 Consumer Lag（kafka-consumer-groups.sh）
         → 2.查 Broker IO（iostat）
         → 3.查 Producer 发送速率
         → 4.查 Consumer 处理耗时
         → 5.查 Partition 数据倾斜

关键命令:
  kafka-consumer-groups.sh --describe --group <group>
  kafka-topics.sh --describe --topic <topic>
  kafka-log-dirs.sh --describe --broker-list <broker>
```

---

## 6. Docker 环境排障

```bash
# 查看容器状态
docker ps -a
docker stats                   # 实时资源使用

# 查看容器日志
docker logs -f --tail 100 <container>

# 进入容器调试
docker exec -it <container> /bin/sh

# 查看容器网络
docker network ls
docker network inspect <network>

# 查看容器内进程
docker top <container>

# 导出容器文件系统（调查磁盘问题）
docker export <container> -o container.tar

# 查看镜像层（优化镜像大小）
docker history <image>

# 清理
docker system prune -af        # 清理所有未使用资源
```

---

## 7. 问题处理全流程模板

每个项目遇到问题时，按此模板记录：

```markdown
## 问题标题

### 1. 问题发现
- **时间：** 2024-xx-xx
- **现象：** 描述可观测的异常表现
- **影响范围：** 哪些功能/用户受影响
- **发现方式：** 监控告警 / 用户反馈 / 日志巡检

### 2. 问题分析
- **初步判断：** 可能的原因列表
- **排查过程：** 使用了哪些工具，看到了什么数据
- **根因定位：** 最终确认的根因

### 3. 问题解决
- **修复方案：** 代码改动 / 配置变更
- **影响评估：** 修复是否有副作用
- **回滚方案：** 如果修复有问题如何回滚

### 4. 问题测试
- **单元测试：** 新增测试用例覆盖此场景
- **回归测试：** 确认其他功能不受影响
- **压力测试：** 在高负载下验证修复

### 5. 问题优化
- **性能改进：** 是否有进一步优化空间
- **代码质量：** 是否需要重构

### 6. 问题监控
- **新增告警：** 补全监控覆盖
- **Dashboard：** 新增关键指标看板

### 7. 问题预防
- **规范更新：** 代码规范 / Review Checklist
- **自动化：** CI 增加什么检查
- **文档：** 更新排障手册
```

---

## 8. 每阶段排障能力要求

| 阶段 | 必须掌握的排障技能 |
|------|-------------------|
| Stage 0-1 | `go test -race`、pprof 基础、dlv 调试 |
| Stage 2-3 | 连接池监控、慢请求追踪、goroutine 泄漏排查 |
| Stage 4 | 磁盘 IO 分析、慢查询分析（EXPLAIN）、死锁排查 |
| Stage 5 | 内存分析、大 Key / 热 Key 排查、连接数监控 |
| Stage 6 | 吞吐瓶颈分析、Consumer Lag 排查、磁盘空间管理 |
| Stage 7 | 网络抓包、超时链路分析、连接泄漏排查 |
| Stage 8 | 分布式日志收集、网络分区诊断、一致性验证 |
| Stage 9 | 限流 / 熔断调优、配置热加载验证 |
| Stage 10 | Prometheus + Grafana 监控、链路追踪、告警规则 |
| Stage 11 | 全链路排障、容量评估、混沌测试 |
