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
	"context"
	"net/http"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service/router"
)

// Builder builds service.
type Builder interface {
	// Logger returns logger of builder.
	Logger() log.Logger
	// SetLogger sets logger to server.
	SetLogger(logger log.Logger)
	// Modifier returns modifier of builder.
	Modifier() DefinitionModifier
	// SetModifier sets definition modifier.
	SetModifier(m DefinitionModifier)
	// Filters returns all request filters.
	Filters() []Filter
	// AddFilters add filters to filter requests.
	AddFilter(filters ...Filter)
	// AddDescriptors adds descriptors to router.
	AddDescriptor(descriptors ...definition.Descriptor) error
	// Middlewares returns all router middlewares.
	Middlewares() map[string][]definition.Middleware
	// Definitions returns all definitions. If a modifier exists, it will be executed.
	Definitions() map[string][]definition.Definition
	// Build builds a service to handle request.
	Build() (Service, error)
}

type binding struct {
	middlewares []definition.Middleware
	definitions []definition.Definition
}

type builder struct {
	bindings map[string]*binding
	modifier DefinitionModifier
	filters  []Filter
	logger   log.Logger
}

// NewBuilder creates a service builder.
func NewBuilder() Builder {
	return &builder{
		bindings: make(map[string]*binding),
		logger:   &log.SilentLogger{},
	}
}

// Filters returns all request filters.
func (b *builder) Filters() []Filter {
	result := make([]Filter, len(b.filters))
	copy(result, b.filters)
	return result
}

// AddFilters add filters to filter requests.
func (b *builder) AddFilter(filters ...Filter) {
	b.filters = append(b.filters, filters...)
}

// Logger returns logger of builder.
func (b *builder) Logger() log.Logger {
	return b.logger
}

// SetLogger sets logger to builder.
func (b *builder) SetLogger(logger log.Logger) {
	if logger != nil {
		b.logger = logger
	} else {
		b.logger = &log.SilentLogger{}
	}
}

// Modifier returns modifier of builder.
func (b *builder) Modifier() DefinitionModifier {
	return b.modifier
}

// SetModifier sets definition modifier.
func (b *builder) SetModifier(m DefinitionModifier) {
	b.modifier = m
}

// AddDescriptor adds descriptors to router.
func (b *builder) AddDescriptor(descriptors ...definition.Descriptor) error {
	for _, descriptor := range descriptors {
		b.addDescriptor("", nil, nil, descriptor)
	}
	return nil
}

func (b *builder) addDescriptor(prefix string, consumes []string, produces []string, descriptor definition.Descriptor) {
	path := strings.Join([]string{prefix, strings.Trim(descriptor.Path, "/")}, "/")
	if descriptor.Consumes != nil {
		consumes = descriptor.Consumes
	}
	if descriptor.Produces != nil {
		produces = descriptor.Produces
	}
	if len(descriptor.Middlewares) > 0 || len(descriptor.Definitions) > 0 {
		bd, ok := b.bindings[path]
		if !ok {
			bd = &binding{}
			b.bindings[path] = bd
		}
		if len(descriptor.Middlewares) > 0 {
			bd.middlewares = append(bd.middlewares, descriptor.Middlewares...)
		}
		if len(descriptor.Definitions) > 0 {
			for _, d := range descriptor.Definitions {
				bd.definitions = append(bd.definitions, *b.copyDefinition(&d, consumes, produces))
			}
		}
	}
	for _, child := range descriptor.Children {
		b.addDescriptor(strings.TrimRight(path, "/"), consumes, produces, child)
	}
}

// copyDefinition creates a copy from original definition. Those fields with type interface{} only have shallow copies.
func (b *builder) copyDefinition(d *definition.Definition, consumes []string, produces []string) *definition.Definition {
	newOne := &definition.Definition{
		Method:      d.Method,
		Summary:     d.Summary,
		Function:    d.Function,
		Description: d.Description,
	}
	if len(d.Consumes) > 0 {
		consumes = d.Consumes
	}
	newOne.Consumes = make([]string, len(consumes))
	copy(newOne.Consumes, consumes)

	if len(d.Produces) > 0 {
		produces = d.Produces
	}
	newOne.Produces = make([]string, len(produces))
	copy(newOne.Produces, produces)

	if len(d.ErrorProduces) > 0 {
		produces = d.ErrorProduces
	}
	newOne.ErrorProduces = make([]string, len(produces))
	copy(newOne.ErrorProduces, produces)

	newOne.Parameters = make([]definition.Parameter, len(d.Parameters))
	for i, p := range d.Parameters {
		newParameter := p
		newParameter.Operators = make([]definition.Operator, len(p.Operators))
		copy(newParameter.Operators, p.Operators)
		newOne.Parameters[i] = newParameter
	}
	newOne.Results = make([]definition.Result, len(d.Results))
	for i, r := range d.Results {
		newResult := r
		newResult.Operators = make([]definition.Operator, len(r.Operators))
		copy(newResult.Operators, r.Operators)
		newOne.Results[i] = newResult
	}
	newOne.Examples = make([]definition.Example, len(d.Examples))
	copy(newOne.Examples, d.Examples)
	return newOne
}

// Middlewares returns all router middlewares.
func (b *builder) Middlewares() map[string][]definition.Middleware {
	result := make(map[string][]definition.Middleware)
	for path, bd := range b.bindings {
		if len(bd.middlewares) > 0 {
			middlewares := make([]definition.Middleware, len(bd.middlewares))
			copy(middlewares, bd.middlewares)
			result[path] = middlewares
		}
	}
	return result
}

// Definitions returns all definitions. If a modifier exists, it will be executed.
// All results are copied from original definitions. Modifications can not affect
// original data.
func (b *builder) Definitions() map[string][]definition.Definition {
	result := make(map[string][]definition.Definition)
	for path, bd := range b.bindings {
		if len(bd.definitions) > 0 {
			definitions := make([]definition.Definition, len(bd.definitions))
			for i, d := range bd.definitions {
				newCopy := b.copyDefinition(&d, nil, nil)
				if b.modifier != nil {
					b.modifier(newCopy)
				}
				definitions[i] = *newCopy
			}
			result[path] = definitions
		}
	}
	return result
}

// Build builds a service to handle request.
func (b *builder) Build() (Service, error) {
	if len(b.bindings) <= 0 {
		return nil, noRouter.Error()
	}
	var root router.Router
	for path, bd := range b.bindings {
		b.logger.V(log.LevelDebug).Infof("Definitions: %d Middlewares: %d Path: %s",
			len(bd.definitions), len(bd.middlewares), path)
		top, leaf, err := router.Parse(path)
		if err != nil {
			b.logger.Errorf("Can't parse path: %s, %s", path, err.Error())
			return nil, err
		}
		if len(bd.definitions) > 0 {
			// RedirectTrailingSlash would redirect "/somepath/" to "/somepath". Any definition under "/somepath/"
			// will never be executed.
			if len(path) > 1 && strings.HasSuffix(path, "/") {
				b.logger.Warningf("If RedirectTrailingSlash filter is enabled, following %d definition(s) would not be executed", len(bd.definitions))
			}
			inspector := newInspector(path, b.logger)
			for _, d := range bd.definitions {
				b.logger.V(log.LevelDebug).Infof("  Method: %s Consumes: %v Produces: %v",
					d.Method, d.Consumes, d.Produces)
				if b.modifier != nil {
					b.modifier(&d)
				}
				if err := inspector.addDefinition(d); err != nil {
					return nil, err
				}
			}

			leaf.SetInspector(inspector)
		}
		for _, m := range bd.middlewares {
			m := m
			leaf.AddMiddleware(func(ctx context.Context, chain router.RoutingChain) error {
				return m(ctx, chain)
			})
		}
		if root == nil {
			root = top
		} else if root, err = root.Merge(top); err != nil {
			return nil, err
		}
	}
	s := &service{
		root:      root,
		filters:   b.filters,
		logger:    b.logger,
		producers: AllProducers(),
	}
	return s, nil
}

// Service handles HTTP requests.
//
// Workflow:
//            Service.ServeHTTP()
//          ----------------------
//          ↓                    ↑
// |-----Filters------|          ↑
//          ↓                    ↑
// |---Router Match---|          ↑
//          ↓                    ↑
// |-------------Middlewares------------|
//          ↓                    ↑
// |-------------Executor---------------|
//          ↓                    ↑
// |-ParameterGenerators-|-DestinationHandlers-|
//          ↓                    ↑
// |------------User Function-----------|
type Service interface {
	http.Handler
}

type service struct {
	root      router.Router
	filters   []Filter
	logger    log.Logger
	producers []Producer
}

func (s *service) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for _, f := range s.filters {
		if !f(resp, req) {
			return
		}
	}
	ctx := newHTTPContext(resp, req)

	executor, err := s.root.Match(ctx, &ctx.container, req.URL.Path)
	if err != nil {
		if err := writeError(ctx, s.producers, err); err != nil {
			s.logger.Error(err)
		}
		return
	}
	err = executor.Execute(ctx)
	if err == nil && ctx.response.HeaderWritable() {
		err = invalidService.Error()
	}
	if err != nil {
		if ctx.response.HeaderWritable() {
			if err := writeError(ctx, s.producers, err); err != nil {
				s.logger.Error(err)
			}
		} else {
			s.logger.Error(err)
		}
	}
}
