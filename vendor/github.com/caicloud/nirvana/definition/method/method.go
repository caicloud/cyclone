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

package method

import (
	"fmt"
	"reflect"
	"sync"
)

var defaultContainer = NewContainer()

// Put puts an instance in this container. The instance must have one or more methods.
func Put(instance interface{}) {
	defaultContainer.Put(instance)
}

// PutInterface puts an instance in this container. The instance must have one or more methods.
// The iface should be like (*ArbitraryInterface)(nil).
func PutInterface(iface interface{}, instance interface{}) {
	defaultContainer.PutInterface(iface, instance)
}

// Get returns a function for specified method. If you want to specify a method from an
// interface, you need to use (*ArbitraryInterface)(nil) as instance.
func Get(instance interface{}, method string) interface{} {
	return defaultContainer.Get(instance, method)
}

// Container contains instances and mappings.
type Container struct {
	instances map[reflect.Type]interface{}
	lock      sync.RWMutex
}

// NewContainer creates a method container.
func NewContainer() *Container {
	return &Container{
		instances: make(map[reflect.Type]interface{}),
	}
}

// Put puts an instance in this container. The instance must have one or more methods.
func (c *Container) Put(instance interface{}) {
	typ := reflect.TypeOf(instance)
	if typ.NumMethod() <= 0 {
		panic(fmt.Errorf("type %s has no method", typ.String()))
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.instances[typ] = instance
}

// PutInterface puts an instance in this container. The instance must have one or more methods.
// The iface should be like (*ArbitraryInterface)(nil).
func (c *Container) PutInterface(iface interface{}, instance interface{}) {
	typ := c.typeOf(iface)
	if typ.Kind() != reflect.Interface {
		panic(fmt.Errorf("type %s doesn't an interface", typ.String()))
	}
	if typ.NumMethod() <= 0 {
		panic(fmt.Errorf("type %s has no method", typ.String()))
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.instances[typ] = instance
}

// funcOf gets function value.
func (c *Container) funcOf(typ reflect.Type, method string) reflect.Value {
	c.lock.RLock()
	ins, ok := c.instances[typ]
	c.lock.RUnlock()
	if !ok {
		panic(fmt.Errorf("no instance for type %s", typ))
	}
	return reflect.ValueOf(ins).MethodByName(method)
}

// Get returns a function for specified method. If you want to specify a method from an
// interface, you need to use (*ArbitraryInterface)(nil) as instance.
func (c *Container) Get(instance interface{}, method string) interface{} {
	typ, funcTyp := c.typesOf(instance, method)
	var lock sync.Mutex
	var value *reflect.Value
	return reflect.MakeFunc(funcTyp, func(args []reflect.Value) (results []reflect.Value) {
		if value == nil {
			// Only get func value once.
			lock.Lock()
			if value == nil {
				v := c.funcOf(typ, method)
				value = &v
			}
			// If it paniced, this mutex won't unlock.
			// It doesn't matter.
			lock.Unlock()
		}
		return value.Call(args)
	}).Interface()
}

// typesOf returns instance type and func type.
func (c *Container) typesOf(instance interface{}, method string) (insType reflect.Type, funcType reflect.Type) {
	if method == "" {
		panic(fmt.Errorf("method must not be empty"))
	}
	insType = c.typeOf(instance)
	if insType.NumMethod() <= 0 {
		panic(fmt.Errorf("type %s has no method", insType.String()))
	}
	m, ok := insType.MethodByName(method)
	if !ok {
		panic(fmt.Errorf("no method named %s in type %s", method, insType.String()))
	}
	funcType = m.Type
	if insType.Kind() != reflect.Interface {
		ins := make([]reflect.Type, funcType.NumIn()-1)
		for i := 1; i < funcType.NumIn(); i++ {
			ins[i-1] = funcType.In(i)
		}
		outs := make([]reflect.Type, funcType.NumOut())
		for i := 0; i < funcType.NumOut(); i++ {
			outs[i] = funcType.Out(i)
		}
		funcType = reflect.FuncOf(ins, outs, funcType.IsVariadic())
	}
	return insType, funcType
}

// typeOf gets original type of instance.
func (c *Container) typeOf(instance interface{}) reflect.Type {
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Interface {
		// Replace *interface{} with interface{}
		typ = typ.Elem()
	}
	return typ
}
