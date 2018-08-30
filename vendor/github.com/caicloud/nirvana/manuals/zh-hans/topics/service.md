# service 包

service 包实现了 Nirvana 的 API 处理框架：
```
           Service.ServeHTTP()
         ----------------------
         ↓                    ↑
|-----Filters------|          ↑
         ↓                    ↑
|---Router Match---|          ↑
         ↓                    ↑
|-------------Middlewares------------|
         ↓                    ↑
|-------------Executor---------------|
         ↓                    ↑
|-ParameterGenerators-|-DestinationHandlers-|
         ↓                    ↑
|------------User Function-----------|
```

service 包的入口是 Builder：
```go
// Builder builds service.
type Builder interface {
	// Logger returns logger of builder.
	Logger() log.Logger
	// SetLogger sets logger to server.
	SetLogger(logger log.Logger)
	// Modifier returns modifier of builder.
	Modifier() DefinitionModifier
	// SetModifier sets definition modifier.
	SetModifier(m DefinitionModifier)
	// Filters returns all request filters.
	Filters() []Filter
	// AddFilters add filters to filter requests.
	AddFilter(filters ...Filter)
	// AddDescriptors adds descriptors to router.
	AddDescriptor(descriptors ...definition.Descriptor) error
	// Middlewares returns all router middlewares.
	Middlewares() map[string][]definition.Middleware
	// Definitions returns all definitions. If a modifier exists, it will be executed.
	Definitions() map[string][]definition.Definition
	// Build builds a service to handle request.
	Build() (Service, error)
}

type Service interface {
	http.Handler
}

// DefinitionModifier is used in Server. It's used to modify definition.
// If you want to add some common data into all definitions, you can write
// a customized modifier for it.
type DefinitionModifier func(d *definition.Definition)
 

// Filter can filter request. It has the highest priority in a request
// lifecycle. It runs before router matching.
// If a filter return false, that means the request should be filtered.
// If a filter want to filter a request, it should handle the request
// by itself.
type Filter func(resp http.ResponseWriter, req *http.Request) bool
 
```
Builder 构建 Service 来提供 HTTP 服务。因此 Builder 提供了多个方法用于设置生成服务需要的日志，Definition 修改器，请求过滤器，API 描述符。构建完成的 Service 实际上是一个 http.Handler，用来处理请求。

其中 Definition 修改器用于在生成路由之前修改 API Definition。请求过滤器则是在 Service 执行的时候才会被调用，请求过滤器的优先级高于路由匹配。也就是说，在路由匹配之前，请求就有可能被过滤器直接过滤掉。

Builder 还会将 API Definition 转换为路由需要的数据结构，涉及到以下内容：
1. 对应 Consumes 和 Produces 的 Consumer 和 Producer  
  Consumer 针对请求的 body，将数据转换为业务函数需要的数据类型（通常是结构体）。  
  Producer 则是将业务函数的返回值转换并写入到响应的 body 中。  
  ```go
  // Consumer handles specifically typed data from a reader and unmarshals it into an object.
  type Consumer interface {
  	// ContentType returns a HTTP MIME type.
  	ContentType() string
  	// Consume unmarshals data from r into v.
  	Consume(r io.Reader, v interface{}) error
  }
  
  // Producer marshals an object to specifically typed data and write it into a writer.
  type Producer interface {
  	// ContentType returns a HTTP MIME type.
  	ContentType() string
  	// Produce marshals v to data and write to w.
  	Produce(w io.Writer, v interface{}) error
  }
  ```
1. 对应 Prefab 类型的 Prefab 生成器  
  这个生成器用于创建业务函数需要的特定实例，一般是服务端实例，即不是从请求里获取的数据而生成的。  
  service 包里提供了一个 Context Prefab 生成器，简单的将参数里的 context 返回出去，供业务函数使用。
  ```go
  // Prefab creates instances for internal type. These instances are not
  // unmarshaled form http request data.
  type Prefab interface {
  	// Name returns prefab name.
  	Name() string
  	// Type is instance type.
  	Type() reflect.Type
  	// Make makes an instance.
  	Make(ctx context.Context) (interface{}, error)
  }
  ```
1. 对应 golang 基础类型的转换器  
  这些转换器一般是用于将请求里的 query，header 等简单字符串数据转换为 golang 的基础类型，供业务函数使用。
  ```go
  // Converter is used to convert []string to specific type. Data must have one
  // element at least or it will panic.
  type Converter func(ctx context.Context, data []string) (interface{}, error)
  ```
1. 用于封装请求的 ValueContainer  
  这个接口是对 Request 的一次封装，方便获取对应位置的字符串数据。
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
  ```
1. 用于封装响应的 ResponseWriter  
  ResponseWriter 是对 http.ResponseWriter 的一个扩展，提供了一些功能方便中间件使用。
  ```go
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
  ```
1. 用于合并请求和响应的 Context  
  HTTPContext 实现了 Context 接口，包装了请求的信息。作为路由上下文使用。
  ```go
  // HTTPContext describes an http context.
  type HTTPContext interface {
  	Request() *http.Request
  	ResponseWriter() ResponseWriter
  	ValueContainer() ValueContainer
  	RoutePath() string
  }
  ```
1. 用于生成业务函数的参数的 ParameterGenerator  
  ParameterGenerator 是真正的参数生成器，通过调用 Consumer，Converter，Prefab 等来完成业务函数的参数生成。
  ```go
  // ParameterGenerator is used to generate object for a parameter.
  type ParameterGenerator interface {
  	// Source returns the source generated by current generator.
  	Source() definition.Source
  	// Validate validates whether defaultValue and target type is valid.
  	Validate(name string, defaultValue interface{}, target reflect.Type) error
  	// Generate generates an object by data from value container.
  	Generate(ctx context.Context, vc ValueContainer, consumers []Consumer, name string, target reflect.Type) (interface{}, error)
  }
  ```
1. 用于将业务函数返回值写入 Response 的 DestinationHandler  
  DestinationHandler 是业务函数返回值处理器，通过调用 Producer 将返回值转换为字节写入响应中。  
  在 DestinationHandler 中，错误是会进行特殊处理的。如果业务函数返回的错误符合 Error 接口，则会根据这个接口来生成错误码和返回数据结构。
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
  
  const (
  	// HighPriority for error type.
  	// If an error occurs, ignore meta and data.
  	HighPriority int = 100
  	// MediumPriority for meta type.
  	MediumPriority int = 200
  	// LowPriority for data type.
  	LowPriority int = 300
  )
  
  // DestinationHandler is used to handle the results from API handlers.
  type DestinationHandler interface {
  	// Type returns definition.Type which the type handler can handle.
  	Destination() definition.Destination
  	// Priority returns priority of the type handler. Type handler with higher priority will prior execute.
  	Priority() int
  	// Validate validates whether the type handler can handle the target type.
  	Validate(target reflect.Type) error
  	// Handle handles a value. If the handler has something wrong, it should return an error.
  	// The handler descides how to deal with value by producers and status code.
  	// The status code is a success status code. If everything is ok, the handler should use the status code.
  	//
  	// There are three cases for return values (goon means go on or continue):
  	// 1. go on is true, err is nil.
  	//    It means that current type handler did nothing (or looks like did nothing) and next type handler
  	//    should take the context.
  	// 2. go on is false, err is nil.
  	//    It means that current type handler has finished the context and next type handler should not run.
  	// 3. err is not nil
  	//    It means that current type handler handled the context but something wrong. All subsequent type
  	//    handlers should not run.
  	Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error)
  }
  ```
 
**注：以上每个接口对应的实例都是可以通过相关的函数注册和修改的。**
  
