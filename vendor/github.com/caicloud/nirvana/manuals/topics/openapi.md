# OpenAPI

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
// +caicloud:openapi=true
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
