# 系统日志插件

包路径: `github.com/caicloud/nirvana/plugins/logger`

系统日志插件是一个伪插件。这个插件本身没有按照 Plugin Framework 编写，只是为了通过 config 包的 Command 暴露 Flag。

这个插件暴露三个选项：
```
// Option contains basic configurations of logger.
type Option struct {
	// Debug is logger level.
	Debug bool `desc:"Debug mode. Output all logs"`
	// Level is logger level.
	Level int32 `desc:"Log level. This field is no sense if debug is enabled"`
	// OverrideGlobal modifies nirvana global logger.
	OverrideGlobal bool `desc:"Override global logger"`
}
```
启用 Debug 模式后，Level 就无效。如果 OverrideGlobal 为 true，那么除了设置当前 Server 的 logger 以外，还会设置全局的 logger。
