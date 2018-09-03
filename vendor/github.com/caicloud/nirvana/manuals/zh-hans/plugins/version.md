# 版本信息插件

包路径: `github.com/caicloud/nirvana/plugins/profiling`

版本信息插件提供一个 API 返回服务的版本信息。API 路径默认为 `/version`。

插件 Configurer：
- Disable() nirvana.Configurer
  - 关闭插件
- Path(path string) nirvana.Configurer
  - 设置 API 路径，默认值为 `/version`
- Name(name string) nirvana.Configurer
  - 设置服务名称
- Version(version string) nirvana.Configurer
  - 设置服务版本号
- Hash(hash string) nirvana.Configurer
  - 设置服务 hash 值。一般情况下可以设置为代码的 commit 值
- Description(description string) nirvana.Configurer
  - 设置服务的描述
 
