# 请求追踪插件

包路径: `github.com/caicloud/nirvana/plugins/tracing`

请求追踪插件基于 OpenTracing 接口实现了请求跟踪机制，会在 `/` 上添加一个中间件用于提取 Tracing 信息。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- CustomTracer(tracer opentracing.Tracer) nirvana.Configurer
  - 使用自定义的 Tracer
- DefaultTracer(serviceName string, agentHostPort string) nirvana.Configurer
  - 使用默认的 Tracer
- AddHook(hook Hook) nirvana.Configurer
  - 设置请求 Hook


Hook 接口：
```go
// Hook allows you to custom information for span.
type Hook interface {
	// Exec before request processing
	Before(ctx context.Context, span opentracing.Span)
	// Exec after request processing
	After(ctx context.Context, span opentracing.Span)
}
```
 
