# 性能分析插件

包路径: `github.com/caicloud/nirvana/plugins/profiling`

性能分析插件添加与 `net/http/pprof` 一致的 API，用于取得服务运行时信息。

默认情况下，插件会添加四个 Descriptor：

1. /debug/pprof
2. /debug/pprof/profile
3. /debug/pprof/symbol
4. /debug/pprof/trace

前缀路径 `/debug/pprof` 可以通过 Path Configurer 修改。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- Path(path string) nirvana.Configurer
  - 设置路径前缀，默认值为 `/debug/pprof`
 
