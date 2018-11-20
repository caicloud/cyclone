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

package router

import (
	"context"
	"reflect"
	"sort"
)

// handler contains middlewares and executor.
type handler struct {
	middlewares []Middleware
	inspector   Inspector
}

// AddMiddleware adds middleware to the router node.
// If the router matches a path, all middlewares in the router
// will be executed by the returned executor.
func (h *handler) AddMiddleware(ms ...Middleware) {
	h.middlewares = append(h.middlewares, ms...)
}

// Middlewares returns all middlewares of the router.
// Don't modify the returned values.
func (h *handler) Middlewares() []Middleware {
	return h.middlewares
}

// SetInspector sets inspector to the router node.
func (h *handler) SetInspector(inspector Inspector) {
	h.inspector = inspector
}

// Inspector gets inspector from the router node.
// Don't modify the returned values.
func (h *handler) Inspector() Inspector {
	return h.inspector
}

// Merge merges middlewares and executors.
func (h *handler) Merge(o *handler) error {
	h.AddMiddleware(o.middlewares...)
	if h.inspector != nil {
		if o.inspector != nil {
			return ConflictInspectors.Error()
		}
	} else {
		h.inspector = o.inspector
	}
	return nil
}

// pack packs middlewares with the executor.
func (h *handler) pack(e Executor) (Executor, error) {
	if e == nil {
		return nil, NoExecutor.Error()
	}
	if len(h.middlewares) <= 0 {
		return e, nil
	}
	return newMiddlewareExecutor(h.middlewares, e), nil
}

// unionExecutor packs middlewares and own executor.
func (h *handler) unionExecutor(ctx context.Context) (Executor, error) {
	if h.inspector == nil {
		return nil, NoInspector.Error()
	}
	e, err := h.inspector.Inspect(ctx)
	if err != nil {
		return nil, err
	}
	return h.pack(e)
}

// charRouter is a router for characters
type charRouter struct {
	char   byte
	router *stringNode
}

// children contains all children routers.
type children struct {
	stringRouters []charRouter
	regexpRouters []Router
	pathRouter    Router
}

// findStringRouter find a router by first char.
func (p *children) findStringRouter(char byte) Router {
	length := len(p.stringRouters)
	if length <= 3 {
		// If the length is less than 3, use linear search.
		for _, cr := range p.stringRouters {
			if cr.char == char {
				return cr.router
			}
		}
		return nil
	}
	// Binary search.
	index := sort.Search(len(p.stringRouters), func(i int) bool {
		return char <= p.stringRouters[i].char
	})
	if index >= length {
		return nil
	}
	target := p.stringRouters[index]
	if char != target.char {
		return nil
	}
	return target.router
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (p *children) Match(ctx context.Context, c Container, path string) (Executor, error) {
	if len(path) <= 0 {
		return nil, RouterNotFound.Error()
	}

	// Two routers may match same path:
	//   /path/{id} without inspector
	//   /path/{name} with inspector
	// When match `/path/some`, the first router won't match it and
	// returns NoInspector. The the second router can match the path.
	// If the second router can't generate an executor, an error is
	// returned by inspector. In this case, resultError should be the
	// assigned with the error from second router.
	// If there are multiple routers match a path, the error is from
	// the last matched router.
	resultError := RouterNotFound.Error()

	// Match string routers
	if len(p.stringRouters) > 0 {
		if router := p.findStringRouter(path[0]); router != nil {
			if executor, err := router.Match(ctx, c, path); err == nil {
				return executor, nil
			} else if !RouterNotFound.Derived(err) &&
				!NoInspector.Derived(err) &&
				!NoExecutor.Derived(err) {
				resultError = err
			}
		}
	}

	// Match regexp routers
	for _, regexp := range p.regexpRouters {
		if executor, err := regexp.Match(ctx, c, path); err == nil {
			return executor, nil
		} else if !RouterNotFound.Derived(err) &&
			!NoInspector.Derived(err) &&
			!NoExecutor.Derived(err) {
			resultError = err
		}
	}

	// Match path router
	if p.pathRouter != nil {
		if executor, err := p.pathRouter.Match(ctx, c, path); err == nil {
			return executor, nil
		} else if !RouterNotFound.Derived(err) &&
			!NoInspector.Derived(err) &&
			!NoExecutor.Derived(err) {
			resultError = err
		}
	}
	return nil, resultError
}

// addRouter adds a router to current progeny.
func (p *children) addRouter(router Router) error {
	switch router.Kind() {
	case String:
		target := router.Target()
		if len(target) <= 0 {
			return EmptyRouterTarget.Error(router.Kind())
		}
		r, ok := router.(*stringNode)
		if !ok {
			return UnknownRouterType.Error(router.Kind(), reflect.TypeOf(router).String())
		}
		c := target[0]
		sr := p.findStringRouter(c)
		if sr != nil {
			_, err := sr.Merge(router)
			return err
		}
		length := len(p.stringRouters)
		index := 0
		if length > 0 {
			index = sort.Search(length, func(i int) bool {
				return c < p.stringRouters[i].char
			})
		}
		cr := charRouter{c, r}
		if index >= length {
			p.stringRouters = append(p.stringRouters, cr)
		} else {
			p.stringRouters = append(p.stringRouters[:index+1], p.stringRouters[index:]...)
			p.stringRouters[index] = cr
		}
	case Regexp:
		found := false
		for _, r := range p.regexpRouters {
			if r.Target() == router.Target() {
				if _, err := r.Merge(router); err != nil {
					return err
				}
				found = true
				break
			}
		}
		if !found {
			p.regexpRouters = append(p.regexpRouters, router)
		}
	case Path:
		if p.pathRouter != nil {
			r, err := p.pathRouter.Merge(router)
			if err != nil {
				return err
			}
			p.pathRouter = r
		} else {
			p.pathRouter = router
		}
	default:
		return UnknownRouterType.Error(router.Kind(), reflect.TypeOf(router).String())
	}
	return nil
}

// merge merges children routers.
func (p *children) merge(o *children) error {
	for _, r := range o.stringRouters {
		if err := p.addRouter(r.router); err != nil {
			return err
		}
	}
	for _, r := range o.regexpRouters {
		if err := p.addRouter(r); err != nil {
			return err
		}
	}
	if o.pathRouter != nil {
		if err := p.addRouter(o.pathRouter); err != nil {
			return err
		}
	}
	return nil
}
