<!-- START doctoc generated TOC please keep comment here to allow auto update -->

<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

**Table of Contents** _generated with [DocToc](https://github.com/thlorenz/doctoc)_

* [API Framework](#api-framework)
  * [Background](#background)
  * [Design Assumptions](#design-assumptions)
  * [Proposed Design](#proposed-design)
    * [Log](#log)
    * [Error](#error)
    * [Tracing](#tracing)
    * [Router](#router)
    * [Middleware](#middleware)
    * [Handler](#handler)
  * [Architecture Overview](#architecture-overview)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# API Framework

* Author: @kdada
* Initial issue: https://github.com/caicloud/nirvana/issues/2

## Background

We need a uniform api framework to standarize our api project.

## Design Assumptions

* API projects run in internal environment. They don't care about connection security.
* API projects are modularized and have low coupling with engaged framework.
* APIs are not performance sensitive.
* APIs are restful.
* No html render requirements.
* No session requirements.
* No streaming uploading requirements.

## Proposed Design

The core components of api framework should have a router and a handler specification. Router is responsible
for dispatching requests to handlers. Then handlers validate requests and inject parameters to api methods
defined by users. After the execution of api methods, handlers analyze the returned values and write
corresponding data to responses.

The main components with dependencies:

* Router
  * Log
  * Error
  * Middleware
* Handler
  * Log
  * Error
  * Validator
  * Injector
  * Context
  * Tracing
* Client
  * Log
  * Error
  * Tracing

### Log

A standard logger interface:

```go
type Verboser interface {
	Info(...interface{})
	Infof(...interface{})
	Infoln(...interface{})
}
type Logger interface {
	V(int) Verboser
	Info(...interface{})
	Infof(...interface{})
	Infoln(...interface{})
	Warning(...interface{})
	Warningf(...interface{})
	Warningln(...interface{})
	Error(...interface{})
	Errorf(...interface{})
	Errorln(...interface{})
	Fatal(...interface{})
	Fatalf(...interface{})
	Fatalln(...interface{})
}
```

### Error

A standard error interface:

```go
type Error interface {
	Code() int
	Message() interface{}
}
```

### Tracing

Tracing system could be [Zipkin](http://zipkin.io/) or [Jeager](https://uber.github.io/jaeger/). The two
implement [opentracing](http://opentracing.io/). And we will integrate opentracing via
[opentracing-go](https://github.com/opentracing/opentracing-go).

### Router

URL path is a long continuous string. But we can convert it to an array of segments. For example, we have an
instance of url path:

```
/collections/object/subresources/resource
```

We can split the path by '/', then we get an array:

```
[collections, object, subresources, resource]
```

Router is a tree. Like this:

```
/
|-- collections
|---- (object:*)
|------ subresources
|-------- (resource:*)
|-- others
```

Every node should match one segement. Pass the array to root node of router, it finds the leaf node matched
the last segement. Then the handler of the leaf node handles the request.

In the example, It shows we need support two kinds of router nodes:

* String node
  A string router can match a fixed segment of url path.
* Regexp node
  A regexp router can match multiple forms of segment.

The handler of node should implement:

```
type Handler interface {
	Handle(context.Context)
}
```

### Middleware

As saying above, a router node can have one handler. But it can have multiple middlewares, middlewares decides
the execution of router.

```
type Middleware interface {
	Prerouting(context.Context) bool
	Postrouting(context.Context) bool
}
```

Middleware can cancel the routing process and return immediately.

**Alternative Middleware**

```
type RoutingChain interface {
	Continue(context.Context)
}

type Middleware interface {
	Handle(context.Context, RoutingChain) bool
}
```

### Handler

```go
type FromType string

const (
	FromPath   FromType = "FromPath"
	FromHeader FromType = "FromHeader"
	FromForm   FromType = "FromForm"
	FromBody   FromType = "FromBody"
)

type ResultType string

const (
	DataResult ResultType = "DataResult"
)

type Parameter struct {
	From     FromType
	Name     string
	Document string
	Type     string
	Required bool
	Default  interface{}
}

type Result struct {
	Type ResultType
}

type Function struct {
	Pointer    interface{}
	Parameters []Parameter
	Results    []Result
	Errors     []Error
}

type ContentType struct {
	Acceptable []string
	Generable  []string
}

type Definition struct {
	HTTPMethod  string
	Summary     string
	Document    string
	ContentType ContentType
	Function    Function
}

type Descriptor struct {
	Path        string
	Middlewares []Middleware
	Definitions []Definition
	Children    []Descriptor
}
```

## Architecture Overview

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34514829-d9a9b43c-f034-11e7-9dee-5c7940755fab.png" height="350px" width="auto"></p>

`log` and `errors` are based on other packages, so we don't explain it in the diagram.

`nirvana` is the root package in the framework. It implements a configurable server structure. All plugins
implement config interface:

```go
// ConfigInstaller is used to install config to service builder.
type ConfigInstaller interface {
	// Name is the external config name.
	Name() string
	// Install installs stuffs before server starting.
	Install(builder service.Builder, config *Config) error
	// Uninstall uninstalls stuffs after server terminating.
	Uninstall(builder service.Builder, config *Config) error
}
```

The startup flow:

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34515618-8cbc6df0-f038-11e7-90b5-71dd770815a2.png" height="350px" width="auto"></p>

Definition modifier is special. It only works in service builder and is used to modify definitions. That means
if you install some API definitions into nirvana server, you have a chance to modify all definitions globally.
The most important usage of modifier is to inject `context.Context` into `Definition.Parameters` as first
parameter.

The shutdown flow:

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34515419-a2dc02d6-f037-11e7-95ec-566efcc193f7.png" height="350px" width="auto"></p>

All plugins are independent, so plugins are installed and uninstalled without order. In this stage, plugins
can close file descriptors or close network connections.

The lifecycle of requests:

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34516016-900f165e-f03a-11e7-9bad-536cc3ac906f.png" height="350px" width="auto"></p>

`ParameterGenerators` are generated from `Definition.Parameters`. A parameter corresponds to a `ParameterGenerator`.

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34516454-7215cda8-f03c-11e7-8fcf-e06147c9d98d.png" height="350px" width="auto"></p>

Every parameter should go through the flow. Operators execute one by one. The original data passes to first
operator, then the result of first operator is as the parameter to pass to the second operator. The last
operator's result is as the final data to user function. If there are no operators, typed data is as the
final data.

The workflow of `DestinationHandlers` is like `ParameterGenerators`. The returned value of user function is
passed to operators. Then the returned value of last operator is as final data to client.
