# Nirvana

Nirvana is a Golang web framework with a focus on developer efficiency and
performance. It handles request routing, input validation, logging, error
handling, etc. with sensible defaults and options for complete control. Built
for developers with ❤️ by developers in [Caicloud](http://caicloud.io/).

## Status

Currently the framework is 🚧 WIP 🚧, any comments or contributions are welcome!

## Getting Started

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
	config := nirvana.NewDefaultConfig("", 8080)
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

```sh
$ curl "http://localhost:8080/echo?msg=test"
test
```

For full example code, see [basics](https://github.com/caicloud/nirvana/tree/master/examples/getting-started/basics).
