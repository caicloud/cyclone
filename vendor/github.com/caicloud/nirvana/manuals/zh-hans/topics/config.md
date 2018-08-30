# config 包

config 包利用 `github.com/spf13/cobra` 和 `github.com/spf13/viper` 实现了 Nirvana Command 和配置读取，为 Nirvana 服务启动提供了命令行支持。

NirvanaCommand 接口如下：
```go
// NirvanaCommand is a nirvana command.
type NirvanaCommand interface {
	// EnablePlugin enables plugins.
	EnablePlugin(plugins ...Plugin) NirvanaCommand
	// AddOption will fill up options from flags/ENV/config after executing.
	// A non-empty prefix is recommended. It's used to divide option namespaces.
	AddOption(prefix string, options ...CustomOption) NirvanaCommand
	// Add adds a field by key.
	// If you don't have any struct to describe an option, you can use the method to
	// add a single field into nirvana command.
	// `pointer` must be a pointer to golang basic data type (e.g. *int, *string).
	// `key` must a config key. It's like 'nirvana.ip' and 'myconfig.name.firstName'.
	// The key will be converted to flag and env (e.g. --nirvana-ip and NIRVANA_IP).
	// If you want a short flag for the field, you can only set a one-char string.
	// `desc` describes the field.
	Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaCommand
	// Execute runs nirvana server.
	Execute(descriptors ...definition.Descriptor) error
	// ExecuteWithConfig runs nirvana server from a custom config.
	ExecuteWithConfig(cfg *nirvana.Config) error
	// Command returns a command for command.
	Command(cfg *nirvana.Config) *cobra.Command
	// SetHook sets nirvana command hook.
	SetHook(hook NirvanaCommandHook)
	// Hook returns nirvana command hook.
	Hook() NirvanaCommandHook
}
```

NirvanaCommand 扩展了 nirvana 包的插件能力：
```
// CustomOption must be a pointer to struct.
//
// Here is an example:
//   type Option struct {
//       FirstName string `desc:"Desc for First Name"`
//       Age       uint16 `desc:"Desc for Age"`
//   }
// The struct has two fields (with prefix example):
//   Field       Flag                   ENV                  Key (In config file)
//   FirstName   --example-first-name   EXAMPLE_FIRST_NAME   example.firstName
//   Age         --example-age          EXAMPLE_AGE          example.age
// When you execute command with `--help`, you can see the help doc of flags and
// descriptions (From field tag `desc`).
//
// The priority is:
//   Flag > ENV > Key > The value you set in option
type CustomOption interface{}
 
// Plugin is for plugins to collect configurations
type Plugin interface {
	// Name returns plugin name.
	Name() string
	// Configure configures nirvana config via current options.
	Configure(cfg *nirvana.Config) error
}
```

Nirvana Command 要求每个插件提供一个 Option，并且实现 Plugin 接口。用户在 Comamnd 中传递 Option 来启用插件，并且将插件 Option 中的公开字段根据一定的规则（规则参考上面的注释）从 flag，环境变量，配置文件中读取。这样可以方便用户通过外部选项来改变运行时的行为。

由于 Nirvana Config 服务配置的特殊性，config 包提供了一个 Option 来表达这些配置：
```go
// Option contains basic configurations of nirvana.
type Option struct {
	// IP is the IP to listen.
	IP string `desc:"Nirvana server listening IP"`
	// Port is the port to listen.
	Port uint16 `desc:"Nirvana server listening Port"`
	// Key is private key for HTTPS.
	Key string `desc:"TLS private key (PEM format) for HTTPS"`
	// Cert is certificate for HTTPS.
	Cert string `desc:"TLS certificate (PEM format) for HTTPS"`
}
```

除了插件 Option 以外，config 包会从以下文件列表中读取配置文件：
```
目录：
  ./
  ./config/
  {ExecutableFilePath}/
  {ExecutableFilePath}/config/
  /etc/nirvana/

配置文件名：
  nirvana.yaml
  nirvana.toml
  nirvana.json
```

如果读取到配置文件，那么除了使用 Option 接收配置以外，还可以通过一些帮助方法获取配置：
```go
// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key
func IsSet(key string) bool

// Set sets the value for the key in the override regiser.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{})

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{}

// GetBool returns the value associated with the key as a bool.
func GetBool(key string) bool

// GetDuration returns the value associated with the key as a time.Duration.
func GetDuration(key string) time.Duration

// GetFloat32 returns the value associated with the key as a float32.
func GetFloat32(key string) float32

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64

// GetInt returns the value associated with the key as a int.
func GetInt(key string) int

// GetInt8 returns the value associated with the key as a int.
func GetInt8(key string) int8

// GetInt16 returns the value associated with the key as a int.
func GetInt16(key string) int16

// GetInt32 returns the value associated with the key as a int32.
func GetInt32(key string) int32

// GetInt64 returns the value associated with the key as a int64.
func GetInt64(key string) int64

// GetString returns the value associated with the key as a string.
func GetString(key string) string

// GetStringSlice returns the value associated with the key as a []string.
func GetStringSlice(key string) []string

// GetUint returns the value associated with the key as a uint.
func GetUint(key string) uint

// GetUint8 returns the value associated with the key as a uint.
func GetUint8(key string) uint8

// GetUint16 returns the value associated with the key as a uint.
func GetUint16(key string) uint16

// GetUint32 returns the value associated with the key as a uint32.
func GetUint32(key string) uint32

// GetUint64 returns the value associated with the key as a uint64.
func GetUint64(key string) uint64
```

**注：如果在 nirvana 包中对 Config 进行了扩展，涉及到字段的改变，也需要在这个包中修改 Option 和相应的功能。**
