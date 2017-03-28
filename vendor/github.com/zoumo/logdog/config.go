// Copyright 2016 Jim Zhang (jim.zoumo@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logdog

import (
	"encoding/json"
	"fmt"
)

// ConfigLoader is an interface which con load map[string]interface{} config
type ConfigLoader interface {
	LoadConfig(map[string]interface{}) error
}

// Config is a alias of map[string]interface{}
type Config map[string]interface{}

// LogConfig defines the configuration of logger
type LogConfig struct {
	DisableExistingLoggers bool                              `json:"disableExistingLoggers"`
	Formatters             map[string]map[string]interface{} `json:"formatters"`
	Handlers               map[string]map[string]interface{} `json:"handlers"`
	Loggers                map[string]map[string]interface{} `json:"loggers"`
}

// LoadJSONConfig loads a json config
// if DisableExistingLoggers is true, all existing loggers will be
// closed, then a new root logger will be created
func LoadJSONConfig(config []byte) error {
	var logConfig LogConfig

	if err := json.Unmarshal(config, &logConfig); err != nil {
		return err
	}

	if logConfig.DisableExistingLoggers {
		DisableExistingLoggers()
	}

	builder := func(name string, conf map[string]interface{}) (ConfigLoader, error) {
		c, ok := conf["class"]
		if !ok {
			return nil, fmt.Errorf("'class' filed is required when building %s", name)
		}
		classname := c.(string)
		class := GetConstructor(classname)
		if class == nil {
			return nil, fmt.Errorf("can not find constructor: %s", classname)
		}

		// if name is not set, use outside name
		if _, ok := conf["name"]; !ok {
			conf["name"] = name
		}

		b := class()
		if err := b.LoadConfig(conf); err != nil {
			return nil, err
		}
		return b, nil
	}

	if logConfig.Formatters != nil {
		for name, conf := range logConfig.Formatters {
			temp, err := builder(name, conf)
			if err != nil {
				return err
			}
			formatter := temp.(Formatter)
			RegisterFormatter(name, formatter)
		}
	}

	if logConfig.Handlers != nil {
		for name, conf := range logConfig.Handlers {
			temp, err := builder(name, conf)
			if err != nil {
				return err
			}
			handler := temp.(Handler)
			RegisterHandler(name, handler)
		}
	}

	if logConfig.Loggers != nil {
		for name, conf := range logConfig.Loggers {
			logger := GetLogger(name)

			// if name is not set, use outside name
			if _, ok := conf["name"]; !ok {
				conf["name"] = name
			}
			if err := logger.LoadConfig(conf); err != nil {
				return err
			}

		}
	}

	return nil
}
