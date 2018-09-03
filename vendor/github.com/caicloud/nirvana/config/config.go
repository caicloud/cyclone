/*
Copyright 2018 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/utils/printer"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

// NewDefaultOption creates a default option.
func NewDefaultOption() *Option {
	return &Option{
		IP:   "",
		Port: 8080,
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ""
}

// Configure configures nirvana config via current option.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		nirvana.IP(p.IP),
		nirvana.Port(p.Port),
		nirvana.TLS(p.Key, p.Cert),
	)
	return nil
}

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

// NewDefaultNirvanaCommand creates a nirvana command with default option.
func NewDefaultNirvanaCommand() NirvanaCommand {
	return NewNirvanaCommand(nil)
}

// NewNirvanaCommand creates a nirvana command. Nil option means default option.
func NewNirvanaCommand(option *Option) NirvanaCommand {
	return NewNamedNirvanaCommand("server", option)
}

// NewNamedNirvanaCommand creates a nirvana command with an unique name.
// Empty name means `server`. Nil option means default option.
func NewNamedNirvanaCommand(name string, option *Option) NirvanaCommand {
	if name == "" {
		name = "server"
	}
	if option == nil {
		option = NewDefaultOption()
	}
	cmd := &command{
		name:    name,
		option:  option,
		plugins: []Plugin{},
		fields:  map[string]*configField{},
		hook:    &NirvanaCommandHookFunc{},
	}
	cmd.EnablePlugin(cmd.option)
	return cmd
}

type configField struct {
	pointer     interface{}
	desired     interface{}
	key         string
	env         string
	shortFlag   string
	longFlag    string
	description string
}

type command struct {
	name    string
	option  *Option
	plugins []Plugin
	fields  map[string]*configField
	hook    NirvanaCommandHook
}

// EnablePlugin enables plugins.
func (s *command) EnablePlugin(plugins ...Plugin) NirvanaCommand {
	s.plugins = append(s.plugins, plugins...)
	for _, plugin := range plugins {
		s.AddOption(plugin.Name(), plugin)
	}
	return s
}

func walkthrough(index []int, typ reflect.Type, f func(index []int, field reflect.StructField)) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			walkthrough(append(index, i), field.Type, f)
		} else {
			if field.Name != "" && field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
				f(append(index, i), field)
			}
		}
	}
}

// AddOption will fill up options from config/ENV/flags after executing.
func (s *command) AddOption(prefix string, options ...CustomOption) NirvanaCommand {
	if prefix != "" {
		prefix += "."
	}
	for _, opt := range options {
		val := reflect.ValueOf(opt)
		typ := val.Type()
		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			panic(errors.InternalServerError.Error("${type} is not a pointer to struct", typ.String()))
		}
		if val.IsNil() {
			panic(errors.InternalServerError.Error("${type} should not be nil", typ.String()))
		}
		val = val.Elem()
		typ = val.Type()
		walkthrough([]int{}, typ, func(index []int, field reflect.StructField) {
			ptr := val.FieldByIndex(index).Addr().Interface()
			s.Add(ptr, prefix+field.Name, "", field.Tag.Get("desc"))
		})
	}
	return s
}

// Add adds a field by key.
func (s *command) Add(pointer interface{}, key string, shortFlag string, desc string) NirvanaCommand {
	if pointer == nil || reflect.ValueOf(pointer).IsNil() {
		panic(errors.InternalServerError.Error("pointer of ${key} should not be nil", key))
	}

	keyParts := splitKey(key)
	cf := &configField{
		pointer:     pointer,
		shortFlag:   shortFlag,
		description: desc,
	}
	for i, parts := range keyParts {
		if i > 0 {
			cf.longFlag += "-"
		}
		cf.longFlag += strings.Join(parts, "-")
	}
	cf.longFlag = strings.ToLower(cf.longFlag)

	nameParts := splitKey(s.name)
	nameParts = append(nameParts, keyParts...)
	for i, parts := range nameParts {
		if i > 0 {
			cf.key += "."
		}
		for j, part := range parts {
			if j == 0 {
				part = strings.ToLower(part)
			}
			cf.key += part
		}
	}
	for i, parts := range nameParts {
		if i > 0 {
			cf.env += "_"
		}
		cf.env += strings.Join(parts, "_")
	}
	cf.env = strings.ToUpper(cf.env)
	if _, ok := s.fields[cf.key]; ok {
		panic(errors.InternalServerError.Error("${key} has been registered", cf.key))
	}
	s.fields[cf.key] = cf
	return s
}

// Execute runs nirvana server.
func (s *command) Execute(descriptors ...definition.Descriptor) error {
	cfg := nirvana.NewDefaultConfig()
	if len(descriptors) > 0 {
		cfg.Configure(nirvana.Descriptor(descriptors...))
	}
	return s.ExecuteWithConfig(cfg)
}

// ExecuteWithConfig runs nirvana server from a custom config.
func (s *command) ExecuteWithConfig(cfg *nirvana.Config) error {
	return s.Command(cfg).Execute()
}

// Command returns a command for nirvana.
func (s *command) Command(cfg *nirvana.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   s.name,
		Short: "Starting a nirvana server to handle http requests",
		Run: func(cmd *cobra.Command, args []string) {
			fs := cmd.Flags()
			// Restore configs.
			for _, f := range s.fields {
				val := reflect.ValueOf(f.pointer).Elem()
				if f.desired != nil && !fs.Lookup(f.longFlag).Changed {
					val.Set(reflect.ValueOf(f.desired))
				}
				Set(f.key, val.Interface())
			}
			if err := s.hook.PreConfigure(cfg); err != nil {
				cfg.Logger().Fatal(err)
			}
			for _, plugin := range s.plugins {
				if err := plugin.Configure(cfg); err != nil {
					cfg.Logger().Fatalf("Failed to install plugin %s: %s", plugin.Name(), err.Error())
				}
			}
			if err := s.hook.PostConfigure(cfg); err != nil {
				cfg.Logger().Fatal(err)
			}
			server := nirvana.NewServer(cfg)
			if err := s.hook.PreServe(cfg, server); err != nil {
				cfg.Logger().Fatal(err)
			}
			cfg.Logger().Infof("Listening on %s:%d", cfg.IP(), cfg.Port())
			if err := s.hook.PostServe(cfg, server, server.Serve()); err != nil {
				cfg.Logger().Fatal(err)
			}
		},
	}
	s.registerFields(cmd.Flags())
	p := printer.NewTable(30)
	rows := make([][]interface{}, 0, len(s.fields))
	for _, f := range s.fields {
		rows = append(rows, []interface{}{0, f.key, f.env, "--" + f.longFlag, f.desired})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][1].(string) < rows[j][1].(string)
	})
	p.AddRow("", "Config Key", "ENV", "Flag", "Current Value")
	for i, row := range rows {
		row[0] = i + 1
		p.AddRow(row...)
	}
	cmd.Long = "ConfigKey-ENV-Flag Mapping Table\n\n" + p.String()
	return cmd
}

func (s *command) registerFields(fs *pflag.FlagSet) {
	for _, f := range s.fields {
		var value, envValue interface{} = nil, nil
		env := os.Getenv(f.env)
		switch v := f.pointer.(type) {
		case *uint8:
			if IsSet(f.key) {
				value = GetUint8(f.key)
			}
			if env != "" {
				envValue = cast.ToUint8(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Uint8VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *uint16:
			if IsSet(f.key) {
				value = GetUint16(f.key)
			}
			if env != "" {
				envValue = cast.ToUint16(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Uint16VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *uint32:
			if IsSet(f.key) {
				value = GetUint32(f.key)
			}
			if env != "" {
				envValue = cast.ToUint32(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Uint32VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *uint64:
			if IsSet(f.key) {
				value = GetUint64(f.key)
			}
			if env != "" {
				envValue = cast.ToUint64(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Uint64VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *uint:
			if IsSet(f.key) {
				value = GetUint(f.key)
			}
			if env != "" {
				envValue = cast.ToUint(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.UintVarP(v, f.longFlag, f.shortFlag, *v, f.description)

		case *int8:
			if IsSet(f.key) {
				value = GetInt8(f.key)
			}
			if env != "" {
				envValue = cast.ToInt8(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Int8VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *int16:
			if IsSet(f.key) {
				value = GetInt16(f.key)
			}
			if env != "" {
				envValue = cast.ToInt16(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Int16VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *int32:
			if IsSet(f.key) {
				value = GetInt32(f.key)
			}
			if env != "" {
				envValue = cast.ToInt32(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Int32VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *int64:
			if IsSet(f.key) {
				value = GetInt64(f.key)
			}
			if env != "" {
				envValue = cast.ToInt64(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Int64VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *int:
			if IsSet(f.key) {
				value = GetInt(f.key)
			}
			if env != "" {
				envValue = cast.ToInt(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.IntVarP(v, f.longFlag, f.shortFlag, *v, f.description)

		case *float32:
			if IsSet(f.key) {
				value = GetFloat32(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat32(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Float32VarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *float64:
			if IsSet(f.key) {
				value = GetFloat64(f.key)
			}
			if env != "" {
				envValue = cast.ToFloat64(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.Float64VarP(v, f.longFlag, f.shortFlag, *v, f.description)

		case *string:
			if IsSet(f.key) {
				value = GetString(f.key)
			}
			if env != "" {
				envValue = cast.ToString(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.StringVarP(v, f.longFlag, f.shortFlag, *v, f.description)
		case *[]string:
			if IsSet(f.key) {
				value = GetStringSlice(f.key)
			}
			if env != "" {
				envValue = cast.ToStringSlice(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.StringSliceVarP(v, f.longFlag, f.shortFlag, *v, f.description)

		case *bool:
			if IsSet(f.key) {
				value = GetBool(f.key)
			}
			if env != "" {
				envValue = cast.ToBool(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.BoolVarP(v, f.longFlag, f.shortFlag, *v, f.description)

		case *time.Duration:
			if IsSet(f.key) {
				value = GetDuration(f.key)
			}
			if env != "" {
				envValue = cast.ToDuration(env)
			}
			f.desired = chooseValue(value, envValue, *v)
			fs.DurationVarP(v, f.longFlag, f.shortFlag, *v, f.description)

		default:
			panic(errors.InternalServerError.Error("unrecognized type ${type} for ${key}", reflect.TypeOf(f.pointer).String(), f.key))
		}
	}
}

// splitKey splits key to parts.
func splitKey(key string) [][]string {
	fieldParts := strings.Split(key, ".")
	nameParts := make([][]string, len(fieldParts))
	for index, field := range fieldParts {
		parts := []string{}
		lastIsCapital := true
		lastIndex := 0
		for i, char := range field {
			if char >= '0' && char <= '9' {
				// Numbers inherit last char.
				continue
			}
			currentIsCapital := char >= 'A' && char <= 'Z'
			if i > 0 && lastIsCapital != currentIsCapital {
				end := 0
				if currentIsCapital {
					end = i
				} else {
					end = i - 1
				}
				if end > lastIndex {
					parts = append(parts, field[lastIndex:end])
					lastIndex = end
				}
			}
			lastIsCapital = currentIsCapital
		}
		if lastIndex < len(field) {
			parts = append(parts, field[lastIndex:])
		}
		nameParts[index] = parts
	}
	return nameParts
}

// chooseValue chooses expected value form multiple configurations.
// Priority: ENV > Config > Default
func chooseValue(cfgValue interface{}, envValue interface{}, defaultValue interface{}) interface{} {
	val := defaultValue
	if cfgValue != nil {
		val = cfgValue
	}
	if envValue != nil {
		val = envValue
	}
	return val
}

// SetHook sets nirvana command hook.
func (s *command) SetHook(hook NirvanaCommandHook) {
	if hook == nil {
		hook = &NirvanaCommandHookFunc{}
	}
	s.hook = hook
}

// Hook returns nirvana command hook.
func (s *command) Hook() NirvanaCommandHook {
	return s.hook
}
