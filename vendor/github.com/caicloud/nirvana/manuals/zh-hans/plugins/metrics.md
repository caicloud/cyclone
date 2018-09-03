# 监控指标插件

包路径: `github.com/caicloud/nirvana/plugins/metrics`

监控指标插件基于 Prometheus，提供了一个 API 用于暴露服务端指标。

启用插件后，可以直接向 prometheus 包注册指标。采集端可以通过 `/metrics` 采集指标数据。

API 路径 `/metrics` 可以通过 Path Configurer 修改。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- Path(path string) nirvana.Configurer
  - 设置 API 路径，默认值为 `/metrics`
- Namespace(ns string) nirvana.Configurer
  - 设置 Prometheus Namespace
 
