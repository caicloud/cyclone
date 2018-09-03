# 配置器机制

包路径: `github.com/caicloud/nirvana`

Nirvana 在启动服务之前需要先创建一个配置，这个配置涵盖了启动服务过程所需要的所有信息。

Nirvana 的 Config 实现了 Configurer 机制，用于单独配置每一项信息：

```go
// Config describes configuration of server.
type Config struct {
	...
}

// Configure configs by configurers. It panics if an error occurs or config is locked.
func (c *Config) Configure(configurers ...Configurer) *Config {
	...
}

// Configurer is used to configure server config.
type Configurer func(c *Config) error
 
```

## Nirvana 提供的 Configurers

### IP(ip string) Configurer

设置监听的 IP 地址。

### Port(port uint16) Configurer

设置监听的端口。

### TLS(certFile, keyFile string) Configurer

设置 TLS 证书和密钥。

### Logger(logger log.Logger) Configurer

设置 Server 在运行过程中使用的 Logger，用于输出错误。

### Filter(filters ...service.Filter) Configurer

添加 Filter。

### Modifier(modifiers ...service.DefinitionModifier) Configurer

添加 Modifier。

### Descriptor(descriptors ...definition.Descriptor) Configurer

添加 API 描述。所有的 API 都通过这个 Configurer 添加到 Nirvana 的 Server 里。

