# 中间件

包路径: `github.com/caicloud/nirvana/definition`

Nirvana 中间件位于路由之后，业务函数执行之前。因此中间件可以对合法的请求进行一些额外的处理。中间件的接口如下： 

```go
// Chain contains all subsequent actions.
type Chain interface {
	// Continue continues to execute the next subsequent actions.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call Chain.Continue() and pass the context.
type Middleware func(context.Context, Chain) error
 
```

一个请求在路由匹配成功后，就会进入执行过程。执行时会先执行匹配到的所有中间件，并由中间件通过 Chain 进行链式调用。也就是说任何一个中间件都可以决定请求是否继续执行。

中间件添加在 Descriptor 中：
```go
def.Descriptor{
	Path:        "/path",
	Middlewares: []def.Middleware{SomeMiddleware},
}
```
添加成功后，所有前缀匹配 `/path` 的请求都会执行这个中间件。


**注意：只有在路由找到了请求的业务函数的情况下，才算匹配成功。没有匹配成功的情况下，中间件不会执行。**


## 中间件执行顺序

如果以下路径都添加了中间件：
```go
1. /
2. /path
3. /path/path2
```
那么如果存在请求 `/path/path2/others` 且成功匹配的情况下，中间件的执行按照 1 -> 2 -> 3 的顺序。

中间件只与 URL Path 有关，因此如果多个 Descriptor 中为相同的路径添加了多个中间件，在路径匹配的情况下，这些中间件都会被执行，但是执行顺序不能确定。因此开发中间件时不要依赖中间件的执行顺序。


