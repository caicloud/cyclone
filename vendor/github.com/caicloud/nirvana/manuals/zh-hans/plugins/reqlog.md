# 请求日志插件

包路径: `github.com/caicloud/nirvana/plugins/reqlog`

请求日志插件会添加一个在 `/` 上的中间件，用于打印所有路由匹配成功的请求的日志。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- Default() nirvana.Configurer
  - 启用插件并使用默认配置
- Logger(l log.Logger) nirvana.Configurer
  - 设置 Logger
- DoubleLog(enable bool) nirvana.Configurer
  - 启用或关闭双重日志，即请求开始一条日志，请求结束一条日志
- SourceAddr(enable bool) nirvana.Configurer
  - 启用或关闭显示源地址
- RequestID(enable bool) nirvana.Configurer
  - 启用或关闭显示请求 ID 
- RequestIDKey(key string) nirvana.Configurer
  - 设置请求 ID 的 key，默认为 `X-Request-ID`

