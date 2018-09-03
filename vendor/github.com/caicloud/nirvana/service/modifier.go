/*
Copyright 2017 Caicloud Authors

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

package service

import (
	"net/http"

	"github.com/caicloud/nirvana/definition"
)

// DefinitionModifier is used in Server. It's used to modify definition.
// If you want to add some common data into all definitions, you can write
// a customized modifier for it.
type DefinitionModifier func(d *definition.Definition)

// DefinitionModifiers is a convenient type for []DefinitionModifier
type DefinitionModifiers []DefinitionModifier

// Combine combines a list of modifiers to one.
func (m DefinitionModifiers) Combine() DefinitionModifier {
	return func(d *definition.Definition) {
		for _, f := range m {
			f(d)
		}
	}
}

// FirstContextParameter adds a context prefab parameter into all definitions.
// Then you don't need to manually write the parameter to every definitions.
func FirstContextParameter() DefinitionModifier {
	return func(d *definition.Definition) {
		if len(d.Parameters) > 0 {
			p := d.Parameters[0]
			if p.Source == definition.Prefab && p.Name == "context" {
				return
			}
		}
		ps := make([]definition.Parameter, len(d.Parameters)+1)
		ps[0] = definition.Parameter{
			Name:   "context",
			Source: definition.Prefab,
		}
		copy(ps[1:], d.Parameters)
		d.Parameters = ps
	}
}

// ConsumeAllIfConsumesIsEmpty adds definition.MIMEAll to consumes if consumes
// is empty.
func ConsumeAllIfConsumesIsEmpty() DefinitionModifier {
	return func(d *definition.Definition) {
		if len(d.Consumes) <= 0 {
			d.Consumes = []string{definition.MIMEAll}
		}
	}
}

// ProduceAllIfProducesIsEmpty adds definition.MIMEAll to consumes if consumes
// is empty.
func ProduceAllIfProducesIsEmpty() DefinitionModifier {
	return func(d *definition.Definition) {
		if len(d.Produces) <= 0 {
			d.Produces = []string{definition.MIMEAll}
		}
	}
}

// ConsumeNoneForHTTPGet adds definition.MIMENone to consumes for get definitions.
// Then you don't need to manually write the consume to every get definitions.
// The get is http get rather than definition.Get.
func ConsumeNoneForHTTPGet() DefinitionModifier {
	return func(d *definition.Definition) {
		if HTTPMethodFor(d.Method) == http.MethodGet {
			found := false
			for _, v := range d.Consumes {
				if v == definition.MIMENone {
					found = true
					break
				}
			}
			if !found {
				d.Consumes = append(d.Consumes, definition.MIMENone)
			}
		}
	}
}

// ConsumeNoneForHTTPDelete adds definition.MIMENone to consumes for delete definitions.
// Then you don't need to manually write the consume to every delete definitions.
// The delete is http delete rather than definition.Delete.
func ConsumeNoneForHTTPDelete() DefinitionModifier {
	return func(d *definition.Definition) {
		if HTTPMethodFor(d.Method) == http.MethodDelete {
			found := false
			for _, v := range d.Consumes {
				if v == definition.MIMENone {
					found = true
					break
				}
			}
			if !found {
				d.Consumes = append(d.Consumes, definition.MIMENone)
			}
		}
	}
}

// ProduceNoneForHTTPDelete adds definition.MIMENone to produces for delete definitions.
// Then you don't need to manually write the produce to every delete definitions.
// The delete is http delete rather than definition.Delete.
func ProduceNoneForHTTPDelete() DefinitionModifier {
	return func(d *definition.Definition) {
		if HTTPMethodFor(d.Method) == http.MethodDelete {
			found := false
			for _, v := range d.Produces {
				if v == definition.MIMENone {
					found = true
					break
				}
			}
			if !found {
				d.Produces = append(d.Produces, definition.MIMENone)
			}
		}
	}
}

// LastErrorResult adds a error result into all definitions.
// Then you don't need to manually write the result to every definitions.
func LastErrorResult() DefinitionModifier {
	return func(d *definition.Definition) {
		length := len(d.Results)
		if length > 0 {
			r := d.Results[length-1]
			if r.Destination == definition.Error {
				return
			}
		}
		d.Results = append(d.Results, definition.Result{
			Destination: definition.Error,
		})
	}
}
