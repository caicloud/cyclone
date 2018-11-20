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

import "github.com/caicloud/nirvana/errors"

// Router `Match` method only can return errors from these error factory:
// RouterNotFound, NoInspector, NoExecutor.

// RouterNotFound means there no router matches path.
var RouterNotFound = errors.NotFound.Build("Nirvana:Router:RouterNotFound",
	"can't find router")

// NoInspector means there is no inspector in router.
var NoInspector = errors.NotFound.Build("Nirvana:Router:NoInspector",
	"no inspector to generate executor")

// NoExecutor means can't pack middlewares for nil executor.
var NoExecutor = errors.NotFound.Build("Nirvana:Router:NoExecutor",
	"no executor to pack middlewares")

// ConflictInspectors can build errors for failure of merging routers.
// If attempts to merge two router and they all have inspector, an error
// should be returned. A merged router can't have two inspectors.
var ConflictInspectors = errors.Conflict.Build("Nirvana:Router:ConflictInspectors",
	"can't merge two routers that all have inspector")

// EmptyRouterTarget means a router node has an invalid empty target.
var EmptyRouterTarget = errors.UnprocessableEntity.Build("Nirvana:Router:EmptyRouterTarget",
	"router ${kind} has no target")

// UnknownRouterType means a node type is unprocessable.
var UnknownRouterType = errors.UnprocessableEntity.Build("Nirvana:Router:UnknownRouterType",
	"router ${kind} has unknown type ${type}")

// UnmatchedRouterKey means a router's key is not matched with another.
var UnmatchedRouterKey = errors.UnprocessableEntity.Build("Nirvana:Router:UnmatchedRouterKey",
	"router key ${keyA} is not matched with ${keyB}")

// UnmatchedRouterRegexp means a router's regexp is not matched with another.
var UnmatchedRouterRegexp = errors.UnprocessableEntity.Build("Nirvana:Router:UnmatchedRouterRegexp",
	"router regexp ${regexpA} is not matched with ${regexpA}")

// NoCommonPrefix means two routers have no common prefix.
var NoCommonPrefix = errors.UnprocessableEntity.Build("Nirvana:Router:NoCommonPrefix",
	"there is no common prefix for the two routers")

// InvalidPath means router path is invalid.
var InvalidPath = errors.UnprocessableEntity.Build("Nirvana:Router:InvalidPath",
	"invalid path")

// InvalidParentRouter means router node has no method to add child routers.
var InvalidParentRouter = errors.UnprocessableEntity.Build("Nirvana:Router:InvalidParentRouter",
	"router ${type} has no method to add children")

// UnmatchedSegmentKeys means segment has unmatched keys.
var UnmatchedSegmentKeys = errors.UnprocessableEntity.Build("Nirvana:Router:UnmatchedSegmentKeys",
	"segment ${value} has unmatched keys")

// UnknownSegment means can't recognize segment.
var UnknownSegment = errors.UnprocessableEntity.Build("Nirvana:Router:UnknownSegment",
	"unknown segment ${value}")

// UnmatchedPathBrace means path has unmatched brace.
var UnmatchedPathBrace = errors.UnprocessableEntity.Build("Nirvana:Router:UnmatchedPathBrace",
	"unmatched braces")

// InvalidPathKey means path key must be the last element.
var InvalidPathKey = errors.UnprocessableEntity.Build("Nirvana:Router:InvalidPathKey",
	"key ${key} should be last element in the path")

// InvalidRegexp means regexp is not notmative.
var InvalidRegexp = errors.UnprocessableEntity.Build("Nirvana:Router:InvalidRegexp",
	"regexp ${regexp} does not have normative format")
