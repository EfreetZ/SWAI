# 中间件分布式部署、高可用 & 线上故障排查手册

> 覆盖 MySQL / Redis / Kafka 三大核心中间件
>
> 定位：**生产级运维知识** + **面试高频场景**，与 Stage 4-6 设计文档配合使用

---

## 目录

- [一、MySQL 高可用 & 故障排查](#一mysql-高可用--故障排查)
- [二、Redis 高可用 & 故障排查](#二redis-高可用--故障排查)
- [三、Kafka 高可用 & 故障排查](#三kafka-高可用--故障排查)
- [四、通用排障 SOP](#四通用排障-sop)

---

# 一、MySQL 高可用 & 故障排查

## 1.1 分布式部署架构

### 主从复制（Master-Slave）

```
┌─────────┐    binlog     ┌─────────┐
│  Master  │ ──────────▶  │  Slave  │  （异步/半同步）
│  (读写)   │              │  (只读)  │
└─────────┘              └─────────┘
```

**核心原理：**
- Master 将变更写入 **binlog**（Binary Log）
- Slave 的 IO Thread 拉取 binlog 写入 **relay log**
- Slave 的 SQL Thread 重放 relay log

**三种复制模式对比：**

| 模式 | 数据安全 | 性能 | 适用场景 |
|------|---------|------|---------|
| 异步复制 | 可能丢数据 | 最高 | 读多写少、允许少量延迟 |
| 半同步复制 | 至少一个 Slave 确认 | 中等 | 生产环境主流 |
| 组复制（MGR） | Paxos 多数派确认 | 较低 | 金融级强一致 |

### 主主复制（Master-Master）

```
┌─────────┐  ◀──────▶  ┌─────────┐
│ Master A │            │ Master B │
│  (读写)   │            │  (读写)   │
└─────────┘            └─────────┘
      │                       │
      ▼                       ▼
┌─────────┐            ┌─────────┐
│ Slave A  │            │ Slave B  │
└─────────┘            └─────────┘
```

**注意事项：**
- 必须设置 `auto_increment_offset` 和 `auto_increment_increment` 避免主键冲突
- 推荐只写一个 Master，另一个做热备
- 双写场景极易产生数据不一致

### MySQL InnoDB Cluster（MGR + MySQL Router + MySQL Shell）

```
┌──────────────┐
│ MySQL Router │  ← 应用连接入口，自动故障转移
└──────┬───────┘
       │
┌──────▼───────┐
│   MGR 集群    │
│ ┌───┐┌───┐┌───┐│
│ │ P ││ S ││ S ││  P=Primary, S=Secondary
│ └───┘└───┘└───┘│
└──────────────┘
```

**Docker 部署示例：**

```yaml
# docker-compose-mysql-cluster.yml
version: "3.8"
services:
  mysql-primary:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_REPLICATION_MODE: master
    ports:
      - "3306:3306"
    command: >
      --server-id=1
      --log-bin=mysql-bin
      --binlog-format=ROW
      --gtid-mode=ON
      --enforce-gtid-consistency=ON
      --log-slave-updates=ON

  mysql-replica-1:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_REPLICATION_MODE: slave
      MYSQL_MASTER_HOST: mysql-primary
    command: >
      --server-id=2
      --log-bin=mysql-bin
      --binlog-format=ROW
      --gtid-mode=ON
      --enforce-gtid-consistency=ON
      --read-only=ON

  mysql-replica-2:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_REPLICATION_MODE: slave
      MYSQL_MASTER_HOST: mysql-primary
    command: >
      --server-id=3
      --log-bin=mysql-bin
      --binlog-format=ROW
      --gtid-mode=ON
      --enforce-gtid-consistency=ON
      --read-only=ON
```

---

## 1.2 线上真实故障场景 & 解决方案

### 场景 1：主从复制延迟

**现象：** 从库查询结果落后于主库，业务读到旧数据

**排查路径：**

```sql
-- 在从库执行
SHOW SLAVE STATUS\G

-- 关注字段：
-- Seconds_Behind_Master: 延迟秒数
-- Relay_Log_Space: relay log 大小
-- Exec_Master_Log_Pos vs Read_Master_Log_Pos: 执行与接收的差距
```

**根因分析：**

| 原因 | 现象 | 解决方案 |
|------|------|---------|
| 主库大事务 | `Exec_Master_Log_Pos` 长时间不变 | 拆分大事务，避免批量 UPDATE/DELETE |
| 从库单线程回放 | CPU 单核打满 | 开启并行复制 `slave_parallel_workers=4` |
| 从库机器性能差 | IO wait 高 | 升级从库配置 / SSD |
| DDL 阻塞 | 某个 DDL 执行很久 | 使用 pt-online-schema-change |
| 网络延迟 | `Relay_Log_Space` 增长慢 | 检查网络带宽 / 跨机房问题 |

**面试回答模板：**
> "主从延迟的排查，我会先 `SHOW SLAVE STATUS` 看 `Seconds_Behind_Master`，然后区分是**IO 线程慢**还是 **SQL 线程慢**。IO 慢看网络，SQL 慢看是否有大事务或 DDL 阻塞，必要时开启并行复制。"

---

### 场景 2：死锁

**现象：** 应用报错 `Deadlock found when trying to get lock`

**排查路径：**

```sql
-- 查看最近死锁信息
SHOW ENGINE INNODB STATUS\G

-- 关注 LATEST DETECTED DEADLOCK 段
-- 包含两个事务持有和等待的锁信息
```

```sql
-- 查看当前锁等待
SELECT * FROM information_schema.INNODB_LOCK_WAITS;
SELECT * FROM information_schema.INNODB_TRX;

-- MySQL 8.0+
SELECT * FROM performance_schema.data_lock_waits;
```

**经典死锁场景：**

```
-- 事务A                        -- 事务B
BEGIN;                          BEGIN;
UPDATE t SET v=1 WHERE id=1;    UPDATE t SET v=2 WHERE id=2;
-- 持有 id=1 的行锁              -- 持有 id=2 的行锁
UPDATE t SET v=1 WHERE id=2;    UPDATE t SET v=2 WHERE id=1;
-- 等待 id=2 → DEADLOCK!        -- 等待 id=1 → DEADLOCK!
```

**解决方案：**
1. **统一加锁顺序：** 所有事务按相同顺序操作行（如按 id 升序）
2. **减小事务粒度：** 缩短持锁时间
3. **降低隔离级别：** RR → RC 减少 Gap Lock
4. **加索引：** 避免全表扫描导致锁住过多行

**面试回答模板：**
> "死锁本质是循环等待。排查用 `SHOW ENGINE INNODB STATUS` 看死锁日志，定位两个事务分别持有和等待的锁。解决核心是统一加锁顺序、缩小事务、加索引避免锁放大。"

---

### 场景 3：慢查询导致 CPU 飙高

**现象：** MySQL CPU 使用率 > 90%，应用响应变慢

**排查路径：**

```sql
-- 1. 查看当前正在执行的查询
SHOW FULL PROCESSLIST;

-- 2. 开启慢查询日志
SET GLOBAL slow_query_log = ON;
SET GLOBAL long_query_time = 1;  -- 超过 1 秒记录

-- 3. 分析慢查询
-- 使用 pt-query-digest 分析慢查询日志
-- pt-query-digest /var/log/mysql/slow.log

-- 4. 对可疑 SQL 执行 EXPLAIN
EXPLAIN SELECT * FROM orders WHERE user_id = 123 AND status = 'paid';
```

**EXPLAIN 关键字段速查：**

| 字段 | 危险值 | 含义 |
|------|-------|------|
| type | ALL | 全表扫描 |
| type | index | 全索引扫描 |
| rows | > 10000 | 扫描行数过多 |
| Extra | Using filesort | 需要额外排序 |
| Extra | Using temporary | 需要临时表 |
| key | NULL | 未使用索引 |

**常见优化：**

| 问题 | 原因 | 优化 |
|------|------|------|
| `WHERE a = 1 AND b > 10 AND c = 'x'` | 联合索引 (a,b,c)，b 用范围后 c 失效 | 调整为 (a,c,b) |
| `SELECT *` | 无法使用覆盖索引 | 只查需要的列 |
| `ORDER BY rand()` | 全表排序 | 应用层随机 |
| 大分页 `LIMIT 100000, 10` | 扫描前 100000 行 | 延迟关联 / 游标分页 |

**大分页优化示例：**

```sql
-- 慢：扫描前 100000 行再丢弃
SELECT * FROM orders ORDER BY id LIMIT 100000, 10;

-- 快：延迟关联
SELECT * FROM orders
WHERE id > (SELECT id FROM orders ORDER BY id LIMIT 100000, 1)
ORDER BY id LIMIT 10;

-- 更快：游标分页（前端传上一页最后一条 id）
SELECT * FROM orders WHERE id > #{lastId} ORDER BY id LIMIT 10;
```

---

### 场景 4：磁盘写满导致主库宕机

**现象：** MySQL 无法写入，报错 `No space left on device`

**排查路径：**

```bash
# 查看磁盘空间
df -h

# 查看 MySQL 数据目录占用
du -sh /var/lib/mysql/*

# 常见大文件
# - binlog：未设置 expire_logs_days
# - undo log：长事务未提交
# - tmp 文件：大查询排序
# - 大表数据文件：.ibd
```

**解决方案：**

| 原因 | 解决 |
|------|------|
| binlog 堆积 | `PURGE BINARY LOGS BEFORE '2025-01-01'`；设置 `expire_logs_days=7` |
| undo log 膨胀 | 找出长事务 `SELECT * FROM INNODB_TRX ORDER BY trx_started`；kill 掉 |
| 临时文件 | 优化大查询；增加 `tmp_table_size` |
| 大表 | 历史数据归档；分库分表 |

---

### 场景 5：连接数耗尽

**现象：** 应用报错 `Too many connections`

**排查路径：**

```sql
-- 查看当前连接
SHOW STATUS LIKE 'Threads_connected';
SHOW STATUS LIKE 'Max_used_connections';
SHOW VARIABLES LIKE 'max_connections';

-- 按用户/Host 统计
SELECT user, host, COUNT(*) as conn_count
FROM information_schema.PROCESSLIST
GROUP BY user, host ORDER BY conn_count DESC;

-- 查找空闲连接
SELECT * FROM information_schema.PROCESSLIST
WHERE COMMAND = 'Sleep' AND TIME > 300;
```

**解决方案：**
1. **应用层：** 使用连接池，合理配置 `maxIdleConns` / `maxOpenConns`
2. **MySQL 侧：** 调大 `max_connections`；设置 `wait_timeout=300` 清理空闲连接
3. **引入 ProxySQL / MaxScale：** 连接复用 + 读写分离

---

## 1.3 MySQL 面试高频问题速查

| # | 问题 | 核心答案 |
|---|------|---------|
| 1 | 主从复制原理 | binlog → relay log → SQL Thread 回放 |
| 2 | 半同步 vs 异步 | 半同步等至少一个 Slave ACK，保证不丢数据 |
| 3 | GTID 的作用 | 全局事务 ID，简化故障转移和复制管理 |
| 4 | 如何解决主从延迟 | 并行复制 / 拆分大事务 / 升级从库 |
| 5 | 分库分表方案 | 水平分片（Hash/Range）；中间件（ShardingSphere / Vitess） |
| 6 | 如何选择分片键 | 高基数 + 查询频繁 + 均匀分布 |
| 7 | 分布式事务怎么办 | 2PC / TCC / Saga / 本地消息表 |
| 8 | Online DDL | MySQL 8.0 INSTANT / pt-osc / gh-ost |
| 9 | binlog 格式选择 | ROW（推荐，精确记录行变更）/ STATEMENT / MIXED |
| 10 | 如何不停机迁移 | 双写 → 同步历史 → 校验 → 切读 → 切写 |

---

# 二、Redis 高可用 & 故障排查

## 2.1 分布式部署架构

### Sentinel 哨兵模式

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│Sentinel 1│  │Sentinel 2│  │Sentinel 3│
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │              │              │
     └──────────────┼──────────────┘
                    │ 监控
            ┌───────▼───────┐
            │    Master     │
            │  (读写)        │
            └───────┬───────┘
           ┌────────┴────────┐
     ┌─────▼─────┐    ┌─────▼─────┐
     │  Slave 1  │    │  Slave 2  │
     │  (只读)    │    │  (只读)    │
     └───────────┘    └───────────┘
```

**核心机制：**
- Sentinel 定期 PING 所有节点
- **主观下线（SDOWN）：** 单个 Sentinel 认为节点挂了
- **客观下线（ODOWN）：** `quorum` 个 Sentinel 都认为挂了，触发故障转移
- **选举新 Master：** 按 `priority / offset / runid` 选出最优 Slave

**Docker 部署示例：**

```yaml
# docker-compose-redis-sentinel.yml
version: "3.8"
services:
  redis-master:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"

  redis-slave-1:
    image: redis:7-alpine
    command: redis-server --appendonly yes --replicaof redis-master 6379

  redis-slave-2:
    image: redis:7-alpine
    command: redis-server --appendonly yes --replicaof redis-master 6379

  sentinel-1:
    image: redis:7-alpine
    command: >
      redis-sentinel /etc/redis/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis/sentinel.conf

  sentinel-2:
    image: redis:7-alpine
    command: redis-sentinel /etc/redis/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis/sentinel.conf

  sentinel-3:
    image: redis:7-alpine
    command: redis-sentinel /etc/redis/sentinel.conf
    volumes:
      - ./sentinel.conf:/etc/redis/sentinel.conf
```

```conf
# sentinel.conf
sentinel monitor mymaster redis-master 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
sentinel parallel-syncs mymaster 1
```

### Redis Cluster 模式

```
┌──────────────────────────────────────────────┐
│              Redis Cluster (6 节点)            │
│                                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐      │
│  │Master 1 │  │Master 2 │  │Master 3 │      │
│  │slot 0-  │  │slot     │  │slot     │      │
│  │ 5460    │  │5461-    │  │10923-   │      │
│  │         │  │ 10922   │  │ 16383   │      │
│  └────┬────┘  └────┬────┘  └────┬────┘      │
│       │             │             │            │
│  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐      │
│  │Slave 1  │  │Slave 2  │  │Slave 3  │      │
│  └─────────┘  └─────────┘  └─────────┘      │
└──────────────────────────────────────────────┘
```

**核心概念：**
- **16384 个 slot：** 通过 `CRC16(key) % 16384` 决定 key 所在 slot
- **MOVED 重定向：** 客户端访问错误节点时返回 `MOVED slot ip:port`
- **ASK 重定向：** 数据迁移过程中的临时重定向
- **Gossip 协议：** 节点间通过 PING/PONG 交换集群状态

---

## 2.2 线上真实故障场景 & 解决方案

### 场景 1：缓存雪崩

**现象：** 大量 key 同时过期，请求全部打到数据库，DB 被打崩

**排查路径：**

```bash
# 抽样 TTL 分布
redis-cli --scan --pattern '*' | head -1000 | while read key; do
  echo "$key $(redis-cli TTL $key)"
done | sort -k2 -n

# 查看 QPS 突变
redis-cli INFO stats | grep instantaneous_ops_per_sec
```

**解决方案：**

| 方案 | 实现 | 适用场景 |
|------|------|---------|
| TTL 加随机偏移 | `TTL = base + rand(0, 300)` | **首选**，简单有效 |
| 永不过期 + 异步更新 | 后台线程定期刷新缓存 | 热点数据 |
| 多级缓存 | 本地缓存（Ristretto）→ Redis → DB | 超高并发 |
| 限流降级 | 识别到 DB 压力大时直接返回降级数据 | 兜底方案 |

---

### 场景 2：缓存击穿（热 Key 过期）

**现象：** 某个超热 key 过期瞬间，大量并发请求同时穿透到 DB

**解决方案：**

```go
// 互斥锁方案（singleflight 模式）
func GetWithMutex(ctx context.Context, key string) (string, error) {
    // 先查缓存
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }

    // 缓存未命中，抢分布式锁
    lockKey := "lock:" + key
    ok, _ := rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if ok {
        defer rdb.Del(ctx, lockKey)
        // 查 DB 并回填缓存
        val, err = queryDB(key)
        if err == nil {
            rdb.Set(ctx, key, val, randomTTL())
        }
        return val, err
    }

    // 未抢到锁，短暂等待后重试
    time.Sleep(50 * time.Millisecond)
    return rdb.Get(ctx, key).Result()
}
```

---

### 场景 3：缓存穿透（查不存在的数据）

**现象：** 恶意请求大量不存在的 key，每次都穿透到 DB

**解决方案：**

| 方案 | 优点 | 缺点 |
|------|------|------|
| 布隆过滤器 | 内存小，拦截效果好 | 有误判率，不支持删除 |
| 空值缓存 | 简单直接 | 占用缓存空间 |
| 参数校验 | 提前拦截非法请求 | 无法覆盖所有场景 |

```go
// 布隆过滤器 + 空值缓存组合
func GetWithBloom(ctx context.Context, key string) (string, error) {
    // 1. 布隆过滤器判断 key 是否可能存在
    if !bloomFilter.MightContain(key) {
        return "", ErrNotFound  // 一定不存在，直接返回
    }

    // 2. 查缓存
    val, err := rdb.Get(ctx, key).Result()
    if err == nil {
        if val == "<nil>" {
            return "", ErrNotFound  // 空值缓存命中
        }
        return val, nil
    }

    // 3. 查 DB
    val, err = queryDB(key)
    if err == ErrNotFound {
        // 缓存空值，短 TTL
        rdb.Set(ctx, key, "<nil>", 60*time.Second)
        return "", ErrNotFound
    }
    rdb.Set(ctx, key, val, randomTTL())
    return val, err
}
```

---

### 场景 4：大 Key 问题

**现象：** 某些 key 存储的 value 过大（如 Hash 有百万 field），导致：
- 读取慢，阻塞其他命令
- 删除时阻塞（DEL 是同步操作）
- 集群数据倾斜

**排查路径：**

```bash
# 扫描大 Key
redis-cli --bigkeys

# 精确查看某个 key 的内存占用
redis-cli MEMORY USAGE mykey

# 查看慢日志
redis-cli SLOWLOG GET 10
```

**解决方案：**

| 类型 | 解决方案 |
|------|---------|
| 大 String（>10KB） | 压缩（gzip）或拆分为多个小 key |
| 大 Hash（>5000 field） | 按范围拆分为多个 Hash：`user:info:{uid%100}` |
| 大 List | 按时间分片：`msg:2025-01-01`、`msg:2025-01-02` |
| 大 Set/ZSet | Hash 分桶或使用 SSCAN 分批处理 |
| 删除大 Key | **UNLINK**（异步删除）代替 DEL |

---

### 场景 5：Redis Cluster 脑裂

**现象：** 网络分区后，旧 Master 和新 Master 同时接受写入，分区恢复后数据丢失

```
分区前：                    分区后：
┌────────┐                 ┌────────┐  ← 网络隔离
│ Master │                 │旧Master│  客户端仍写入（即将丢失）
└────────┘                 └────────┘
                           ─ ─ ─ ─ ─  网络分区
┌────────┐                 ┌────────┐
│ Slave  │                 │新Master│  Sentinel 提升
└────────┘                 └────────┘
```

**解决方案：**

```conf
# redis.conf 关键配置
# 至少有 1 个 Slave 连接且延迟 < 10s 才允许写入
min-replicas-to-write 1
min-replicas-max-lag 10
```

**面试回答模板：**
> "脑裂的本质是网络分区导致出现两个 Master。解决方案是配置 `min-replicas-to-write`，如果 Master 发现连接的 Slave 数量不足，主动拒绝写入，牺牲可用性保证一致性。"

---

### 场景 6：内存突增 OOM

**排查路径：**

```bash
# 查看内存使用
redis-cli INFO memory
# used_memory_human     → 实际数据内存
# used_memory_rss_human → 操作系统分配的内存（含碎片）
# mem_fragmentation_ratio → 碎片率（>1.5 需关注）

# 内存分析
redis-cli --memkeys  # Redis 7.0+

# 碎片整理
redis-cli CONFIG SET activedefrag yes
```

**常见原因：**

| 原因 | 排查 | 解决 |
|------|------|------|
| 大量 Key 未设置过期 | `redis-cli DBSIZE` 对比预期 | 补设 TTL |
| 内存碎片 | `mem_fragmentation_ratio > 1.5` | 开启 `activedefrag` |
| 大 Key | `--bigkeys` 扫描 | 拆分 / 压缩 |
| 客户端 buffer 过大 | `CLIENT LIST` | 限制 `client-output-buffer-limit` |

---

## 2.3 Redis 面试高频问题速查

| # | 问题 | 核心答案 |
|---|------|---------|
| 1 | Sentinel 故障转移流程 | SDOWN → ODOWN → 选 leader Sentinel → 选新 Master → 通知切换 |
| 2 | Cluster 数据迁移 | MIGRATING/IMPORTING 状态 + ASK 重定向 |
| 3 | 为什么 16384 个 slot | Gossip 心跳包中包含 slot bitmap，16384 bit = 2KB 正好合适 |
| 4 | Redis 6.0 多线程 | IO 线程多线程读写，命令执行仍单线程 |
| 5 | 持久化 RDB vs AOF | RDB 快照（快但可能丢数据）；AOF 日志（安全但大）；混合模式推荐 |
| 6 | 分布式锁 Redlock | N 节点多数派加锁；争议：时钟漂移 + 网络延迟可能导致锁失效 |
| 7 | 热 Key 发现 | `--hotkeys`（LFU 策略下）/ Proxy 统计 / 客户端采样 |
| 8 | 数据一致性 | Cache Aside（旁路缓存）：先更新 DB 再删缓存 + 延迟双删 |
| 9 | Pipeline vs 事务 | Pipeline 减少网络 RTT；MULTI/EXEC 保证原子性但无回滚 |
| 10 | Stream vs Pub/Sub | Stream 支持消费组 + 持久化 + ACK，适合 MQ 场景 |

---

# 三、Kafka 高可用 & 故障排查

## 3.1 分布式部署架构

### Kafka 集群架构

```
┌──────────────────────────────────────────────────┐
│                  Kafka Cluster                     │
│                                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │Broker 1  │  │Broker 2  │  │Broker 3  │       │
│  │          │  │          │  │          │       │
│  │ Topic-A  │  │ Topic-A  │  │ Topic-A  │       │
│  │ P0(L)    │  │ P0(F)    │  │ P1(L)    │       │
│  │ P1(F)    │  │ P2(L)    │  │ P2(F)    │       │
│  └──────────┘  └──────────┘  └──────────┘       │
│                                                    │
│  L=Leader  F=Follower                              │
└──────────────────────────────────────────────────┘
         │                │
┌────────▼────────┐  ┌───▼─────────────┐
│   ZooKeeper /   │  │    KRaft        │
│   (旧架构)       │  │   (新架构3.0+)  │
└─────────────────┘  └─────────────────┘
```

**核心高可用机制：**

| 机制 | 说明 |
|------|------|
| **Partition 副本** | 每个 Partition 有 N 个副本（replication.factor=3） |
| **ISR（In-Sync Replicas）** | 与 Leader 保持同步的副本集合 |
| **Leader 选举** | ISR 中的 Follower 可以被选为新 Leader |
| **min.insync.replicas** | 最小 ISR 数量，低于此值 Producer 写入失败 |
| **acks=all** | Producer 等所有 ISR 确认，保证不丢消息 |

**不丢消息的黄金配置：**

```properties
# Producer 端
acks=all
retries=3
max.in.flight.requests.per.connection=1  # 严格有序场景

# Broker 端
min.insync.replicas=2
unclean.leader.election.enable=false     # 禁止非 ISR 成为 Leader

# Consumer 端
enable.auto.commit=false                 # 手动提交 offset
```

**Docker 部署示例（KRaft 模式，无 ZooKeeper）：**

```yaml
# docker-compose-kafka-cluster.yml
version: "3.8"
services:
  kafka-1:
    image: bitnami/kafka:3.6
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka-1:9093,2@kafka-2:9093,3@kafka-3:9093
      KAFKA_KRAFT_CLUSTER_ID: abcdefghijklmnopqrstuv
      KAFKA_CFG_DEFAULT_REPLICATION_FACTOR: 3
      KAFKA_CFG_MIN_INSYNC_REPLICAS: 2
    ports:
      - "9092:9092"

  kafka-2:
    image: bitnami/kafka:3.6
    environment:
      KAFKA_CFG_NODE_ID: 2
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka-1:9093,2@kafka-2:9093,3@kafka-3:9093
      KAFKA_KRAFT_CLUSTER_ID: abcdefghijklmnopqrstuv

  kafka-3:
    image: bitnami/kafka:3.6
    environment:
      KAFKA_CFG_NODE_ID: 3
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka-1:9093,2@kafka-2:9093,3@kafka-3:9093
      KAFKA_KRAFT_CLUSTER_ID: abcdefghijklmnopqrstuv
```

---

## 3.2 线上真实故障场景 & 解决方案

### 场景 1：Consumer Lag 持续增长（消费积压）

**现象：** Consumer Group 的 Lag 指标持续上涨，消费跟不上生产

**排查路径：**

```bash
# 查看 Consumer Group 状态
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group my-group

# 关注字段：
# CURRENT-OFFSET  → 已消费位置
# LOG-END-OFFSET  → 最新消息位置
# LAG             → 积压量
```

**根因分析 & 解决：**

| 原因 | 排查方法 | 解决方案 |
|------|---------|---------|
| 消费处理慢 | 消费端 CPU / RT 监控 | 优化处理逻辑；异步处理 |
| Partition 数不足 | Consumer 数 > Partition 数（空闲消费者） | 增加 Partition |
| 消费者频繁 Rebalance | Consumer 日志 `Revocation` | 增大 `max.poll.interval.ms`；减小 `max.poll.records` |
| GC 导致心跳超时 | Consumer GC 日志 | 优化 JVM / Go GC 参数 |
| 下游依赖慢（DB / RPC） | 下游 RT 监控 | 批量写入；引入本地缓冲 |

**紧急处理 — 临时扩容消费：**

```bash
# 1. 增加 Partition（不可回退！）
kafka-topics.sh --alter --topic my-topic --partitions 12

# 2. 部署更多 Consumer 实例（最多 = Partition 数）

# 3. 跳过消费（谨慎！仅应急）
kafka-consumer-groups.sh --reset-offsets \
  --group my-group --topic my-topic \
  --to-latest --execute
```

---

### 场景 2：消息丢失

**排查路径：发送端 → Broker → 消费端**

| 环节 | 丢失原因 | 解决 |
|------|---------|------|
| **Producer** | `acks=0/1`，Leader 挂了 | `acks=all` + `retries > 0` |
| **Producer** | 发送超时未重试 | 设置合理 `delivery.timeout.ms` |
| **Broker** | ISR 缩减后 unclean election | `unclean.leader.election.enable=false` |
| **Broker** | `min.insync.replicas=1` | 设为 2 |
| **Consumer** | 自动提交 offset 后处理失败 | 手动提交：处理成功后再 commit |
| **Consumer** | Rebalance 导致重复/丢失 | 实现幂等消费 |

**面试回答模板：**
> "保证不丢消息需要三端配合：Producer 端 `acks=all` + 重试；Broker 端 `min.insync.replicas=2` + 禁止 unclean 选举；Consumer 端手动提交 offset + 幂等消费。"

---

### 场景 3：Rebalance 风暴

**现象：** Consumer Group 频繁触发 Rebalance，消费停顿

**触发条件：**
1. Consumer 加入/离开 Group
2. `session.timeout.ms` 内未发送心跳（被踢出）
3. `max.poll.interval.ms` 内未调用 poll（处理太慢）
4. Topic 的 Partition 数变化
5. 订阅的 Topic 正则匹配到新 Topic

**解决方案：**

```properties
# Consumer 配置优化
session.timeout.ms=25000           # 默认 10s，适当调大
heartbeat.interval.ms=5000          # 默认 3s
max.poll.interval.ms=300000         # 默认 5min，按业务调整
max.poll.records=100                # 每次拉取少量，处理快

# 使用 CooperativeStickyAssignor（增量 Rebalance）
partition.assignment.strategy=org.apache.kafka.clients.consumer.CooperativeStickyAssignor
```

---

### 场景 4：Broker 磁盘写满

**排查路径：**

```bash
# 查看各 Topic 磁盘占用
kafka-log-dirs.sh --bootstrap-server localhost:9092 --describe

# 查看 retention 配置
kafka-configs.sh --describe --topic my-topic \
  --bootstrap-server localhost:9092
```

**解决方案：**

| 方案 | 命令 / 配置 |
|------|------------|
| 调小保留时间 | `retention.ms=86400000`（1天） |
| 调小保留大小 | `retention.bytes=10737418240`（10GB） |
| 清理指定 Topic | `kafka-topics.sh --delete --topic old-topic` |
| 手动删 Segment | 危险操作，仅紧急时使用 |

---

### 场景 5：Partition 数据倾斜

**现象：** 某些 Partition 的数据量远大于其他 Partition

**原因：**
- 消息 Key 分布不均（如 user_id 集中在某个范围）
- 自定义 Partitioner 逻辑有 bug

**排查 & 解决：**

```bash
# 查看各 Partition 大小
kafka-log-dirs.sh --describe --bootstrap-server localhost:9092 \
  | jq '.brokers[].logDirs[].partitions[]'

# 解决：
# 1. 更换分区策略（如引入随机后缀打散）
# 2. 使用 kafka-reassign-partitions 迁移分区
```

---

## 3.3 Kafka 面试高频问题速查

| # | 问题 | 核心答案 |
|---|------|---------|
| 1 | Kafka 为什么吞吐高 | 顺序写磁盘 + Page Cache + 零拷贝(sendfile) + 批量压缩 |
| 2 | ISR 机制 | 与 Leader 保持同步的副本集合；落后太多被移出 ISR |
| 3 | Controller 的作用 | 管理 Partition 状态 + Leader 选举 + 副本分配 |
| 4 | KRaft vs ZooKeeper | KRaft 3.0+ 内置共识，去除 ZK 依赖，简化运维 |
| 5 | 顺序消费怎么保证 | 同一 Partition 内有序；业务上用同一 Key 路由到同一 Partition |
| 6 | Exactly-once 怎么实现 | Producer 幂等（PID + SeqNum）+ 事务（atomic multi-partition writes） |
| 7 | 消费者数量 > Partition 数 | 多余的消费者空闲，不会分配到 Partition |
| 8 | 零拷贝原理 | `sendfile()` 系统调用，数据不经过用户态，直接内核态传输 |
| 9 | Partition 扩容影响 | 新 Partition 无历史数据；Hash 路由改变导致顺序性打破 |
| 10 | 消息重复消费 | Rebalance / offset 提交失败导致；需消费端幂等（唯一键 / 版本号） |

---

# 四、通用排障 SOP

## 线上故障响应流程

```
发现异常 → 确认影响范围 → 止血（降级/限流/切流量）→ 定位根因 → 修复 → 复盘
```

### 排障黄金三步

```
1. 看监控（Metrics）  → 发生了什么？
2. 查日志（Logging）  → 为什么发生？
3. 追链路（Tracing）  → 影响了哪些？
```

### 中间件通用排查 Checklist

| # | 检查项 | 命令 / 工具 |
|---|--------|------------|
| 1 | CPU / 内存 / 磁盘 / 网络 | `top` / `free -h` / `df -h` / `iftop` |
| 2 | 进程状态 | `ps aux` / `systemctl status` |
| 3 | 连接数 | `ss -tnp` / `netstat -an` |
| 4 | 慢查询 / 慢日志 | MySQL SLOWLOG / Redis SLOWLOG / Kafka metrics |
| 5 | 主从/副本状态 | 各中间件状态命令 |
| 6 | 磁盘 IO | `iostat -xz 1` / `iotop` |
| 7 | 网络延迟 | `ping` / `mtr` / `tcpdump` |
| 8 | GC / 内存泄漏 | Go pprof / JVM GC log |

---

## 面试故障场景万能回答框架

```
1. 现象描述：用户报 XXX，监控看到 XXX
2. 影响范围：影响了哪些服务/用户
3. 止血措施：做了什么紧急处理（降级/切流量/扩容）
4. 根因定位：通过 XXX 工具/方法 定位到 XXX
5. 修复方案：短期 XXX，长期 XXX
6. 复盘改进：监控完善 / 告警优化 / 预案建设
```

**示例：**
> "线上 Redis 集群突然报 OOM，影响订单服务缓存查询。紧急措施是开启 `maxmemory-policy allkeys-lru` 并扩容实例。排查发现某个 Hash Key 存了 200 万 field，是批量导入脚本写的，没有拆分。修复方案是拆分大 Key + 加入大 Key 告警监控。复盘后增加了 Key 大小的自动巡检脚本。"
