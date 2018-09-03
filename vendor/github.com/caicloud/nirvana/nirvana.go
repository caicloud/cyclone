/*
Copyright 2017 Caicloud Authors

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

package nirvana

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

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

// Config describes configuration of server.
type Config struct {
	// tls cert file
	certFile string
	// tls ket file
	keyFile string
	// ip is the ip to listen. Empty means `0.0.0.0`.
	ip string
	// port is the port to listen.
	port uint16
	// logger is used to output info inside framework.
	logger log.Logger
	// descriptors contains all APIs.
	descriptors []definition.Descriptor
	// filters is http filters.
	filters []service.Filter
	// modifiers is definition modifiers
	modifiers service.DefinitionModifiers
	// configSet contains all configurations of plugins.
	configSet map[string]interface{}
	// locked is for locking current config. If the field
	// is not 0, any modification causes panic.
	locked int32
}

// lock locks config. If succeed, it will return ture.
func (c *Config) lock() bool {
	return atomic.CompareAndSwapInt32(&c.locked, 0, 1)
}

// Locked checks if the config is locked.
func (c *Config) Locked() bool {
	return atomic.LoadInt32(&c.locked) != 0
}

// IP returns listenning ip.
func (c *Config) IP() string {
	return c.ip
}

// Port returns listenning port.
func (c *Config) Port() uint16 {
	return c.port
}

// Logger returns logger.
func (c *Config) Logger() log.Logger {
	return c.logger
}

// Configurer is used to configure server config.
type Configurer func(c *Config) error

var immutableConfig = errors.InternalServerError.Build("Nirvana:ImmutableConfig", "config has been locked and must not be modified")

// Configure configs by configurers. It panics if an error occurs or config is locked.
func (c *Config) Configure(configurers ...Configurer) *Config {
	if c.Locked() {
		panic(immutableConfig.Error())
	}
	for _, configurer := range configurers {
		if err := configurer(c); err != nil {
			panic(err)
		}
	}
	return c
}

// Config gets external config by name. This method is for plugins.
func (c *Config) Config(name string) interface{} {
	return c.configSet[name]
}

// Set sets external config by name. This method is for plugins.
// Set a nil config will delete it.
func (c *Config) Set(name string, config interface{}) {
	if config == nil {
		delete(c.configSet, name)
	} else {
		c.configSet[name] = config
	}
}

// forEach traverse all plugin configs.
func (c *Config) forEach(f func(name string, config interface{}) error) error {
	for name, cfg := range c.configSet {
		if err := f(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// NewDefaultConfig creates default config.
// Default config contains:
//  Filters: RedirectTrailingSlash, FillLeadingSlash, ParseRequestForm.
//  Modifiers: FirstContextParameter,
//             ConsumeAllIfConsumesIsEmpty, ProduceAllIfProducesIsEmpty,
//             ConsumeNoneForHTTPGet, ConsumeNoneForHTTPDelete,
//             ProduceNoneForHTTPDelete.
func NewDefaultConfig() *Config {
	return NewConfig().Configure(
		Logger(log.DefaultLogger()),
		Filter(
			service.RedirectTrailingSlash(),
			service.FillLeadingSlash(),
			service.ParseRequestForm(),
		),
		Modifier(
			service.FirstContextParameter(),
			service.ConsumeAllIfConsumesIsEmpty(),
			service.ProduceAllIfProducesIsEmpty(),
			service.ConsumeNoneForHTTPGet(),
			service.ConsumeNoneForHTTPDelete(),
			service.ProduceNoneForHTTPDelete(),
		),
	)
}

// NewConfig creates a pure config. If you don't know how to set filters and
// modifiers for specific scenario, please use NewDefaultConfig().
func NewConfig() *Config {
	return &Config{
		ip:          "",
		port:        8080,
		logger:      &log.SilentLogger{},
		filters:     []service.Filter{},
		descriptors: []definition.Descriptor{},
		modifiers:   []service.DefinitionModifier{},
		configSet:   make(map[string]interface{}),
	}
}

// server is nirvana server.
type server struct {
	lock    sync.Mutex
	config  *Config
	server  *http.Server
	builder service.Builder
	cleaner func() error
}

// NewServer creates a nirvana server. After creation, don't modify
// config. Also don't create another server with current config.
func NewServer(c *Config) Server {
	if c.Locked() || !c.lock() {
		panic(immutableConfig.Error())
	}
	return &server{
		config: c,
	}
}

var noConfigInstaller = errors.InternalServerError.Build("Nirvana:NoConfigInstaller", "no config installer for external config name ${name}")

// Builder create a service builder for current server. Don't use this method directly except
// there is a special server to hold http services. After server shutdown, clean resources via
// returned cleaner.
// This method always returns same builder until cleaner is called. Then it will
// returns new one.
func (s *server) Builder() (builder service.Builder, cleaner func() error, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.builder != nil {
		return s.builder, s.cleaner, nil
	}
	builder = service.NewBuilder()
	builder.SetLogger(s.config.logger)
	builder.AddFilter(s.config.filters...)
	builder.SetModifier(s.config.modifiers.Combine())
	if err := builder.AddDescriptor(s.config.descriptors...); err != nil {
		return nil, nil, err
	}
	if err := s.config.forEach(func(name string, config interface{}) error {
		installer := ConfigInstallerFor(name)
		if installer == nil {
			return noConfigInstaller.Error(name)
		}
		return installer.Install(builder, s.config)
	}); err != nil {
		return nil, nil, err
	}
	s.builder = builder
	s.cleaner = func() (err error) {
		// Clean builder and plugins.
		s.lock.Lock()
		defer func() {
			if err == nil {
				s.builder = nil
				s.cleaner = nil
			}
			s.lock.Unlock()
		}()
		if s.builder == nil {
			return nil
		}
		return s.config.forEach(func(name string, config interface{}) error {
			installer := ConfigInstallerFor(name)
			if installer == nil {
				return noConfigInstaller.Error(name)
			}
			return installer.Uninstall(builder, s.config)
		})
	}
	return s.builder, s.cleaner, nil
}

var builderInUse = errors.InternalServerError.Build("Nirvana:BuilderInUse", "service builder is in use, clean it before serving")

// Serve starts to listen and serve requests.
// The method won't return except an error occurs.
func (s *server) Serve() (e error) {
	s.lock.Lock()
	if s.builder != nil || s.cleaner != nil {
		return builderInUse.Error()
	}
	s.lock.Unlock()
	builder, cleaner, err := s.Builder()
	if err != nil {
		return err
	}
	defer func() {
		if err := cleaner(); err != nil {
			s.config.logger.Error(err)
			if e == nil {
				e = err
			}
		}
	}()

	service, err := builder.Build()
	if err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.ip, s.config.port),
		Handler: service,
	}

	if len(s.config.certFile) != 0 && len(s.config.keyFile) != 0 {
		return s.server.ListenAndServeTLS(s.config.certFile, s.config.keyFile)
	}
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.
func (s *server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ConfigInstaller is used to install config to service builder.
type ConfigInstaller interface {
	// Name is the external config name.
	Name() string
	// Install installs stuffs before server starting.
	Install(builder service.Builder, config *Config) error
	// Uninstall uninstalls stuffs after server terminating.
	Uninstall(builder service.Builder, config *Config) error
}

var installers = map[string]ConfigInstaller{}

// ConfigInstallerFor gets installer by name.
func ConfigInstallerFor(name string) ConfigInstaller {
	return installers[name]
}

// RegisterConfigInstaller registers a config installer.
func RegisterConfigInstaller(ci ConfigInstaller) {
	if ConfigInstallerFor(ci.Name()) != nil {
		panic(fmt.Sprintf("Config installer %s has been installed.", ci.Name()))
	}
	installers[ci.Name()] = ci
}

// IP returns a configurer to set ip into config.
func IP(ip string) Configurer {
	return func(c *Config) error {
		c.ip = ip
		return nil
	}
}

// TLS returns a configurer to set certFile and keyFile in config.
func TLS(certFile, keyFile string) Configurer {
	return func(c *Config) error {
		c.certFile = certFile
		c.keyFile = keyFile
		return nil
	}
}

// Port returns a configurer to set port into config.
func Port(port uint16) Configurer {
	return func(c *Config) error {
		c.port = port
		return nil
	}
}

// Logger returns a configurer to set logger into config.
func Logger(logger log.Logger) Configurer {
	return func(c *Config) error {
		if logger == nil {
			c.logger = &log.SilentLogger{}
		} else {
			c.logger = logger
		}
		return nil
	}
}

// Descriptor returns a configurer to add descriptors into config.
func Descriptor(descriptors ...definition.Descriptor) Configurer {
	return func(c *Config) error {
		c.descriptors = append(c.descriptors, descriptors...)
		return nil
	}
}

// Filter returns a configurer to add filters into config.
func Filter(filters ...service.Filter) Configurer {
	return func(c *Config) error {
		c.filters = append(c.filters, filters...)
		return nil
	}
}

// Modifier returns a configurer to add definition modifiers into config.
func Modifier(modifiers ...service.DefinitionModifier) Configurer {
	return func(c *Config) error {
		c.modifiers = append(c.modifiers, modifiers...)
		return nil
	}
}
