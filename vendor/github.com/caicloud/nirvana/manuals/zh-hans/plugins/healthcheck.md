# 健康检查插件

包路径: `github.com/caicloud/nirvana/plugins/healthcheck`

健康检查插件提供一个 API 返回服务当前是否健康。API 默认路径为 `/healthz`。

插件提供了一个函数接口：
```
type HealthChecker func(ctx context.Context) error
```
如果服务正常，则 checker 应该返回 nil。如果服务异常，则返回相应的错误。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- Path(path string) nirvana.Configurer
  - 设置 API 路径，默认值为 `/healthz`
- Checker(checker HealthChecker) nirvana.Configurer
  - 设置 Checker 用于检查服务是否正常。
 
