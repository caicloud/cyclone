/*
Copyright 2016 caicloud authors. All rights reserved.

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

package register

import (
	"fmt"
	"os"
	"sync"
)

var (
	defaultConfig = &Config{
		OverrideAllowed: false,
	}
)

// Register is a struct binding name and interface such as Constructor
type Register struct {
	data            map[string]interface{}
	mu              sync.RWMutex
	overrideAllowed bool
}

// Config is a struct containing all config for register
type Config struct {

	// OverrideAllowed allows the register to override
	// an already registered interface by name if it is true,
	// otherwise register will panic.
	OverrideAllowed bool
}

// NewRegister returns a new register
func NewRegister(config *Config) *Register {
	if config == nil {
		config = defaultConfig
	}

	return &Register{
		data:            make(map[string]interface{}),
		overrideAllowed: config.OverrideAllowed,
	}
}

// Register registers a interface by name.
// It will panic if name corresponds to an already registered interface
// and the register does not allow user to override the interface.
func (r *Register) Register(name string, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[name]
	if ok {
		if !r.overrideAllowed {
			panic("[Register] Repeated registration key: " + name)
		} else {
			fmt.Fprintln(os.Stderr, "[Register] Repeated registration key: "+name)
		}
	}
	r.data[name] = v

}

// Lock locks the register. Generally be used before you iterating the register
func (r *Register) Lock() {
	r.mu.Lock()
}

// Unlock unlocks the register. Generally be used after you iterating the register
func (r *Register) Unlock() {
	r.mu.Unlock()
}

// Get returns an interface registered with the given name
func (r *Register) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[name]
	return v, ok
}

// Contains returns true if name corresponds to an already registered interface
func (r *Register) Contains(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.data[name]
	return ok
}

// Iter returns an iterable map for <name, interface> pair
func (r *Register) Iter() map[string]interface{} {
	return r.data
}

// Keys returns the name of all registered interfaces
func (r *Register) Keys() []string {
	names := []string{}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for name := range r.data {
		names = append(names, name)
	}
	return names
}

// Values returns all registered interfaces
func (r *Register) Values() []interface{} {
	ret := []interface{}{}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.data {
		ret = append(ret, v)
	}
	return ret
}

// Clear cleans up the registered items
func (r *Register) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = make(map[string]interface{})
}
