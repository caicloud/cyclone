# Destination Handler

包路径: `github.com/caicloud/nirvana/service`

Nirvana 默认提供了 3 种类型的 Destination：Meta，Data，Error。

每种 Destination 对应一个 Handler。这些 Handler 负责一种类型的返回结果的数据转换工作。

```go
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

如果 Nirvana 默认提供的 Handler 不能满足实际的业务需求，可以通过 service 包提供的方法注册自定义的 Handler：
```go
// RegisterDestinationHandler registers a type handler.
func RegisterDestinationHandler(handler DestinationHandler) error
```

Definition Handler 存在优先级，优先级高的 Handler 先执行。并且执行之后会返回 `goon`，用来确定是否需要执行下一个 Handler。
 

