# Routes

## Definitions

Defining routes and be as easy as defining such functions:

```go
func handler(request Request) (Response, error) {
  // do stuff
  // return data
  // or error if anything wrong happened
}
```

However we can do a lot more with Nirvana framework!

### Declarative approach to routes definition

In Nirvana, routes definitions _are_ documentations, i.e. you will only need to follow the _DRY_ rule, define your routes (w.r.t specific formats) and then we'll try to cover these for you:

* basic test-case generation
* OpenAPI docs generation
* parameter checking (against function signatures)
* etc.

### `definition.Descriptor` struct

This is the core data structure that you'll use, and a typical example looks like this:

```go
definition.Descriptor{
  Path:        "/api",
  Description: "here goes your api description",
  Produces:    []string{"text/plain"},
  Definitions: []definition.Definition{
    {
      Method: definition.Get,
      Function: func(ctx context.Context) (string, error) {
        return "Hello World!", nil
      },
      Parameters: []definition.Parameter{},
      Results: []definition.Result{
        {
          Type: definition.Data,
        },
        {
          Type: definition.Error,
        },
      },
    },
  },
}
```

TODO: add more
