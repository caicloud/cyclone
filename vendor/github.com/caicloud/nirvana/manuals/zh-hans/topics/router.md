# router 包

router 包实现了基于前缀树的路由，并提供了三种类型的路由节点：
1. 字符串类型节点
1. 正则类型节点（包括一个全匹配类型节点）
1. 剩余路径类型节点

字符串类型节点：
```go
// stringNode describes a string router node.
type stringNode struct {
	handler
	children
	// prefix is the fixed string to match path.
	prefix string
}
```
字符串类型节点进行字符串匹配。在匹配过程中，路径前缀必须与 prefix 完全匹配。

正则类型节点：
```go
// index contains the key and it's index of the submatches.
type index struct {
	// Key is the name for the value.
	Key string
	// Pos is the index of value in submatches.
	Pos int
}

// regexpNode contains information for matching a regexp segment.
type regexpNode struct {
	handler
	children
	// indices contains all positions to get values from submatches.
	indices []index
	// exp is the regular expression.
	exp string
	// regexp is a regexp instance to match.
	regexp *regexp.Regexp
}
```
正则类型节点则根据正则表达式匹配一段路径，并提取出匹配的字符串。

全匹配类型节点是正则类型节点的一个特例，即等价于正则表达式的 `.*`：
```go
// fullMatchRegexpNode is an optimizing of RegexpNode.
type fullMatchRegexpNode struct {
	handler
	children
	// key is the name for the only value.
	key string
}
```
这是对正则表达式的一个常用特例的优化，快速进行全匹配。

正则类型节点和全匹配类型节点存在一个特殊的限制：只能匹配到下一个 / 之前。

剩余路径类型节点：
```go
// pathNode matches all rest path.
type pathNode struct {
	handler
	// key is the key for the rest path.
	key string
}
```
这个节点匹配剩余所有路径。

下面用一个例子来说明这几个节点：
```go
API Path: /apis/v1/{regexp:[a-z]{1,2}}/{fullmatch}/{path:*}
Matched Path: /apis/v1/ab/something/the/rest/path
Generated Nodes:
  String Node: /apis/v1/           ->    /apis/v1/
  Regexp Node: regexp:[a-z]{1,2}   ->    ab
  String Node: /                   ->    /
  Full Match Node: fullmatch:.*    ->    something
  String Node: /                   ->    /
  Path Node: *                     ->    the/rest/path
```

一个路径在前缀树的节点中从根节点开始进行匹配，直到某个后代节点匹配完了整个路径，那么这个后代节点就会作为这个路径的匹配链的最后一个节点，即使这个节点还有后代节点，也不会继续进行匹配。因此每个节点都可能是某些路径的最后一个节点。

在上面的路径节点中，每个节点都可以绑定一个 Inspector：
```go
// Inspector can select an executor to execute.
type Inspector interface {
	// Inspect finds a valid executor to execute target context.
	// It returns an error if it can't find a valid executor.
	Inspect(context.Context) (Executor, error)
}

// Executor executs with a context.
type Executor interface {
	// Execute executes with context.
	Execute(context.Context) error
}
```
路径匹配完成后，会调用匹配链的最后一个节点的 Inspector 来生成一个能够处理当前的路由上下文的 Executor。Inspector 如果能返回一个 Executor，router 会将整个匹配链上所有节点按照从根节点开始的顺序将中间件串联起来，并绑定上这个 Executor。如果不能返回 Executor，则认为路由匹配失败。匹配失败的话，中间件都不会执行。

中间件接口如下：
```go
// RoutingChain contains the call chain of middlewares and executor.
type RoutingChain interface {
	// Continue continues to execute the next middleware or executor.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call RoutingChain.Continue() and pass the context.
type Middleware func(context.Context, RoutingChain) error
```
对于中间件而言，处理完当前的任务之后只需要调用 RoutingChain 将 Context 通过 Continue 传递下去即可。这是一个阻塞过程，只有后续的中间件执行完成了才会返回。而最后一个中间件调用 Continue 实际上是调用的 Executor，因此所有中间件的 Continue 执行完成之后，请求也处理完成了。

**注：这个包里所有的接口都不会被用户直接使用，用户只能通过 definition 包进行 API 定义，然后由 service 包进行路由构建和匹配。**









