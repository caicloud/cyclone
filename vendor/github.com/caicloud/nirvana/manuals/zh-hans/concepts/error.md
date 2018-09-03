# 错误包

包路径: `github.com/caicloud/nirvana/errors`

## Error 介绍

在业务函数中，除了正常的返回数据以外，还需要处理各种错误。在 golang 中，错误需要实现 error 接口。即：
```go
type error interface {
	Error() string
}
```
这种形式的 error 接口只能返回一个特定的字符串，而不能携带更多的关于错误的信息。因此 Nirvana 对于这种错误，都会以 500 Internal Server Error 的形式返回给客户端。但是通常情况下，业务为了区分不同的错误，会使用错误码和错误字符串来共同标识一个错误，方便客户端识别。

于是 Nirvana 提供了一个新的接口：
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

在业务函数中，仍然以 error 的形式返回错误。但是框架会检查返回的错误是否实现了 Error 接口。如果实现了，就会以 `Code()` 的返回值作为 HTTP 状态码，`Message()` 的返回值作为数据返回。

### errors 包

为了方便使用，Nirvana 提供了 `errors` 包用于生成 error。创建 error 的方式有两种：

1. 方法一

```go
var somethingNotCorrect = errors.BadRequest.Build("ProjectName:ModuleName:SomethingNotCorrect", "${name} is not correct")

func SomeFunction() error {
	// Do something
	return somethingNotCorrect.Error(something.Name)
}
```

2. 方法二

```go
func SomeFunction() error {
	return errors.BadRequest.Error("${name} is not correct", something.Name)
}  
```

这两种方法都可以创建 error，但是第一种方法比第二种多出更多特性：

- 第一种方法支持使用 `somethingNotCorrect.Derived(err)` 的形式判断一个 err 是否由这个错误工厂生成。
- 第一种方法带有 Reason，方便客户端判断错误类型。

因此我们建议始终使用第一种方法进行错误的创建。只有在错误仅在内部使用，而且不需要进行错误判断的情况下，为了简化错误创建的过程，可以使用第二种方法。

### error 的 Reason

在实际的业务中，HTTP 状态码并不足以表达业务中繁复的错误。因此我们将 HTTP 状态码视为错误大类（比如 NotFound 表示资源不存在或者找不到），然后在大类下定义业务需要的具体错误。

这样做有两个优势：
1. 通过 HTTP 状态码即可大致判断一个错误的行为
2. 通过具体错误的 Reason 来唯一确定一个错误

在常见的商业 API 中，我们也经常看到使用数字 ID 来标志的错误。但是数字 ID 的表达性很差，而且全局唯一性比较难维护。因此我们使用字符串作为错误的 ID，也就是 Reason。

我们建议 Reason 的格式满足：

`项目名[:模块名]:错误名`

### 国际化

使用 errors 包生成的错误会记录每个占位符的名称和值，保存在 data 字段中，可以在客户端使用 data 里的值进行文本国际化。

## 使用范例

在业务函数中使用 errors：
```go

var messageNotExist = errors.NotFound.Build("MyProject:Message:MessageNotExist", "there is no message with id ${id}")

// GetMessage return a message by id.
func GetMessage(ctx context.Context, id int) (*Message, error) {
	if id > 100 {
		return nil, messageNotExist.Error(id)
	}
	return &Message{
		ID:      id,
		Title:   "This is an example",
		Content: "Example content",
	}, nil
}
```
编译运行后可以得到结果：

访问 `curl http://localhost:8080/apis/v1/messages/100`，即可得到一个 200 的响应，响应体为：
```
{"id":100,"title":"This is an example","content":"Example content"}
```

访问 `curl http://localhost:8080/apis/v1/messages/101`，即可得到一个 404 的响应，响应体为：
```
{"reason":"MyProject:Message:MessageNotExist","message":"there is no message with id 101","data":{"id":"101"}}
```

