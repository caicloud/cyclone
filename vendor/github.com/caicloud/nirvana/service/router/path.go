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
)

// pathNode matches all rest path.
type pathNode struct {
	handler
	// key is the key for the rest path.
	key string
}

// Target returns the matching target of the node.
func (n *pathNode) Target() string {
	return ""
}

// Kind returns the kind of the router node.
func (n *pathNode) Kind() RouteKind {
	return Path
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *pathNode) Match(ctx context.Context, c Container, path string) (Executor, error) {
	c.Set(n.key, path)
	return n.handler.unionExecutor(ctx)
}

// Merge merges r to the current router. The type of r should be same
// as the current one.
func (n *pathNode) Merge(r Router) (Router, error) {
	node, ok := r.(*pathNode)
	if !ok {
		return nil, UnknownRouterType.Error(r.Kind(), reflect.TypeOf(r).String())
	}
	if n.key != node.key {
		return nil, UnmatchedRouterKey.Error(n.key, node.key)
	}
	if err := n.handler.Merge(&node.handler); err != nil {
		return nil, err
	}
	return n, nil
}
