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
	"github.com/caicloud/nirvana"
)

// NirvanaCommandHook provides several hook points for NirvanaCommand.
type NirvanaCommandHook interface {
	// PreConfigure runs before installing plugins.
	PreConfigure(config *nirvana.Config) error
	// PostConfigure runs after installing plugins and before creating nirvana server.
	PostConfigure(config *nirvana.Config) error
	// PreServe runs before nirvana server serving.
	PreServe(config *nirvana.Config, server nirvana.Server) error
	// PostServe runs after nirvana server shutting down or any error occurring.
	PostServe(config *nirvana.Config, server nirvana.Server, err error) error
}

// NirvanaCommandHookFunc is a helper to generate NirvanaCommandHook. Hook points
// are optional.
type NirvanaCommandHookFunc struct {
	PreConfigureFunc  func(config *nirvana.Config) error
	PostConfigureFunc func(config *nirvana.Config) error
	PreServeFunc      func(config *nirvana.Config, server nirvana.Server) error
	PostServeFunc     func(config *nirvana.Config, server nirvana.Server, err error) error
}

// PreConfigure runs before installing plugins.
func (h *NirvanaCommandHookFunc) PreConfigure(config *nirvana.Config) error {
	if h.PreConfigureFunc != nil {
		return h.PreConfigureFunc(config)
	}
	return nil
}

// PostConfigure runs after installing plugins and before creating nirvana server.
func (h *NirvanaCommandHookFunc) PostConfigure(config *nirvana.Config) error {
	if h.PostConfigureFunc != nil {
		return h.PostConfigureFunc(config)
	}
	return nil
}

// PreServe runs before nirvana server serving.
func (h *NirvanaCommandHookFunc) PreServe(config *nirvana.Config, server nirvana.Server) error {
	if h.PreServeFunc != nil {
		return h.PreServeFunc(config, server)
	}
	return nil
}

// PostServe runs after nirvana server shutting down or any error occurring.
func (h *NirvanaCommandHookFunc) PostServe(config *nirvana.Config, server nirvana.Server, err error) error {
	if h.PostServeFunc != nil {
		return h.PostServeFunc(config, server, err)
	}
	return err
}
