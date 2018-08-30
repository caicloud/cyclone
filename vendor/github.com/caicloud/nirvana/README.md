# Nirvana

<img align="right" width="225px" src="https://user-images.githubusercontent.com/2191361/35839723-e9e5cdfa-0b2c-11e8-853a-8d3870f9e7ac.png">

[![Build Status](https://travis-ci.org/caicloud/nirvana.svg?branch=master)](https://travis-ci.org/caicloud/nirvana)
[![Coverage Status](https://coveralls.io/repos/github/caicloud/nirvana/badge.svg?branch=master)](https://coveralls.io/github/caicloud/nirvana?branch=master)
[![GoDoc](http://godoc.org/github.com/caicloud/nirvana?status.svg)](http://godoc.org/github.com/caicloud/nirvana)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/nirvana)](https://goreportcard.com/report/github.com/caicloud/nirvana)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

Nirvana is a golang API framework designed for productivity and usability. It aims to be the building block for
all golang services in Caicloud. The high-level goals and features include:

* consistent API behavior, structure and layout across all golang projects
* improve engineering productivity with openAPI and client generation, etc
* validation can be added by declaring validation method as part of API definition
* out-of-box instrumentation support, e.g. metrics, profiling, tracing, etc
* easy and standard configuration management, as well as standard cli interface

Nirvana is also extensible and performant, with the goal to support fast developmenet velocity.

## Installation

```
go get -u github.com/caicloud/nirvana

# for openapi generation
go get -u github.com/caicloud/nirvana/cmd/openapi-gen
```

## Getting Started

### API quick start

In Nirvana, APIs are defined via `definition.Descriptor`. We will not introduce details of the concept `Descriptor`,
instead, let's take a look at a contrived example:

```go
// API descriptor.
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method: definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEText},
			Parameters: []definition.Parameter{
				{
					Source: definition.Query,
					Name: "msg",
					Description: "Corresponding to the second parameter",
				},
			},
			Results: []definition.Result{
				{
					Destination: definition.Data,
					Description: "Corresponding to the first result",
				},
				{
					Destination: definition.Error,
					Description: "Corresponding to the second result",
				},
			},
		},
	},
}
```

This is an echo server API descriptor. The descriptor is a bit complex at first glance, but is actually quite
simple. Below is a partially translated HTTP language:

```
HTTP Path: /echo[?msg=]
HTTP Method: Get
HTTP Headers:
    Content-Type: Any Type
    Accept: text/plain or */*
```

The request handler `Echo` receives two parameters and returns two results, as defined in our descriptor.
Note the first parameter is always `context.Context` - it is injected by default config.

```go
// API function.
func Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}
```

Nirvana will parse incoming request and generate function parameters for `Echo` function as defined via
`Definition.Parameters` - parameters will be converted into the exact type defined in `Echo`. Once done,
Nirvana collects the results and sends back response.

With our API descriptors ready, we can now create a server to serve requests:

```go
package main

import (
	"context"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

func main() {
	config := nirvana.NewDefaultConfig()
	config.Configure(nirvana.Descriptor(echo))
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}
}
```

Now run the server and test it:

```
go run ./examples/getting-started/basics/echo.go
INFO  0202-16:34:38.663+08 echo.go:65 | Listening on :8080
INFO  0202-16:34:38.663+08 builder.go:163 | Definitions: 1 Middlewares: 0 Path: /echo
INFO  0202-16:34:38.663+08 builder.go:178 |   Method: Get Consumes: [*/*] Produces: [text/plain]
```

In another terminal:

```
$ curl "http://localhost:8080/echo?msg=test"
test
```

For full example code, see [basics](./examples/getting-started/basics).

### Validate it!

Now you are tired of echoing non-sense testing message and want to only reply message longer than 10 characters, such
validation can be easily added when defining your descriptor:

```go
Parameters: []definition.Parameter{
	{
		Source:      definition.Query,
		Name:        "msg",
		Description: "Corresponding to the second parameter",
		Operators:   []definition.Operator{validator.String("gt=10")},
	},
},
```

`Operator` is a concept in Nirvana to allow framework user to operate on input request; validation is one of several
pre-defined operators. Another example of `operator` is `convertor`, which allows user to convert between different
versions of an input.

Under the hood, Nirvana uses [go-playground/validator.v9](https://github.com/go-playground/validator) for validation,
which defines a list of useful tags. It also supports custom validation. Nirvana integrates smoothly with the package,
see user guide for more advanced usage.

Now run our new echo server and verify validation works:

```
$ go run ./examples/getting-started/validator/echo.go
INFO  0202-11:18:50.235+08 echo.go:67 | Listening on :8080
INFO  0202-11:18:50.235+08 builder.go:163 | Definitions: 1 Middlewares: 0 Path: /echo
INFO  0202-11:18:50.235+08 builder.go:178 |   Method: Get Consumes: [*/*] Produces: [text/plain]
```

In another terminal:

```
$ curl "http://localhost:8080/echo?msg=test"
Key: '' Error:Field validation for '' failed on the 'gt' tag

$ curl "http://localhost:8080/echo?msg=testtesttest"
testtesttest
```

It works! The above example teaches us two facts:

1. Adding validation support with Nirvana is very simple
2. 10 characters validation is not enough to prevent spam :)

For full example code, see [validator](./examples/getting-started/validator). Checkout the source code to see
how to add your own validation.

### Is it popular?

It's time to expose some metrics to help understand and diagnose our service! Nirvana has out-of-box support for
instrumentation. To enable exposing request metrics, just add one more configuration:

```go
config := nirvana.NewDefaultConfig().
	Configure(
		metrics.Path("/metrics"),
	)
```

The actual configuration is done with `metrics` plugin. `plugin` is another concept in Nirvana - we can always
add more functionalities to Nirvana via plugin, and each plugin can be individually enabled or disabled. How
plugins are implemented depends on plugin author. For example, some plugins are simply static configuration,
while some are more complex middlewares. All plugins are registered into config. The server will install them
when the server starts.

Now if we start our server, we'll see a wealth of information exposed as [prometheus](https://prometheus.io) format.
The metrics are exposed via `/metrics` endpoint.

```
$ go run ./examples/getting-started/metrics/echo.go
```

Use ab (ApacheBench) to simulate some user load; in another terminal, run:

```
ab -n 1000 'http://localhost:8080/echo?msg=testtesttest'
```

Once done, let's checkout some default metrics from metrics plugin. The metric `nirvana_request_count` tells
us how many requests we've seen in total. Since we use `-n 1000`, there will be 1000 requests to `/echo` endpoint.

```
$ curl http://localhost:8080/metrics 2>&1 | grep nirvana_request_count
# HELP nirvana_request_count Counter of server requests broken out for each verb, API resource, client, and HTTP response contentType and code.
# TYPE nirvana_request_count counter
nirvana_request_count{client="ApacheBench/2.3",code="200",contentType="",method="GET",path="/echo"} 1000
```

The metric `nirvana_request_latencies` shows distribution of our service latencies. We've added a random sleep
between [0, 100) in our service; therefore, p90 is around 90m.

```
$ curl http://localhost:8080/metrics 2>&1 | grep "nirvana_request_latencies"
# HELP nirvana_request_latencies Response latency distribution in milliseconds for each verb, resource and client.
# TYPE nirvana_request_latencies histogram
nirvana_request_latencies_bucket{method="GET",path="/echo",le="0.1"} 11
nirvana_request_latencies_bucket{method="GET",path="/echo",le="0.2"} 11
nirvana_request_latencies_bucket{method="GET",path="/echo",le="0.4"} 11
nirvana_request_latencies_bucket{method="GET",path="/echo",le="0.8"} 11
nirvana_request_latencies_bucket{method="GET",path="/echo",le="1.6"} 28
nirvana_request_latencies_bucket{method="GET",path="/echo",le="3.2"} 41
nirvana_request_latencies_bucket{method="GET",path="/echo",le="6.4"} 73
nirvana_request_latencies_bucket{method="GET",path="/echo",le="12.8"} 126
nirvana_request_latencies_bucket{method="GET",path="/echo",le="25.6"} 260
nirvana_request_latencies_bucket{method="GET",path="/echo",le="51.2"} 507
nirvana_request_latencies_bucket{method="GET",path="/echo",le="102.4"} 995
nirvana_request_latencies_bucket{method="GET",path="/echo",le="204.8"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="409.6"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="819.2"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="1638.4"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="3276.8"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="6553.6"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="13107.2"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="26214.4"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="52428.8"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="+Inf"} 1000
nirvana_request_latencies_sum{method="GET",path="/echo"} 50554
nirvana_request_latencies_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="0.1"} 0
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="0.2"} 0
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="0.4"} 0
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="0.8"} 0
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="1.6"} 0
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="3.2"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="6.4"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="12.8"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="25.6"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="51.2"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="102.4"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="204.8"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="409.6"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="819.2"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="1638.4"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="3276.8"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="6553.6"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="13107.2"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="26214.4"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="52428.8"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="+Inf"} 1
nirvana_request_latencies_sum{method="GET",path="/metrics"} 3
nirvana_request_latencies_count{method="GET",path="/metrics"} 1
# HELP nirvana_request_latencies_summary Response latency summary in microseconds for each verb and resource.
# TYPE nirvana_request_latencies_summary summary
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.5"} 55
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.9"} 90
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.99"} 101
nirvana_request_latencies_summary_sum{method="GET",path="/echo"} 50554
nirvana_request_latencies_summary_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.5"} 3
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.9"} 3
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.99"} 3
nirvana_request_latencies_summary_sum{method="GET",path="/metrics"} 3
nirvana_request_latencies_summary_count{method="GET",path="/metrics"} 1
```

See user guide for more information about metrics plugin (and others). For full example code, see [metrics](./examples/getting-started/metrics).

### Show me the doc

You've upgraded your service to provide a new endpoint to create an echo message, i.e.

```
curl -H "Content-Type: application/json" -X POST -d '{"name": "alice", "message": "echo to myself"}' http://localhost:8080/echo
```

This is a complicated enpoint. To make it easy for your user, you decide to provide API documentation.
Nirvana has built-in support to generate openapi documentation. To generate the docs, you need to first
define where types come from. In our example, it's in the `api` package:

```go
package api

// Message defines the message to echo and to whom the message will be sent.
// +nirvana:openapi=true
type Message struct {
	Name    string `json:"name" validate:"required"`
	Message string `json:"message" validate:"gt=10"`
}
```

Next step is to generate openapi definitions:

```
openapi-gen \
  -i github.com/caicloud/nirvana/examples/getting-started/openapi/pkg/api \
  -p github.com/caicloud/nirvana/examples/getting-started/openapi/pkg/api
```

Finally, we can build our openapi specification:

```go
swagger, err := builder.BuildOpenAPISpec(&echo, &common.Config{
	Info: &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "echo server openAPI",
			Description: "This is open API documentation of echo server",
			Contact: &spec.ContactInfo{
				Name: "nirvana",
				URL:  "https://gonirvana.io",
			},
			License: &spec.License{
				Name: "Apache License, Version 2.0",
				URL:  "http://www.apache.org/licenses/LICENSE-2.0",
			},
			Version: "v1.0.0",
		},
	},
	GetDefinitions: api.GetOpenAPIDefinitions,
})
if err != nil {
	panic(err)
}
encoder := json.NewEncoder(os.Stdout)
if err := encoder.Encode(swagger); err != nil {
	panic(err)
}
```

Now run the following command, we can generate our swagger.json file. Put it into https://editor.swagger.io/,
we'll be able to view our generated API docs.

```
go run ./examples/getting-started/openapi/echo.go > /tmp/swagger.json
```

For full example code, see [openapi](./examples/getting-started/openapi).

## User Guide

### API Descriptor

API Descriptor is the core data structure in Nirvana: it holds all API definitions, and is usually the starting
point to write your services with Nirvana. Following is the golang type definition of `Descriptor`:

```go
// Descriptor describes a descriptor for API definitions.
type Descriptor struct {
	// Path is the url path. It will inherit parent's path.
	//
	// If parent path is "/api/v1", current is "/some",
	// It means current definitions handles "/api/v1/some".
	Path string
	// Consumes indicates content types that current definitions
	// and child definitions can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types that current definitions
	// and child definitions can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Definitions contains definitions for current path.
	Definitions []Definition
	// Children is used to place sub-descriptors.
	Children []Descriptor
	// Description describes the usage of the path.
	Description string
}
```

**A single descriptor contains API definitions for a single path.** It sets `Content-Type` to be produced and
consumed by the path handler. Each descriptor has an array of children; they will all inherit `Content-Type`
from the parent descriptor, for example:

```go
definition.Descriptor{
	Path:        "/path",
	Consumes:    []string{definition.MIMEAll},
	Produces:    []string{definition.MIMEText},
	Definitions: SomeDefinitions,
	Children: []definition.Descriptor{
		{
			Path:        "/child",
			Produces:    []string{definition.MIMEJSON},
			Definitions: SomeDefinitions,
		},
	},
}
```

The child descriptor is identical to:

```go
definition.Descriptor{
	Path:        "/path/child",
	Consumes:    []string{definition.MIMEAll},
	Produces:    []string{definition.MIMEJSON},
	Definitions: SomeDefinitions,
}
```

### Consumes and Produces

Consumes and Produces indicate content types that current definitions and child definitions support. Following
is a table of all supported MIME types and their data types:

| MIME            | Consume                        | Produce                        | Note                                                               |
| --------------- | ------------------------------ | ------------------------------ | ------------------------------------------------------------------ |
| MIMENone        | nil                            | nil                            | Can be used into `Consumes` of Get/List and `Produces` of `Delete` |
| MIMEText        | string/[]byte/io.Reader        | string/[]byte/io.Reader        |                                                                    |
| MIMEJSON        | string/[]byte/io.Reader/struct | string/[]byte/io.Reader/struct |                                                                    |
| MIMEXML         | string/[]byte/io.Reader/struct | string/[]byte/io.Reader/struct |                                                                    |
| MIMEOctetStream | string/[]byte/io.Reader        | string/[]byte/io.Reader        |                                                                    |
| MIMEURLEncoded  | nil                            | nil                            | Depends on `Source`. Only be used in `Consumes`                    |
| MIMEFormData    | nil                            | nil                            | Depends on `Source`. Only be used in `Consumes`                    |

### Middleware

Middleware is a convenient mechanism to intercept HTTP requests entering your application. To use middleware
in Nirvana, just add your middlewaare definition to API descriptor. For example, below is the code snippet for
metrics plugin:

```
monitorMiddleware := definition.Descriptor{
	Path:        "/",
	Middlewares: []definition.Middleware{newMetricsMiddleware(c.namespace)},
}

func newMetricsMiddleware(namespace string) definition.Middleware {
	...

	// This is the middleware function to be called for each request.
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)

		httpCtx := service.HTTPContextFrom(ctx)
		req := httpCtx.Request()
		resp := httpCtx.ResponseWriter()
		path := req.URL.Path
		elapsed := float64((time.Since(startTime)) / time.Millisecond)

		requestCounter.WithLabelValues(req.Method, path, getHTTPClient(req), req.Header.Get("Content-Type"), strconv.Itoa(resp.StatusCode())).Inc()
		requestLatencies.WithLabelValues(req.Method, path).Observe(elapsed)
		requestLatenciesSummary.WithLabelValues(req.Method, path).Observe(elapsed)

		return err
	}
}
```

Usually, Nirvana users do not care about how middlewares are implemented: they only need to find useful
middlewares and add them to their descriptors. But if necessary, writing your own middleware is also quite
straightforward, as shown above.

Unlike `Consumes` or `Produces`, middlewares are not scoped within a single descriptor, which means a
middleware for `/some/path` will impact all paths with prefix `/some/path`, even though they are in different
descriptors. For example:

```go
definition.Descriptor{
	Path:        "/path",
	Middlewares: SomeMiddlewares,
}
definition.Descriptor{
	Path:        "/path/child",
}
```

The two descriptors do not have any relationship but their path have common prefix, i.e. path of the first
descriptor is a prefix of the second descriptor. In such case, `SomeMiddlewares` are also valid for the second
descriptor. For more details, check the design doc of router.

### API Definition

API definition is another core data structure in Nirvana: it defines all handlers for your services. Following
is the golang type definition of `Definition`:

```go
// Definition defines an API handler.
type Definition struct {
	// Method is definition method.
	Method Method
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Function is a function handler. It must be func type.
	Function interface{}
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Description describes the API handler.
	Description string
	// Examples contains many examples for the API handler.
	Examples []Example
}
```

Each descriptor has multiple API definitions, and **A single API definition contains handler for a single path
and method combination.** For example, here we define a descriptor to handle endpoint `/echo`, with two methods
`Get` and `Create`:

```
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: EchoGet,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
		},
		{
			Method:   definition.Create,
			Function: EchoCreate,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
		},
	},
}
```

Below is a list of all supported methods, as well as its corresponding HTTP method and success status code. By
convention, every API method corresponds to a HTTP method and **ONE** success status code. If an API function
returns no error, Nirvana will return the success status code.

| Method      | HTTP Method | Success Status Code |
| ----------- | ----------- | ------------------- |
| List        | GET         | 200                 |
| Get         | GET         | 200                 |
| Create      | POST        | 201                 |
| Update      | PUT         | 200                 |
| Patch       | PATCH       | 200                 |
| Delete      | DELETE      | 204                 |
| AsyncCreate | POST        | 202                 |
| AsyncUpdate | PUT         | 202                 |
| AsyncPatch  | PATCH       | 202                 |
| AsyncDelete | DELETE      | 202                 |

### Parameter

`Parameter` describes corresponding handler parameters of an API definition. Your request handler will receive
the exact number of parameters, with the same index as defined in your API definition. Note most of the times,
you will start your service using `nirvana.NewDefaultConfig()`, which adds request context as the first
parameter. Therefore, parameters defined in descriptor appear in the second parameter of your request handler.
For example, in the following example, our endpoint `/echo` has two query parameters, and our handler `Echo`
receives three parameters: `context`, `msg1` and `msg2`.

```
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg1",
					Description: "First message to echo",
				},
				{
					Source:      definition.Query,
					Name:        "msg2",
					Description: "Second message to echo",
				},
			},
			Results: []definition.Result{
				{
 					Destination: definition.Data,
					Description: "Result to return if success",
				},
				{
					Destination: definition.Error,
					Description: "Error to return if not success",
				},
			},
		},
	},
}

// API function.
func Echo(ctx context.Context, msg1 string, msg2 string) (string, error) {
	return msg, nil
}
```

Below is the golang type definition of `Parameter`:

```go
// Parameter describes a function parameter.
type Parameter struct {
	// Source is the parameter value generated from.
	Source Source
	// Name is the name to get value from a request.
	// ex. a query name, a header key, etc.
	Name string
	// Default value is used when a request does not provide a value
	// for the parameter.
	Default interface{}
	// Operators can modify and validate the target value.
	// Parameter value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to target function.
	Operators []Operator
	// Description describes the parameter.
	Description string
}
```

`Source` is the source of a parameter, and `Name` is the key of `Source`.

| Source | Description                                                                                               |
| ------ | --------------------------------------------------------------------------------------------------------- |
| Path   | Value from URL path                                                                                       |
| Query  | Value from URL query string                                                                               |
| Header | Value from HTTP request header                                                                            |
| Form   | Value from HTTP body. `Content-Type` must be "application/x-www-form-urlencoded" or "multipart/form-data" |
| File   | Value from HTTP body. `Content-Type` must be "multipart/form-data"                                        |
| Body   | Value from HTTP body. Parameters of this type don't need a name                                           |
| Auto   | Data receiver must be a struct. Parameters of the type don't need a name.                                 |
| Prefab | Value from internal method. See `Advanced Usage`                                                          |

The source **Auto** is for combining fields in a struct:

```go
// Here is an example for `Auto` struct.
// The struct has some fields. Every field has a tag with name `source`.
// The source should obey the format:
//     Source,Name[,default=value]
// `Source` and `Name` are the same as before.
// `default` is optional. its value should be basic data type (bool, int*, uint*, float*, string).
type Example struct {
	ID     int    `source:"Path,id"`
	Start  int    `source:"Query,id,default=100"`
	Tenant string `source:"Header,X-Tenant,default=test"`
}
```

If you have lots of fields from a request, you can use `Auto` with a struct to get values from request.
Don't use it when you only have a few parameters: separated parameters is more readable.

All values from HTTP request are string. Nirvana has a mechanism to convert strings to specific types for
API function. The behavior is customizable via `operator`, which allows you to modify input request. In case
there is custom operator, input request will be converted to parameter type of the first operator. Here is
the data flow for a parameter:

<p align="center"><img src="https://user-images.githubusercontent.com/13895988/34516454-7215cda8-f03c-11e7-8fcf-e06147c9d98d.png" height="350px" width="auto"></p>

If `Data` is empty and `Parameter.Default` is not nil, default value is used as `Typed Data` .

### Result

`Result` is similar but simpler than `Parameter`. Its `Destination` indicates the target to write data. Just
like `Parameter`, we can modify output response via `operator`; the final returned type will be the return
type of the last operator.

```go
// Result describes how to handle a result from function results.
type Result struct {
	// Destination is the target for the result. Different types make different behavior.
	Destination Destination
	// Operators can modify the result value.
	// Result value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to destination handler.
	Operators []Operator
	// Description describes the result.
	Description string
}
```

| Destination | Description                                                                                                                   |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Meta        | Indicates the value should be written to HTTP response header. Its type must be `map[string]string`                           |
| Data        | Indicates the value should be written to HTTP response body. The format is decided by HTTP `Accept` and `Definition.Produces` |
| Error       | If an error occurs, `Meta` and `Data` is ignored. Error message will be written to HTTP response body                         |

### Validation

Validation is used to validate request input, including request body, query parameter, etc. In Nirvana,
validation is implemented as a parameter operator, so it naturally has access to all request attributes.
There are three categories of validation: Var, Struct and Custom, where Var is used to validate basic
built-in types like `string`, `int`, `bool`, etc; Struct is for struct validation and Custom is for writing
custom validation.

For Var validation, simply add the validation operator including the type to validate. For example, the
following example shows a validation used to validate input string length is longer than 10 characters.

```go
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
					Operators:   []definition.Operator{validator.String("gt=10")},
				},
			},
			...
		},
	},
}

// API function.
func Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}
```

Note we are using `Validator.String` here since our API handler takes string as input. As an other example,
if we want to validate input parameter is a number larger than 10, we should use:

```go
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Message to echo",
					Operators:   []definition.Operator{validator.Int("gt=10")},
				},
			},
			...
		},
	},
}

// API function.
func Echo(ctx context.Context, msg int) (string, error) {
	return strconv.Itoa(msg), nil
}
```

Here we've changed validator to `validator.Int`, and API handler has input parameter `int`.

For Struct validation, the first step is to add a `validate` tag to our struct, e.g.

```go
// Message defines the message to echo and to whom the message will be sent.
type Message struct {
	Name    string `json:"name" validate:"required"`
	Message string `json:"message" validate:"gt=10"`
}
```

Then, similar to Var validation, we need to add an operator to our API descriptor. A struct instance is
required for Nirvana to make sure the type to validate actually matches handler parameter type.

```go
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Create,
			Function: EchoV2,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Body,
					Name:        "msg",
					Description: "Message to echo",
					Operators:   []definition.Operator{validator.Struct(&api.Message{})},
				},
			},
			...
		},
	},
}

// API function.
func EchoV2(ctx context.Context, msg *api.Message) (string, error) {
	return msg.Message, nil
}
```

For Custom validation, you'll write your own operator and use it in API descriptor. The `operators/validator`
package contains helper funtions to create custom validator. For example, the following example uses custom
validation to validate the input request body. Nirvana will convert input request to validator's parameter
type.

```go
Operators: []definition.Operator{
	validator.NewCustom(
		func(ctx context.Context, body *Body) error {
			if body.Name == "" {
				return errors.BadRequest.Error("you should have a name!")
			}
			if body.Name != "nirvana" {
				return errors.BadRequest.Error("name ${name} must be nirvana!", body.Name)
			}
			return nil
		},
		"validate your name"),
},
```

### OpenAPI

Nirvana can generate OpenAPI 2.0 document from code simply.

In the example, swagger will be generated by builder of OpenAPI spec. 
There are two parts of the code, one is meta info and the other is the generated function `GetOpenAPIDefinitions`

```go
// swagger is the struct which can be encoded into whole OpenAPI document
swagger, err := builder.BuildOpenAPISpec(&yourDescriptor, &common.Config{
	Info: &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "echo server openAPI",
			Description: "This is open API documentation of echo server",
			Contact: &spec.ContactInfo{
				Name: "nirvana",
				URL:  "https://gonirvana.io",
			},
			License: &spec.License{
				Name: "Apache License, Version 2.0",
				URL:  "http://www.apache.org/licenses/LICENSE-2.0",
			},
			Version: "v1.0.0",
		},
	},
	GetDefinitions: api.GetOpenAPIDefinitions,
})
```

`GetOpenAPIDeinitions` is generated from Go types you defined. 
Add tag `+nirvana:openapi=true` to the `doc.go` file in package of api types just like the follow code

```
// +nirvana:openapi=true
package api
```

And run cmd to generate the function `GetOpenAPIDefinitions`.
If input(-i) packages are more than one, comma-separated list can be used.

```
openapi-gen \
  -i /go/package/to/your/types \
  -p /go/package/to/your/generated/function
```

You can output the documents in json format by json encoder(or yaml format by yaml encoder)

```
encoder := json.NewEncoder(os.Stdout)
if err := encoder.Encode(swagger); err != nil {
    panic(err)
}
```

You can also serve the documents in an OpenAPI endpoint, e.g. /v2/openapi.
`NOTICE: Don't add openapi descriptor into the descriptor passed to the builder.`

```
var openapi = definition.Descriptor{
	Path:        "/v2/openapi",
	Description: "OpenAPI endpoints",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: OpenAPI,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Results: []definition.Result{
				{
 					Destination: definition.Data,
					Description: "OpenAPI documents struct",
				},
				{
					Destination: definition.Error,
					Description: "Error to return if not success",
				},
			},
		},
	},
}

func OpenAPI() (*spec.Swagger, error) {
	swagger, err := builder.BuildOpenAPISpec(&yourDescriptor, &common.Config{
			...
		},
		GetDefinitions: api.GetOpenAPIDefinitions,
	}
	return swagger, err
}

```


### Configurer

Nirvana has a mechanism to set partial options into config. Here is an example mentioned above:

```go
config.Configure(nirvana.Descriptor(echo))
```

In the example, `nirvana.Descriptor` returns a configurer and the configurer will install descriptors into nirvana config.

There are some inside configurers in the table:

| Configurer  | Description                                |
| ----------- | ------------------------------------------ |
| IP          | Set listening ip. Defaults to "0.0.0.0"    |
| Port        | Set listening port. Defaults to 8080       |
| Logger      | Set custom logger                          |
| Descriptor  | Add API descriptors                        |
| Filter      | Add request filters                        |
| Modifier    | Add definition modifiers                   |

Plugins should also use configurers to configure plugins. For more details, see also [Plugins](#plugins)

### Error

In Nirvana core, error always means HTTP status code 500 - we try to avoid adding busniess logic into Nirvana.
That is, for error code other than 500, you are responsible to write your own error implementation, which only
needs to satisfy the following interface:

```go
// Error is a common interface for error.
// If an error implements the interface, type handlers can
// use Code() to get a specified HTTP status code.
type Error interface {
	// Code is a HTTP status code.
	Code() int
	// Message is an object which contains information of the error.
	Message() interface{}
}
```

An error contains status code and error message. Package `github.com/caicloud/nirvana/errors` provides standard
errors implementation and many helper functions. For example:

```go
// Example 1:
// Directly create an error.
// Fields (e.g. ${customer}) in format correspond to args (e.g. customer.Name) in order.
errors.NotFound.Error("${customer} not found", customer.Name)

// Example 2:
// Create an error factory at first.
var CustomerNotFount = errors.NotFound.Build("Project:Customer:CustomerNotFount", "${customer} not found")
// Then create error by factory.
CustomerNotFount.Error(customer.Name)
// You can check if an error is derived by specified factory.
if CustomerNotFount.Derived(err) {
	// Do something.
}
```

Use interface `errors.Error` in function signature is strongly discouraged. You should always use standard
`error` interface and create errors by the methods referred above.

### Logging

Nirvana provides a default logging implementation, the API mirrors [glog](https://github.com/golang/glog).
Following logging methods are provided with increasing severity.

```
Info
Warning
Error
Fatal
```

Keep in mind that:

* Each level comes with formatter and newliner method, i.e. `Infof` and `Infoln`
* `Info` has verbosity level, for example, you can use `log.V(4).Info` for unimportant logs
* `Fatal` error will terminate program execution

For more details, see `github.com/caicloud/nirvana/log` package.

### Plugins

#### Metrics

This plugin provides a lot of metrics with standard [prometheus](https://prometheus.io/) format. You can simply
enable it via:

```go
config.Configure(metrics.Default())
```

The plugin will register a middleware and a descriptor into your nirvana server, installing metrics at endpoint
`http://host:port/metrics`.

There are two config knobs in the plugin:
- Namespace: Metrics namespace is the prefix of all metric names. Defaults to `nirvana`.
- Path: Path is the descriptor path. Users can get metrics by the path. Defaults to `/metrics`

You can use following two configurers to change the settings:
- `metrics.Namespace(ns string)`: The function can modify metrics namespace.
- `metrics.Path(path string)`: The function can modity metrics descriptor path.

For more information about installed metrics, please check [Prometheus Doc](https://prometheus.io/docs/introduction/overview/).

#### Profiling

This plugin provides capability to install `pprof` into nirvana server, which is a direct reflection of golang
standard library `net/http/pprof`.

You can install the plugin via:

```go
config.Configure(profiling.Path("myprof"))
```

Then the plugin handles requests for the following paths:
- "/myprof": Show profiling index page.
- "/myprof/profile": Show cpu profile page.
- "/myprof/symbol": Show symbol page.
- "/myprof/trace": Show trace page.

The plugin has two configurers:
- `Path(path string)`: The function can change profiling descriptor path. Defaults to `/debug/pprof`
- `Contention(enable bool)`: Use to enable contention profiling. Defauts to `false`.

For more information about `pprof`, please check [PProf Doc](https://golang.org/pkg/net/http/pprof/).

#### Tracing

TBD

## Developer Guide and Proposals

### Proposals

- [kickoff](./docs/proposals/kickoff.md)
- [framework](./docs/proposals/framework.md)

### Plugin framework

Following is a framework for writing nirvana plugin. All aforementioned built-in plugins are written with the
framework: they are the best reference implementations if you ever want to draft a new plugin.

```go
func init() {
	// Register your config installer into nirvana.
	nirvana.RegisterConfigInstaller(&pluginInstaller{})
}

// ExternalConfigName is the external config name for your plugin. Please ensure that the
// name is unique and won't conflict with other plugins.
const ExternalConfigName = "pluginName"

type pluginInstaller struct{}

// Name is the external config name.
func (i *pluginInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder. You can get plugin config from nirvana config. Then
// install/initialize what you need.
func (i *pluginInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {...}

// Uninstall uninstalls stuffs after server terminating.
func (i *pluginInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {...)

// ConfigA configures fieldA. Be careful, you should get/save plugin config into nirvana config
// by `c.Config(ExternalConfigName)`/`c.Set(ExternalConfigName, cfg)` rather than a global
// plugin config.
func ConfigA(fieldA FieldType) nirvana.Configurer {...}

// ConfigB configures fieldB.
func ConfigB() nirvana.Configurer {...}

// Disable returns a configurer to disable current plugin for a certain nirvana server.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		// Set to nil will delete plugin config from nirvana config.
		c.Set(ExternalConfigName, nil)
		return nil
	}
}
```

Then user can use the plugin by:

```go
import "/path/to/plugin"

func main() {
	config := nirvana.NewDefaultConfig()
	config.Configure(plugin.ConfigA(fieldValue))
}
```
