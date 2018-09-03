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

// Container contains informations to generate APIs.
type Container struct {
	modifiers     service.DefinitionModifiers
	descriptors   []definition.Descriptor
	typeContainer *TypeContainer
	analyzer      *Analyzer
}

// NewContainer creates API container.
func NewContainer() *Container {
	return &Container{
		typeContainer: NewTypeContainer(),
		analyzer:      NewAnalyzer(),
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
