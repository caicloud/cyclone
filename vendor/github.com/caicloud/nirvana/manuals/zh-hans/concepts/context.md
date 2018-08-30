# Context

包路径: `github.com/caicloud/nirvana/service`

在 Nirvana 中，Context 用于传递请求的上下文。Context 中包含 HTTP 的 Request 和 ResponseWriter。可是使用 service 包的 `HTTPContextFrom()` 方法获得 HTTP Context。HTTP Context 相关接口如下：
```go
// ValueContainer contains values from a request.
type ValueContainer interface {
	// Path returns path value by key.
	Path(key string) (string, bool)
	// Query returns value from query string.
	Query(key string) ([]string, bool)
	// Header returns value by header key.
	Header(key string) ([]string, bool)
	// Form returns value from request. It is valid when
	// http "Content-Type" is "application/x-www-form-urlencoded"
	// or "multipart/form-data".
	Form(key string) ([]string, bool)
	// File returns a file reader when "Content-Type" is "multipart/form-data".
	File(key string) (multipart.File, bool)
	// Body returns a reader to read data from request body.
	// The reader only can read once.
	Body() (reader io.ReadCloser, contentType string, ok bool)
}

// ResponseWriter extends http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	// HeaderWritable can check whether WriteHeader() has
	// been called. If the method returns false, you should
	// not recall WriteHeader().
	HeaderWritable() bool
	// StatusCode returns status code.
	StatusCode() int
	// ContentLength returns the length of written content.
	ContentLength() int
}

// HTTPContext describes an http context.
type HTTPContext interface {
	Request() *http.Request
	ResponseWriter() ResponseWriter
	ValueContainer() ValueContainer
	RoutePath() string
}
```

Nirvana 框架会为每个请求构建这样的 HTTPContext。如有必要，可以通过这些接口拿到与请求相关的所有数据。

在一个请求路由匹配成功后，Nirvana 会把对应的 HTTPContext 传递给中间件，然后由中间件调用链继续传递。最终经由 ContextPrefab 传递给业务函数。

**中间件不应该修改 HTTPContext，除非您明确知道如何修改。**
