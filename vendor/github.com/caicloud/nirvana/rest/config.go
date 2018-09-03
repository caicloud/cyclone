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

package rest

import (
	"net/http"
)

// RequestExecutor implements a http client.
type RequestExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config is rest client config.
type Config struct {
	// Scheme is http scheme. It can be "http" or "https".
	Scheme string
	// Host must be a host string, a host:port or a URL to a server.
	Host string
	// Executor is used to execute http requests.
	// If it is empty, http.DefaultClient is used.
	Executor RequestExecutor
}

// DeepCopy returns a new config copied from the current one.
func (c *Config) DeepCopy() *Config {
	cfg := *c
	return &cfg
}

// Complete completes the config and returns an error if something is wrong.
func (c *Config) Complete() error {
	if c.Scheme == "" {
		c.Scheme = "http"
	}
	if c.Scheme != "http" && c.Scheme != "https" {
		return unrecognizedHTTPScheme.Error(c.Scheme)
	}
	if c.Host == "" {
		return noHTTPHost.Error()
	}
	if c.Executor == nil {
		c.Executor = &http.Client{}
	}
	return nil
}
