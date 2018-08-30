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
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// RoutingChain contains the call chain of middlewares and executor.
type RoutingChain interface {
	// Continue continues to execute the next middleware or executor.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call RoutingChain.Continue() and pass the context.
type Middleware func(context.Context, RoutingChain) error

// Inspector can select an executor to execute.
type Inspector interface {
	// Inspect finds a valid executor to execute target context.
	// It returns an error if it can't find a valid executor.
	Inspect(context.Context) (Executor, error)
}

// Executor executs with a context.
type Executor interface {
	// Execute executes with context.
	Execute(context.Context) error
}

// RouteKind is kind of routers.
type RouteKind string

const (
	// String means the router has a fixed string.
	String RouteKind = "String"
	// Regexp means the router has a regular expression.
	Regexp RouteKind = "Regexp"
	// Path means the router matches the rest. Path router only can
	// be placed at the leaf node.
	Path RouteKind = "Path"
)

// Container is a key-value container. It saves key-values from path.
type Container interface {
	// Set sets key-value into the container.
	Set(key, value string)
	// Get gets a value by key from the container.
	Get(key string) (string, bool)
}

// Router describes the interface of a router node.
type Router interface {
	// Target returns the matching target of the node.
	// It can be a fixed string or a regular expression.
	Target() string
	// Kind returns the kind of the router node.
	Kind() RouteKind
	// Match find an executor matched by path.
	// The context contains information to inspect executor.
	// The container can save key-value pair from the path.
	// If the router is the leaf node to match the path, it will return
	// the first executor which Inspect() returns true.
	Match(ctx context.Context, c Container, path string) (Executor, error)
	// AddMiddleware adds middleware to the router node.
	// If the router matches a path, all middlewares in the router
	// will be executed by the returned executor.
	AddMiddleware(ms ...Middleware)
	// Middlewares returns all middlewares of the router.
	// Don't modify the returned values.
	Middlewares() []Middleware
	// SetInspector sets inspector to the router node.
	SetInspector(inspector Inspector)
	// Inspector gets inspector from the router node.
	// Don't modify the returned values.
	Inspector() Inspector
	// Merge merges r to the current router. The type of r should be same
	// as the current one or it panics.
	//
	// For instance:
	//  Router A: /namespaces/ -> {namespace}
	//  Router B: /nameless/ -> {other}
	// Result:
	//  /name -> spaces/ -> {namespace}
	//       |-> less/ -> {other}
	Merge(r Router) (Router, error)
}

const (
	// FullMatchTarget is a match for full regular expression. All regexp router without expression
	// will use the expression.
	FullMatchTarget = ".*"
	// TailMatchTarget is a match expression for tail only.
	TailMatchTarget = "*"
)

// Parse parses a path to a router tree. It returns the root router and
// the leaf router. you can add middlewares and executor to the routers.
// A valid path should like:
//  /segments/{segment}/resources/{resource}
//  /segments/{segment:[a-z]{1,2}}.log/paths/{path:*}
func Parse(path string) (Router, Router, error) {
	paths, err := Split(path)
	if err != nil {
		return nil, nil, err
	}
	if len(paths) <= 0 {
		return nil, nil, InvalidPath.Error()
	}
	segments, err := reorganize(paths)
	if err != nil {
		return nil, nil, err
	}
	var root Router
	var leaf Router
	var parent Router
	for i, seg := range segments {
		router, err := segmentToRouter(seg)
		if err != nil {
			return nil, nil, err
		}
		if i == 0 {
			root = router
		}
		if i == len(segments)-1 {
			leaf = router
		}
		if parent != nil {
			if c, ok := parent.(interface {
				addRouter(router Router) error
			}); ok {
				if err := c.addRouter(router); err != nil {
					return nil, nil, err
				}
			} else {
				return nil, nil, InvalidParentRouter.Error(reflect.TypeOf(parent).String())
			}
		}
		parent = router
	}
	return root, leaf, nil
}

// segmentToRouter converts segment to a router.
func segmentToRouter(seg *segment) (Router, error) {
	switch seg.kind {
	case String:
		return &stringNode{
			prefix: seg.match,
		}, nil
	case Regexp:
		if len(seg.keys) == 1 && seg.match == (&expSegment{FullMatchTarget, seg.keys[0]}).Target() {
			return &fullMatchRegexpNode{
				key: seg.keys[0],
			}, nil
		}

		node := &regexpNode{
			exp: seg.match,
		}
		r, err := regexp.Compile("^" + seg.match + "$")
		if err != nil {
			return nil, InvalidRegexp.Error(seg.match)
		}
		node.regexp = r
		names := r.SubexpNames()
		j := 0
		for i := 0; i < len(names) && j < len(seg.keys); i++ {
			if names[i] == seg.keys[j] {
				node.indices = append(node.indices, index{names[i], i})
				j++
			}
		}
		if j != len(seg.keys) {
			return nil, UnmatchedSegmentKeys.Error(seg)
		}
		return node, nil

	case Path:
		return &pathNode{
			key: seg.keys[0],
		}, nil
	}
	return nil, UnknownSegment.Error(seg)
}

// Split splits string segments and regexp segments.
//
// For instance:
//  /segments/{segment:[a-z]{1,2}}.log/paths/{path:*}
// TO:
//  /segments/ {segment:[a-z]{1,2}} .log/paths/ {path:*}
func Split(path string) ([]string, error) {
	result := make([]string, 0, 5)
	lastElementPos := 0
	braceCounter := 0
	for i, c := range path {
		switch c {
		case '{':
			braceCounter++
			if braceCounter == 1 {
				if i > lastElementPos {
					result = append(result, path[lastElementPos:i])
				}
				lastElementPos = i
			}
		case '}':
			braceCounter--
			if braceCounter == 0 {
				result = append(result, path[lastElementPos:i+1])
				lastElementPos = i + 1
			}
		}
	}
	if braceCounter > 0 {
		return nil, UnmatchedPathBrace.Error()
	}
	if lastElementPos < len(path) {
		result = append(result, path[lastElementPos:])
	}
	return result, nil
}

// segment contains information to construct a router.
type segment struct {
	// match is the target string.
	match string
	// keys contains keys from segments.
	keys []string
	// kind is the router kind which the segment can be converted to.
	kind RouteKind
}

// reorganize reorganizes the form of paths.
//
// For instance:
//  /segments/ {segment:[a-z]{1,2}} .log/paths/ {path:*}
// To:
//  {/segments/ {} String} {(?P<segment>[a-z]{1,2})\.log {segment} Regexp} {/paths/ {} String} { {path} Path}
func reorganize(paths []string) ([]*segment, error) {
	segments := make([]*segment, 0, len(paths))
	var target *segment
	for i := 0; i < len(paths); i++ {
		p := paths[i]
		if !strings.HasPrefix(p, "{") {
			if target == nil {
				// String segment
				segments = append(segments, &segment{p, nil, String})
			} else {
				// Regexp segment
				slashPos := strings.Index(p, "/")
				if slashPos < 0 {
					// No slash
					target.match += regexp.QuoteMeta(p)
				} else {
					target.match += regexp.QuoteMeta(p[:slashPos])
					segments = append(segments, target, &segment{p[slashPos:], nil, String})
					target = nil
				}
			}
		} else {
			// Regexp segment
			seg, err := parseExpSegment(p)
			if err != nil {
				return nil, err
			}
			if seg.Tail() {
				if i != len(paths)-1 {
					return nil, InvalidPathKey.Error(seg.key)
				}
				if target != nil {
					segments = append(segments, target)
					target = nil
				}
				segments = append(segments, &segment{"", []string{seg.key}, Path})
				break
			}
			if target == nil {
				target = &segment{"", []string{}, Regexp}
			}
			target.match += seg.Target()
			target.keys = append(target.keys, seg.key)
		}
	}
	if target != nil {
		segments = append(segments, target)
	}
	return segments, nil
}

// expSegment describes a regexp segment.
type expSegment struct {
	// exp is the regular expression.
	exp string
	// key is the key for the expression.
	key string
}

// parseExpSegment parses a regexp segment to ExpSegment.
func parseExpSegment(exp string) (*expSegment, error) {
	if !strings.HasPrefix(exp, "{") || !strings.HasSuffix(exp, "}") {
		return nil, InvalidRegexp.Error(exp)
	}
	exp = exp[1 : len(exp)-1]
	pos := strings.Index(exp, ":")
	seg := &expSegment{}
	if pos < 0 {
		seg.exp = FullMatchTarget
		seg.key = exp
	} else {
		seg.exp = exp[pos+1:]
		seg.key = exp[:pos]
	}
	return seg, nil
}

// Tail returns whether the segment contains a tail match target.
func (es *expSegment) Tail() bool {
	return es.exp == TailMatchTarget
}

// Target returns the whole regular expression for the segment.
func (es *expSegment) Target() string {
	return fmt.Sprintf("(?P<%s>%s)", es.key, es.exp)
}
