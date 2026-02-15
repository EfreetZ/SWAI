# Stage 7 — mini-rpc

对应设计文档：`docs/design/stage7-mini-rpc.md`

## 已实现能力（精简）

- 自定义二进制协议（Header + Payload + Metadata）
- JSON Codec
- RPC Server（服务注册 + 反射调用 + 中间件）
- RPC Client（同步调用 + 简单连接池）
- 注册中心（MemoryRegistry）
- 负载均衡（RoundRobin / Random）
- 容错（Timeout / Retry / CircuitBreaker）
- 监控与告警

## 快速开始

```bash
make test
make race
make bench
make run-server
```

服务默认地址：`127.0.0.1:18080`
