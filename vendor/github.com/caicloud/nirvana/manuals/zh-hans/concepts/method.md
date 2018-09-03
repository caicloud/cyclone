# 方法包

包路径: `github.com/caicloud/nirvana/definition/method`

在 Nirvana 中，业务函数是 API 请求的 Handler，而且表现形式是函数或者同等形式的闭包。但是很多时候我们的业务函数可能是某个实例的方法。

为了保证业务与 API 定义的无关性，Nirvana 提供了 method 包，用于分离实例的创建和 API 定义。

API 定义：
```go
var listMessages = def.Definition{
	Method:      def.List,
	Summary:     "List Messages",
	Description: "Query a specified number of messages and returns an array",
	Function:    method.Get(&message.Container{}, "ListMessages"),
	Parameters: []def.Parameter{
		{
			Source:      def.Query,
			Name:        "count",
			Default:     10,
			Description: "Number of messages",
		},
	},
	Results: def.DataErrorResults("A list of messages"),
}
```

业务方法：
```go
// Container contains example title and content of messages.
type Container struct {
	Title   string
	Content string
}

// NewContainer creates Container
func NewContainer(title, content string) *Container {
	return &Container{title, content}
}

// ListMessages returns all messages.
func (m *Container) ListMessages(ctx context.Context, count int) ([]Message, error) {
	messages := make([]Message, count)
	for i := 0; i < count; i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("%s %d", m.Title, i)
		messages[i].Content = fmt.Sprintf("%s %d", m.Content, i)
	}
	return messages, nil
}
```

main.go 中创建实例并通过 `Put()` 函数放到方法容器中（需要在服务启动之前完成）：
```go
method.Put(message.NewContainer("Method Example", "Method Content"))
```

然后编译运行，访问 `http://localhost:8080/apis/v1/messages` 即可得到返回结果：
```
[{"id":0,"title":"Method Example 0","content":"Method Content 0"},{"id":1,"title":"Method Example 1","content":"Method Content 1"},{"id":2,"title":"Method Example 2","content":"Method Content 2"},{"id":3,"title":"Method Example 3","content":"Method Content 3"},{"id":4,"title":"Method Example 4","content":"Method Content 4"},{"id":5,"title":"Method Example 5","content":"Method Content 5"},{"id":6,"title":"Method Example 6","content":"Method Content 6"},{"id":7,"title":"Method Example 7","content":"Method Content 7"},{"id":8,"title":"Method Example 8","content":"Method Content 8"},{"id":9,"title":"Method Example 9","content":"Method Content 9"}]
```

## method 包介绍

method 包是一个全局实例容器，每种类型对应一个实例。

### 具体实例类型

使用具体实例类型是一种比较常见的情况，比如上面的例子中提到的 `*message.Container` 实例。

其中 `Put(ins insterface{})` 函数用于将一个实例放置到全局容器中，形成 类型-实例 的对应关系。

`Get(typIns interface{}, method string)` 用于生成一个匿名函数，生成的函数的签名没有方法的 receiver 部分，例如：
```
func (m *Container) ListMessages(ctx context.Context, count int) ([]Message, error)
生成匿名函数：
func (ctx context.Context, count int) ([]Message, error)
```
但是匿名函数的执行部分在创建时是不确定的，只有等到这个匿名函数第一次被调用的时候，才会去全局实例容器里找到对应类型的实例，然后调用指定的方法。

这就是 `Get()` 函数的实例延迟加载特性。一旦延迟加载完成后，之后即使全局实例容器里的实例发生改变，匿名函数仍然会调用旧的实例的方法。如果加载时对应的类型没有实例，则会 panic。

注意，`Get()` 函数的第一个参数仅仅用于获取类型，其实例值没有任何意义。因此不需要给这个实例的任何字段设置值。也就是说 `&Container{}` 和 `(*Container)(nil)` 是一样的。

### 接口实例类型

除了直接使用具体的类型以外，还支持使用接口类型。和具体类型的区别只是需要明确的指定接口类型的指针。

设置接口实例：
```go
method.PutInterface((*ArbitraryInterface)(nil), instance)
```

获取接口实例：
```go
method.Get((*ArbitraryInterface)(nil), "MethodName")
```

