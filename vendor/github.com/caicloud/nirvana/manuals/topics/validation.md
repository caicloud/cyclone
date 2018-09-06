# Validation

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
2. 10 characters validation is not enough to prevent spam :) (checkout guide below to add your own validation)

For full example code, see [validator](https://github.com/caicloud/nirvana/tree/master/examples/getting-started/validator).
