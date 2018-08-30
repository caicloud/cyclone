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

import "context"

// middlewareExecutor is a combination of middlewares and executor.
type middlewareExecutor struct {
	// Middlewares contains all middlewares for the executor.
	middlewares []Middleware
	// Index is used to record the count of executed middleware.
	index int
	// Executor executes after middlewares.
	executor Executor
}

// newMiddlewareExecutor creates a new executor with middlewares.
func newMiddlewareExecutor(ms []Middleware, e Executor) Executor {
	return &middlewareExecutor{ms, 0, e}
}

// Execute executes middlewares and executor.
func (me *middlewareExecutor) Execute(c context.Context) error {
	me.index = 0
	defer func() {
		me.index = 0
	}()
	return me.Continue(c)
}

// Continue continues to execute the next middleware or executor.
func (me *middlewareExecutor) Continue(c context.Context) error {
	if me.index >= len(me.middlewares) {
		if me.executor != nil {
			return me.executor.Execute(c)
		}
		return nil
	}
	m := me.middlewares[me.index]
	me.index++
	return m(c, me)
}
