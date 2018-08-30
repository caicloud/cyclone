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
	"strings"
)

// stringNode describes a string router node.
type stringNode struct {
	handler
	children
	// prefix is the fixed string to match path.
	prefix string
}

// Target returns the matching target of the node.
func (n *stringNode) Target() string {
	return n.prefix
}

// Kind returns the kind of the router node.
func (n *stringNode) Kind() RouteKind {
	return String
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *stringNode) Match(ctx context.Context, c Container, path string) (Executor, error) {
	if n.prefix != "" && !strings.HasPrefix(path, n.prefix) {
		// No match
		return nil, RouterNotFound.Error()
	}
	if len(n.prefix) < len(path) {
		// Match prefix
		executor, err := n.children.Match(ctx, c, path[len(n.prefix):])
		if err != nil {
			return nil, err
		}
		return n.handler.pack(executor)
	}
	// Match self
	return n.handler.unionExecutor(ctx)
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *stringNode) Merge(r Router) (Router, error) {
	node, ok := r.(*stringNode)
	if !ok {
		return nil, UnknownRouterType.Error(r.Kind(), reflect.TypeOf(r).String())
	}
	commonPrefix := 0
	for commonPrefix < len(n.prefix) && commonPrefix < len(node.prefix) {
		if n.prefix[commonPrefix] != node.prefix[commonPrefix] {
			break
		}
		commonPrefix++
	}
	if commonPrefix <= 0 {
		return nil, NoCommonPrefix.Error()
	}
	switch {
	case commonPrefix == len(n.prefix) && commonPrefix == len(node.prefix):
		if err := n.handler.Merge(&node.handler); err != nil {
			return nil, err
		}
		if err := n.children.merge(&node.children); err != nil {
			return nil, err
		}
	case commonPrefix == len(n.prefix):
		node.prefix = node.prefix[commonPrefix:]
		if err := n.addRouter(node); err != nil {
			return nil, err
		}
	case commonPrefix == len(node.prefix):
		copy := *n
		copy.prefix = copy.prefix[commonPrefix:]
		*n = *node
		if err := n.addRouter(&copy); err != nil {
			return nil, err
		}
	default:
		copy := *n
		copy.prefix = copy.prefix[commonPrefix:]
		node.prefix = node.prefix[commonPrefix:]
		n.handler = handler{}
		n.children = children{}
		n.prefix = n.prefix[:commonPrefix]
		if err := n.addRouter(&copy); err != nil {
			return nil, err
		}
		if err := n.addRouter(node); err != nil {
			return nil, err
		}
	}
	return n, nil
}
