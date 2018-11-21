<!-- START doctoc generated TOC please keep comment here to allow auto update -->

<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

**Table of Contents** _generated with [DocToc](https://github.com/thlorenz/doctoc)_

* [Introduction](#introduction)
  * [Background](#background)
  * [Goal](#goal)
* [Proposal](#proposal)
  * [Survey](#survey)
  * [Requirements](#requirements)
  * [None-requirements](#none-requirements)
  * [Design and Implementation](#design-and-implementation)
    * [Project Timeline](#project-timeline)
    * [Feature Design and Planning](#feature-design-and-planning)
    * [Meetings](#meetings)
    * [Assignees](#assignees)
  * [References](#references)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Introduction

api framework, or http web framework, is a long standing engineering backlog across engineering teams. The
proposal aims to shed some light on how we want to move forward. This is by no means a thorough design, we
hope to gather enough feedback to get going. Feel free to comment.

* Author: ddysher@
* Initial issue: https://github.com/caicloud/nirvana/issues/1

## Background

No guideline is provided at this moment to start up a new apiserver from scratch, results in a divergent
approaches on setting up new web projects. No conventions enforced at framework level also results in bugs
due to breaking apis, validation errors, etc. On the other hand, engineering effort is wasted on solving
the same range of problems which should ideally be solved in an API framework.

## Goal

* Reduce api level errors and inconsistency
* Improve engineering productivity via removing repeated work, adding code generation, etc
* Adding new resource type should only require defining struct definition
* Adding validation should only require declaring validation method as part of struct definition
* Consistent behavior, structure and layout across all golang server projects
* Out-of-box server instrumentation, e.g. metrics, tracing, profiling, etc

# Proposal

## Survey

Below is a short survey on existing projects:

* kubernetes-admin style (go-restful)
  * routing wrapper around go-restful
  * a well-defined layout
  * count for 70% of all related projects
* loadbalancer-admin style (gin)
  * use gin directly
  * no well-defined project layout
  * count for 20% of all related projects
* liubo style (go-restful)
  * mirrors closely with kubernetes style
  * count for 10% of all related projects

## Requirements

Following is all the required functionalities for our framework.

**Routing, Request & Response**

* Routes mapping from request to function
  * support for path parameter (e.g. /api/v1/animals/{id})
  * support for query parameter (e.g. /api/v1/animals?type=dog)
  * support regex
* Grouping Routes
  * routes can be grouped for better arragement, but doesn't have to be first class concept
* Request/Response API object
  * to/from structs in json
  * easy access to path,query,header parameters
* Middleware support
  * general middleware support is required for developer to add custom middleware
  * required default middleware: request logging, recovery, tracing
* Contextual process chain and parameter injection
  * all handlers accepts a context value representing a request context
  * requset parameters can be validated and injected into handler, see below
* API error convention
  * define canonical error code and type based on api convention
* multipart/urlencoded form support (low priority)
* file upload support (low priority)
* grpc support (low priority)

**Instrumentation**

* Provide default metrics at well-known endpoints for prometheus to scrape
  * need to define set of default metrics
* Tracing should be provided by default to allow better troubleshooting
  * follow opentracing, which is industry standard
  * use `jaeger` as tracing system
* Profiling can be enabled in debug mode for troubleshooting
  * use go profile

**Validation**

* Provide default validation on api types with json, e.g. required, within a range, etc
* Support custom validations defined by developers on api types
* support validation on all parameters (path, query, etc)

**Usability**

* A working project should be brought up with a few lines of code using the framework
* Framework must follow engineering conventions to help developers focus on business logic
* OpenAPI (swagger 2.0) specification can be generated automatically with no extra work
* Provides a well-established layout conforming to golang project layout
* Easy and standard configuration
* A reasonable support for websocket

**Performance**

* Performance is not a hard requirement for initial stage; but it should not introduce observable latency

## None-requirements

A list of functionalities commonly seen in other frameworks that are not in our scope

* https support. While it's not hard to serve https in go, https is out of scope since all services are expected to be running behind API gateway
* html rendering, orm, etc. Let's stick it to a simple restful framework
* testing. testing should be a different library

## Design and Implementation

### Project Timeline

* [x] Due 09.15: launch the project (@ddysher)
* [x] Due 09.23: draft feature set (@ddysher)
* [x] Due 09.29: finalize feature set and roadmap (@ddysher)
* [x] Due 12.01: check in most features (framework, cli, validation, metrics)
* [x] Due 12.20: check in all features, with reasonable test coverage (openapi, tracing)
* [x] Due 12.31: documentation, examples, website should be ready

### Feature Design and Planning

Following is a list of planned features, their respective design and planning.

* [Basic framework](https://github.com/caicloud/nirvana/issues/2)
* [API Validation](https://github.com/caicloud/nirvana/issues/3)
* [Configuration Management](https://github.com/caicloud/nirvana/issues/4)
* [OpenAPI Generation](https://github.com/caicloud/nirvana/issues/5)
* [Instrumentation and Profiling](https://github.com/caicloud/nirvana/issues/6)
* [Tracing](https://github.com/caicloud/nirvana/issues/7)

### Meetings

Regular meetings will be held to discuss above topics.

* [1st meeting](https://github.com/caicloud/nirvana/issues/1)
* [2nd meeting](https://github.com/caicloud/nirvana/issues/11)
* [3rd meeting](https://github.com/caicloud/nirvana/issues/12)
* [4th meeting](https://github.com/caicloud/nirvana/issues/26)
* [5th meeting](https://github.com/caicloud/nirvana/issues/38)
* [6th meeting](https://github.com/caicloud/nirvana/issues/61)

### Assignees

* Project tracking and proposal maintenance (@ddysher)
* Overall structure, project layout, etc (@ddysher)
* Routing, multiplexing, middleware, API error package, etc (@kdada)
* Configuration management (@zuomo)
* API validation (@walktall)
* OpenAPI generation (@liubog2008 @Jimexist )
* Prometheus instrumentation (@caitong93)
* tracing (@yejiayu)

## References

* [go-restful](https://github.com/emicklei/go-restful)
* [gin](https://github.com/gin-gonic/gin)
* [chi](https://github.com/go-chi/chi)
* [validator](https://github.com/go-playground/validator)
* [code-generator](https://github.com/kubernetes/code-generator)
* [prometheus](https://github.com/prometheus/prometheus)
* [jaeger](https://github.com/uber/jaeger)
