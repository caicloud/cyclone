# 插件机制

Nirvana 的 Config 除了使用 Configurer 配置基本信息以外，还提供了插件机制。

插件接口：
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

插件实现了这个接口之后，即可通过 nirvana 包提供的方法进行注册：

```go
func RegisterConfigInstaller(ci ConfigInstaller)
```

一般情况下，插件应该通过插件 package 的 `init()` 进行注册。然后提供相应的 Configurer 在 Nirvana 的 Config 中添加插件配置。

## Plugin Framework

一个基本的插件框架如下：
```go
func init() {
	// Register your config installer into nirvana.
	nirvana.RegisterConfigInstaller(&pluginInstaller{})
}

// ExternalConfigName is the external config name for your plugin. Please ensure that the
// name is unique and won't conflict with other plugins.
const ExternalConfigName = "pluginName"

type pluginInstaller struct{}

// Name is the external config name.
func (i *pluginInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder. You can get plugin config from nirvana config. Then
// install/initialize what you need.
func (i *pluginInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {...}

// Uninstall uninstalls stuffs after server terminating.
func (i *pluginInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {...)

// ConfigA configures fieldA. Be careful, you should get/save plugin config into nirvana config
// by `c.Config(ExternalConfigName)`/`c.Set(ExternalConfigName, cfg)` rather than a global
// plugin config.
func ConfigA(fieldA FieldType) nirvana.Configurer {...}

// ConfigB configures fieldB.
func ConfigB() nirvana.Configurer {...}

// Disable returns a configurer to disable current plugin for a certain nirvana server.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		// Set to nil will delete plugin config from nirvana config.
		c.Set(ExternalConfigName, nil)
		return nil
	}
}
```
