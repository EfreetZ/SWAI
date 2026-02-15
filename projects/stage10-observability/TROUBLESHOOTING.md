# Stage 10 排障

## 指标为空
- 检查是否调用了中间件埋点。
- 检查 /metrics 是否可访问。

## Trace 丢失
- 检查下游请求是否透传 trace_id。
- 检查日志字段是否输出 trace_id/span_id。

## 告警过多
- 提高阈值并引入 for 窗口。
- 区分 warning 与 critical。
