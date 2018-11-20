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

package service

import (
	"context"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"
)

type definitionMap struct {
	def definition.Definition
	err errors.Factory
}

func TestAddDefinition(t *testing.T) {
	inspector := newInspector("/test", &log.SilentLogger{})
	units := []definitionMap{
		{
			definition.Definition{
				Method: definition.Method(""),
			},
			definitionNoMethod,
		},
		{
			definition.Definition{
				Method: definition.Get,
			},
			definitionNoConsumes,
		},
		{
			definition.Definition{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
			},
			definitionNoProduces,
		},
		{
			definition.Definition{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
				Produces: []string{definition.MIMEJSON},
			},
			definitionNoErrorProduces,
		},

		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
			},
			definitionNoFunction,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function:      1,
			},
			definitionInvalidFunctionType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{"invalid-content-type"},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			definitionNoConsumer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{"invalid-content-type"},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			definitionNoProducer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{"invalid-content-type"},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			definitionNoProducer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func(ctx context.Context) {
				},
			},
			definitionUnmatchedParameters,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: "InvalidSource",
					},
				},
				Function: func(a int) {
				},
			},
			noParameterGenerator,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source:  definition.Path,
						Name:    "a",
						Default: "InvalidDefaultValue",
					},
				},
				Function: func(a int) {
				},
			},
			unassignableType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   "a",
					},
				},
				Function: func(a []*int) {
				},
			},
			noConverter,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Query,
						Name:   "a",
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
						},
					},
				},
				Function: func(a []int) {
				},
			},
			invalidOperatorOutType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Query,
						Name:   "a",
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) ([]int, error) {
								return []int{1}, nil
							}),
						},
					},
				},
				Function: func(a []int) {
				},
			},
			invalidOperatorInType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() int {
					return 0
				},
			},
			definitionUnmatchedResults,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Results: []definition.Result{
					{
						Destination: definition.Destination("InvalidDestination"),
					},
				},
				Function: func() int {
					return 0
				},
			},
			noDestinationHandler,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
						},
					},
				},
				Function: func() int {
					return 0
				},
			},
			invalidOperatorInType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			nil,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			definitionConflict,
		},
	}

	for _, unit := range units {
		err := inspector.addDefinition(unit.def)
		if unit.err != nil {
			if err == nil {
				t.Errorf("Expected error but got nil for %+v", unit.def)
			} else if !unit.err.Derived(err) {
				t.Fatalf("Unexpected err: %v for %+v", err, unit.def)
			}
		} else if err != nil {
			t.Fatalf("Unexpected err: %v for %+v", err, unit.def)
		}
	}
}
