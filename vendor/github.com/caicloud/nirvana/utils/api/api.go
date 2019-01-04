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

package api

import (
	"fmt"
	"reflect"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

// Definitions describes all APIs and its related object types.
type Definitions struct {
	// Definitions holds mappings between path and API descriptions.
	Definitions map[string][]Definition
	// Types contains all types used by definitions.
	Types map[TypeName]*Type
}

// Subset returns a subset required by a definition filter.
func (d *Definitions) Subset(require func(path string, def *Definition) bool) *Definitions {
	definitions := map[string][]Definition{}
	for path, defs := range d.Definitions {
		for _, def := range defs {
			if require(path, &def) {
				definitions[path] = append(definitions[path], def)
			}
		}
	}
	if len(definitions) > 0 {
		result := &Definitions{definitions, map[TypeName]*Type{}}
		d.complete(result)
		return result
	}
	return nil
}

// complete fills types for a new definitions. target definitions must be a subset of this definitions.
func (d *Definitions) complete(definitions *Definitions) {
	for _, defs := range definitions.Definitions {
		for _, def := range defs {
			d.fillTypes(definitions.Types, def.Function)
			for _, parameter := range def.Parameters {
				d.fillTypes(definitions.Types, parameter.Type)
			}
			for _, result := range def.Results {
				d.fillTypes(definitions.Types, result.Type)
			}
		}
	}
}

// fillTypes puts a type into a map. The type must be in this definitions.
func (d *Definitions) fillTypes(types map[TypeName]*Type, name TypeName) {
	if types[name] != nil {
		return
	}
	typ, ok := d.Types[name]
	if !ok {
		panic(fmt.Errorf("no type named %s", name))
	}
	types[name] = typ
	switch typ.Kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Interface:
		// For basic data types, there are no related types need to handle.
		// For interfaces, there are no concrete entities to handle.
	case reflect.Array, reflect.Slice, reflect.Ptr:
		d.fillTypes(types, typ.Elem)
	case reflect.Map:
		d.fillTypes(types, typ.Key)
		d.fillTypes(types, typ.Elem)
	case reflect.Struct:
		for _, field := range typ.Fields {
			d.fillTypes(types, field.Type)
		}
	case reflect.Func:
		for _, param := range typ.In {
			d.fillTypes(types, param.Type)
		}
		for _, result := range typ.Out {
			d.fillTypes(types, result.Type)
		}
	default:
		panic(fmt.Errorf("can't recognize type %s with kind %s", name, typ.Kind.String()))
	}
}

// Container contains informations to generate APIs.
type Container struct {
	modifiers     service.DefinitionModifiers
	descriptors   []definition.Descriptor
	typeContainer *TypeContainer
	analyzer      *Analyzer
}

// NewContainer creates API container.
func NewContainer(root string) *Container {
	return &Container{
		typeContainer: NewTypeContainer(),
		analyzer:      NewAnalyzer(root),
	}
}

// AddModifier add definition modifiers to container.
func (ac *Container) AddModifier(modifiers ...service.DefinitionModifier) {
	ac.modifiers = append(ac.modifiers, modifiers...)
}

// AddDescriptor add descriptors to container.
func (ac *Container) AddDescriptor(descriptors ...definition.Descriptor) {
	ac.descriptors = append(ac.descriptors, descriptors...)
}

// Generate generates API definitions.
func (ac *Container) Generate() (*Definitions, error) {
	builder := service.NewBuilder()
	builder.SetModifier(ac.modifiers.Combine())
	if err := builder.AddDescriptor(ac.descriptors...); err != nil {
		return nil, err
	}
	definitions := builder.Definitions()
	result, err := NewPathDefinitions(ac.typeContainer, definitions)
	if err != nil {
		return nil, err
	}
	err = ac.typeContainer.Complete(ac.analyzer)
	return &Definitions{
		Definitions: result,
		Types:       ac.typeContainer.Types(),
	}, err
}
