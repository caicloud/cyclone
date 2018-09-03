# errors 包

errors 包类似于标准库的 errors 包，但是提供了方法用于生成格式化的错误和错误判断。


errors 包中存在三个概念，分别是 Builder，Factory 和 Error。其中 Builder 用于构建 Factory，Factory 则用于构建特定的 Error。

Factory 接口如下：
```go
// Factory can create error from a fixed format.
type Factory interface {
	// Error generates an error from v.
	Error(v ...interface{}) error
	// Derived checks if an error was derived from current factory.
	Derived(e error) bool
}
```

Error 接口如下（这个接口并没有显式定义在 errors 包中）：
```go
type Error interface {
	error
    Code() int
    Message() interface{}
}
```

首先看 Facotry 的两个方法：
1. Error 用于传入参数输出一个真正的错误。
2. Derived 则用于判断一个错误是否是由当前的 Factory 生成。
这样就能够非常方便的错误创建和错误判断了。

Factory 是一个具有特定的错误工厂，生成的错误都具有一样的形式（通常是指具有相同 string format，比如 "${user} is not found"）。

通常情况下，有 Factory 和 Error 就足够了。但是为了让错误能够以 HTTP API 的形式向客户端返回，我们还需要给 Factory 加上一些附加属性，用来表示返回的错误码等信息。
因此在 Factory 之上，构建了 Builder 接口，用于创建具有一类特征的 Factory（比如一类表示 NotFound 的错误）。

Builder 接口如下：
```go
// Builder can build error factories and errros.
type Builder interface {
	// Build builds a factory to generate errors with predefined format.
	Build(reason Reason, format string) Factory
	// Error immediately creates an error without reason.
	Error(format string, v ...interface{}) error
}
```

Builder 可以构建带有 reason 和 format 的 Factory。也直接提供了 Error 方法用于直接创建出 Error。

目前 errors 包提供的 Builder 主要是以 HTTP 状态码作为基础的：
```go
// These factory builders is used to build error factory.
var (
	BadRequest                   = newKind(400) // RFC 7231, 6.5.1
	Unauthorized                 = newKind(401) // RFC 7235, 3.1
	PaymentRequired              = newKind(402) // RFC 7231, 6.5.2
	Forbidden                    = newKind(403) // RFC 7231, 6.5.3
	NotFound                     = newKind(404) // RFC 7231, 6.5.4
	MethodNotAllowed             = newKind(405) // RFC 7231, 6.5.5
	NotAcceptable                = newKind(406) // RFC 7231, 6.5.6
	ProxyAuthRequired            = newKind(407) // RFC 7235, 3.2
	RequestTimeout               = newKind(408) // RFC 7231, 6.5.7
	Conflict                     = newKind(409) // RFC 7231, 6.5.8
	Gone                         = newKind(410) // RFC 7231, 6.5.9
	LengthRequired               = newKind(411) // RFC 7231, 6.5.10
	PreconditionFailed           = newKind(412) // RFC 7232, 4.2
	RequestEntityTooLarge        = newKind(413) // RFC 7231, 6.5.11
	RequestURITooLong            = newKind(414) // RFC 7231, 6.5.12
	UnsupportedMediaType         = newKind(415) // RFC 7231, 6.5.13
	RequestedRangeNotSatisfiable = newKind(416) // RFC 7233, 4.4
	ExpectationFailed            = newKind(417) // RFC 7231, 6.5.14
	Teapot                       = newKind(418) // RFC 7168, 2.3.3
	UnprocessableEntity          = newKind(422) // RFC 4918, 11.2
	Locked                       = newKind(423) // RFC 4918, 11.3
	FailedDependency             = newKind(424) // RFC 4918, 11.4
	UpgradeRequired              = newKind(426) // RFC 7231, 6.5.15
	PreconditionRequired         = newKind(428) // RFC 6585, 3
	TooManyRequests              = newKind(429) // RFC 6585, 4
	RequestHeaderFieldsTooLarge  = newKind(431) // RFC 6585, 5
	UnavailableForLegalReasons   = newKind(451) // RFC 7725, 3

	InternalServerError           = newKind(500) // RFC 7231, 6.6.1
	NotImplemented                = newKind(501) // RFC 7231, 6.6.2
	BadGateway                    = newKind(502) // RFC 7231, 6.6.3
	ServiceUnavailable            = newKind(503) // RFC 7231, 6.6.4
	GatewayTimeout                = newKind(504) // RFC 7231, 6.6.5
	HTTPVersionNotSupported       = newKind(505) // RFC 7231, 6.6.6
	VariantAlsoNegotiates         = newKind(506) // RFC 2295, 8.1
	InsufficientStorage           = newKind(507) // RFC 4918, 11.5
	LoopDetected                  = newKind(508) // RFC 5842, 7.2
	NotExtended                   = newKind(510) // RFC 2774, 7
	NetworkAuthenticationRequired = newKind(511) // RFC 6585, 6
)
```
这个包方便了用户创建能够被 Nirvana 识别的错误，但是如果业务逻辑中如果不希望引入对 errors 包的依赖，可以自行实现错误包，只要产出的错误符合 Error 接口即可。
