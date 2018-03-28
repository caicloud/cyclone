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

import "sync"

type Register interface {
	Register(string, interface{})
	Get(string) interface{}
}

// register is a struct binds name and interface such as Constructor
type register struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewRegister returns a new register
func NewRegister() Register {
	return &register{
		data: make(map[string]interface{}),
	}
}

// Register binds name and interface
// It will panic if name already exists
func (r *register) Register(name string, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[name]
	if ok {
		panic("Repeated registration key: " + name)
	}
	r.data[name] = v

}

// Get returns an interface registered with the given name
func (r *register) Get(name string) interface{} {
	// need lock ?
	return r.data[name]
}
