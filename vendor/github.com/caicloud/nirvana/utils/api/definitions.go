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
	"encoding/json"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

// Parameter describes a function parameter.
type Parameter struct {
	// Source is the parameter value generated from.
	Source definition.Source
	// Name is the name to get value from a request.
	Name string
	// Description describes the parameter.
	Description string
	// Type is parameter object type.
	Type TypeName
	// Default is encoded default value.
	Default []byte
}

// Result describes a function result.
type Result struct {
	// Destination is the target for the result.
	Destination definition.Destination
	// Description describes the result.
	Description string
	// Type is result object type.
	Type TypeName
}

// Example is just an example.
type Example struct {
	// Description describes the example.
	Description string
	// Type is result object type.
	Type TypeName
	// Instance is encoded instance data.
	Instance []byte
}

// Definition is complete version of def.Definition.
type Definition struct {
	// Method is definition method.
	Method definition.Method
	// HTTPMethod is http method.
	HTTPMethod string
	// HTTPCode is http success code.
	HTTPCode int
	// Summary is a brief of this definition.
	Summary string
	// Description describes the API handler.
	Description string
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// ErrorProduces is used to generate data for error. If this field is empty,
	// it means that this field equals to Produces.
	// In some cases, successful data and error data should be generated in
	// different ways.
	ErrorProduces []string
	// Function is a function handler. It must be func type.
	Function TypeName
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Examples contains many examples for the API handler.
	Examples []Example
}

// NewDefinition creates openapi.Definition from definition.Definition.
func NewDefinition(tc *TypeContainer, d *definition.Definition) (*Definition, error) {
	cd := &Definition{
		Method:        d.Method,
		HTTPMethod:    service.HTTPMethodFor(d.Method),
		HTTPCode:      service.HTTPCodeFor(d.Method),
		Summary:       d.Summary,
		Description:   d.Description,
		Consumes:      d.Consumes,
		Produces:      d.Produces,
		ErrorProduces: d.ErrorProduces,
		Function:      tc.NameOfInstance(d.Function),
	}
	functionType := tc.Type(cd.Function)
	for i, p := range d.Parameters {
		param := Parameter{
			Source:      p.Source,
			Name:        p.Name,
			Description: p.Description,
			Type:        functionType.In[i].Type,
		}
		if p.Default != nil {
			data, err := encode(p.Default)
			if err != nil {
				return nil, err
			}
			param.Default = data
		}
		if len(p.Operators) > 0 {
			param.Type = tc.NameOf(p.Operators[0].In())
		}
		cd.Parameters = append(cd.Parameters, param)
	}
	for i, r := range d.Results {
		result := Result{
			Destination: r.Destination,
			Description: r.Description,
			Type:        functionType.Out[i].Type,
		}
		if len(r.Operators) > 0 {
			result.Type = tc.NameOf(r.Operators[len(r.Operators)-1].Out())
		}
		cd.Results = append(cd.Results, result)
	}
	for _, e := range d.Examples {
		example := Example{
			Description: e.Description,
		}
		if e.Instance != nil {
			example.Type = tc.NameOfInstance(e.Instance)
			data, err := encode(e.Instance)
			if err != nil {
				return nil, err
			}
			example.Instance = data
		}
		cd.Examples = append(cd.Examples, example)
	}
	return cd, nil
}

// NewDefinitions creates a list of definitions.
func NewDefinitions(tc *TypeContainer, definitions []definition.Definition) ([]Definition, error) {
	result := make([]Definition, len(definitions))
	for i, d := range definitions {
		cd, err := NewDefinition(tc, &d)
		if err != nil {
			return nil, err
		}
		result[i] = *cd
	}
	return result, nil
}

// NewPathDefinitions creates a list of definitions with path.
func NewPathDefinitions(tc *TypeContainer, definitions map[string][]definition.Definition) (map[string][]Definition, error) {
	result := make(map[string][]Definition)
	for path, defs := range definitions {
		cds, err := NewDefinitions(tc, defs)
		if err != nil {
			return nil, err
		}
		result[path] = cds

	}
	return result, nil
}

// encode encodes instance to json format.
func encode(ins interface{}) ([]byte, error) {
	return json.Marshal(ins)
}
