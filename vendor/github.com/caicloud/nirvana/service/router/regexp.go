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
	"regexp"
	"strings"
)

// index contains the key and it's index of the submatches.
type index struct {
	// Key is the name for the value.
	Key string
	// Pos is the index of value in submatches.
	Pos int
}

// regexpNode contains information for matching a regexp segment.
type regexpNode struct {
	handler
	children
	// indices contains all positions to get values from submatches.
	indices []index
	// exp is the regular expression.
	exp string
	// regexp is a regexp instance to match.
	regexp *regexp.Regexp
}

// Target returns the matching target of the node.
func (n *regexpNode) Target() string {
	return n.exp
}

// Kind returns the kind of the router node.
func (n *regexpNode) Kind() RouteKind {
	return Regexp
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *regexpNode) Match(ctx context.Context, c Container, path string) (Executor, error) {
	// Match self
	index := strings.IndexByte(path, '/')
	if index < 0 {
		index = len(path)
	}
	segment := path[:index]
	result := n.regexp.FindStringSubmatch(segment)
	if result == nil {
		return nil, RouterNotFound.Error()
	}
	// Match progeny
	var executor Executor
	var err error
	if index < len(path) {
		executor, err = n.children.Match(ctx, c, path[index:])
		if err == nil {
			executor, err = n.handler.pack(executor)
		}
	} else {
		executor, err = n.unionExecutor(ctx)
	}

	if err != nil {
		// Unmatched
		return nil, err
	}

	// Set values
	for _, i := range n.indices {
		c.Set(i.Key, result[i.Pos])
	}
	return executor, nil
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *regexpNode) Merge(r Router) (Router, error) {
	node, ok := r.(*regexpNode)
	if !ok {
		return nil, UnknownRouterType.Error(r.Kind(), reflect.TypeOf(r).String())
	}
	if n.exp != node.exp {
		return nil, UnmatchedRouterRegexp.Error(n.exp, node.exp)
	}
	if err := n.handler.Merge(&node.handler); err != nil {
		return nil, err
	}
	if err := n.children.merge(&node.children); err != nil {
		return nil, err
	}
	return n, nil
}

// fullMatchRegexpNode is an optimizing of RegexpNode.
type fullMatchRegexpNode struct {
	handler
	children
	// key is the name for the only value.
	key string
}

// Target returns the matching target of the node.
func (n *fullMatchRegexpNode) Target() string {
	return (&expSegment{FullMatchTarget, n.key}).Target()
}

// Kind returns the kind of the router node.
func (n *fullMatchRegexpNode) Kind() RouteKind {
	return Regexp
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *fullMatchRegexpNode) Match(ctx context.Context, c Container, path string) (Executor, error) {
	index := strings.IndexByte(path, '/')
	var executor Executor
	var err error
	if index > 0 {
		executor, err = n.children.Match(ctx, c, path[index:])
		if err == nil {
			executor, err = n.handler.pack(executor)
		}
	} else {
		index = len(path)
		executor, err = n.unionExecutor(ctx)
	}
	if err != nil {
		// Unmatched
		return nil, err
	}
	c.Set(n.key, path[:index])
	return executor, nil
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *fullMatchRegexpNode) Merge(r Router) (Router, error) {
	node, ok := r.(*fullMatchRegexpNode)
	if !ok {
		return nil, UnknownRouterType.Error(r.Kind(), reflect.TypeOf(r).String())
	}
	if n.key != node.key {
		return nil, UnmatchedRouterKey.Error(n.key, node.key)
	}
	if err := n.handler.Merge(&node.handler); err != nil {
		return nil, err
	}
	if err := n.children.merge(&node.children); err != nil {
		return nil, err
	}
	return n, nil
}
