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

package logger

import (
	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/log"
)

// ExternalConfigName is the external config name of logger.
const ExternalConfigName = "logger"

// Level Configurer set a new StdLogger with specified log level.
func Level(level log.Level) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Configure(nirvana.Logger(log.NewStdLogger(level)))
		return nil
	}
}

// Option contains basic configurations of logger.
type Option struct {
	// Debug is logger level.
	Debug bool `desc:"Debug mode. Output all logs"`
	// Level is logger level.
	Level int32 `desc:"Log level. This field is no sense if debug is enabled"`
	// OverrideGlobal modifies nirvana global logger.
	OverrideGlobal bool `desc:"Override global logger"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Debug:          false,
		Level:          0,
		OverrideGlobal: false,
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	if p.Debug {
		p.Level = int32(log.LevelDebug)
	}
	l := log.NewStdLogger(log.Level(p.Level))
	if p.OverrideGlobal {
		log.SetDefaultLogger(l)
	}
	cfg.Configure(
		nirvana.Logger(l),
	)
	return nil
}
