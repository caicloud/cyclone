# Prefab

包路径: `github.com/caicloud/nirvana/service`

Prefab 是 Nirvana 中一类特殊的 Source。其他的 Source 的数据来源都是来自于请求，而 Prefab 来自于服务端本身。

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

可以通过 service 包的 `RegisterPrefab()` 注册需要的 Prefab。

**Prefab 类型的参数在生成文档和客户端的时候会被忽略，因此不要使用 Prefab 从请求中获取数据。**

## Nirvana 提供的 Prefab

### ContextPrefab

ContextPrefab 是 Nirvana 中实现的唯一一个 Prefab，即 `service.ContextPrefab`。这个 Prefab 将框架传递给它的与请求绑定的 context 返回回去。

使用方法如下：
```go
var someAPI = def.Definition{
	...
	Parameters: []def.Parameter{
		...
		{
			Source:  def.Prefab,
			Name:    "context",
		},
		...
	},
	...
}
```
只需要将业务函数对应位置的 Parameter 设置为 Prefab，名称为 `context` 即可。

但是一般情况下，我们不应该这样使用 `ContextPrefab`。请参考 [Modifier](modifier.md) 和 [Context](context.md)
