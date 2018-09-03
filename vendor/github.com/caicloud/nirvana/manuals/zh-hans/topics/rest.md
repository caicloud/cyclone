# rest 包

rest 包提供了一个简单的 REST Client 用于访问 API 服务。

在这个包中，主要包含三个概念：Config，Client 和 Request。Config 是 Client 的配置，用于创建 Client。而 Request 则由 Client 创建，用来表示每一个 REST 请求。
```go
// RequestExecutor implements a http client.
type RequestExecutor interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config is rest client config.
type Config struct {
	// Scheme is http scheme. It can be "http" or "https".
	Scheme string
	// Host must be a host string, a host:port or a URL to a server.
	Host string
	// Executor is used to execute http requests.
	// If it is empty, http.DefaultClient is used.
	Executor RequestExecutor
}

// Client implements builder pattern for http client.
type Client struct {
	...
}

// NewClient creates a client.
func NewClient(cfg *Config) (*Client, error)

// Request creates an request with specific method and url path.
// The code is only for checking if status code of response is right.
func (c *Client) Request(method string, code int, url string) *Request

// Request describes a http request.
type Request struct {
	...
}

// Path sets path parameter.
func (r *Request) Path(name string, value interface{}) *Request

// Query sets query parameter.
func (r *Request) Query(name string, values ...interface{}) *Request

// Header sets header parameter.
func (r *Request) Header(name string, values ...interface{}) *Request

// Form sets form parameter.
func (r *Request) Form(name string, values ...interface{}) *Request

// File sets file parameter.
func (r *Request) File(name string, file interface{}) *Request

// Body sets body parameter.
func (r *Request) Body(contentType string, value interface{}) *Request

// Meta sets header result.
func (r *Request) Meta(value *map[string]string) *Request

// Data sets body result. value must be a pointer.
func (r *Request) Data(value interface{}) *Request

// Do executes the request.
func (r *Request) Do(ctx context.Context) error
```
Request 保存了一个请求的数据，用 Path，Query，Header，Form，File，Body 来设置请求的相关值，Meta 和 Data 来设置用于接收响应的值（都是指针）。然后 Do 用于真正发起请求，并完成 Meta 和 Data 的填充。

Request 的方法与 API Definition 的除了 Prefab 和 Error 之外的参数和返回值类型一一对应，这样可以十分方便的设置对应的值。由于 Prefab 是由服务端生成而不由客户端提交，Error 由 Do 方法返回，因此这两种类型没有对应的方法。

**注：这个 Client 会被由命令 nirvana client 生成的客户端依赖，因此需要确保两者的一致性。如果用户自定义了一些新的请求和响应类型，也需要对这个客户端进行扩展。** 
