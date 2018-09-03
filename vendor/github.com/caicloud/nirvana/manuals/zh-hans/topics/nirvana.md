# nirvana 包

nirvana 包在根目录中，实现了 Nirvana Server 和插件系统。这个包放置在根目录中是因为这个包是 Nirvana 提供的用于生成 API Server 的顶级包，而且其依赖的所有包只来自 Nirvana 自身和标准库（config 包依赖了 nirvana 包和其他第三方的包，实际上是一个借助了 Nirvana 和第三方功能的扩展）。


Server 接口如下：
```go
// Server is a complete API server.
// The server contains a router to handle all requests form clients.
type Server interface {
	// Serve starts to listen and serve requests.
	// The method won't return except an error occurs.
	Serve() error
	// Shutdown gracefully shuts down the server without interrupting any
	// active connections.
	Shutdown(ctx context.Context) error
	// Builder create a service builder for current server. Don't use this method directly except
	// there is a special server to hold http services. After server shutdown, clean resources via
	// returned cleaner.
	// This method always returns same builder until cleaner is called. Then it will
	// returns new one.
	Builder() (builder service.Builder, cleaner func() error, err error)
}
```

创建 Server 需要使用 Config：
```go
// Config describes configuration of server.
type Config struct {
	...
}

// Locked checks if the config is locked.
func (c *Config) Locked() bool

// IP returns listenning ip.
func (c *Config) IP() string

// Port returns listenning port.
func (c *Config) Port() uint16

// Logger returns logger.
func (c *Config) Logger() log.Logger

// Configurer is used to configure server config.
type Configurer func(c *Config) error

// Configure configs by configurers. It panics if an error occurs or config is locked.
func (c *Config) Configure(configurers ...Configurer) *Config

// Config gets external config by name. This method is for plugins.
func (c *Config) Config(name string) interface{}

// Set sets external config by name. This method is for plugins.
// Set a nil config will delete it.
func (c *Config) Set(name string, config interface{})

// NewServer creates a nirvana server. After creation, don't modify
// config. Also don't create another server with current config.
func NewServer(c *Config) Server
```
在 Config 中，存在一些 Server 级别的配置，这些配置是针对当前服务的。而对应的 Configurer 也在当前包中。如果需要对配置进行扩展，增强 Server 功能，则可以增加相应字段，否则应该使用插件机制增加功能。

在 Config 中可以看到 Config 和 Set 方法，这两个方法就是提供给插件允许插件设置自身的配置的。
```go
// ConfigInstaller is used to install config to service builder.
type ConfigInstaller interface {
	// Name is the external config name.
	Name() string
	// Install installs stuffs before server starting.
	Install(builder service.Builder, config *Config) error
	// Uninstall uninstalls stuffs after server terminating.
	Uninstall(builder service.Builder, config *Config) error
}
```
注册了插件之后，在服务启动的时候，会遍历所有插件配置，然后调用插件的 Install 方法安装插件。

