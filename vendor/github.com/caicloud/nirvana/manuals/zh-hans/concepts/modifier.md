# Modifier

包路径: `github.com/caicloud/nirvana/service`

在 Nirvana 中，每一个 API 都有一个对应的 Definition。在我们实际开发过程中，经常会要求 API 和 业务函数有一致的行为表现。比如每个业务函数的第一个参数都是 Context。在这种场景下，如果每个 Definition 都需要去描述这个参数，那么 Definition 会显得非常冗余。因此 Nirvana 提供了 Definition Modifer 机制，允许在 Definition 生效之前，对 Definition 进行修改。

这样就能通过 Modifier 完成 Definition 公共部分的构建，而每个 Definition 实际上要填写的部分就是只与自身业务相关的信息。Modifier 如下：
```go
type DefinitionModifier func(d *definition.Definition)
```

## Nirvana 提供的 Modifiers

在使用 `nirvana init` 创建的项目中，可以在 `pkg/apis/modifiers` 下查看启用的 Modifiers。 

默认启用的 Modifiers 包括：FirstContextParameter，ConsumeAllIfConsumesIsEmpty，ProduceAllIfProducesIsEmpty，ConsumeNoneForHTTPGet，ConsumeNoneForHTTPDelete，ProduceNoneForHTTPDelete。

### FirstContextParameter

这个 Modifier 为所有 Definition 的第一个参数添加上名为 `context` 的 Prefab。启用之后，所有业务函数的第一个参数必须是 `context.Context`。

### ConsumeAllIfConsumesIsEmpty

这个 Modifier 为所有 Consumes 为空的 Definition 加上 `*/*`。

### ProduceAllIfProducesIsEmpty

这个 Modifier 为所有 Produces 为空的 Definition 加上 `*/*`。

### ConsumeNoneForHTTPGet

这个 Modifier 为所有 HTTP GET 类型的 Definition 加上空的 Consumer，即允许请求体为空。

### ProduceNoneForHTTPDelete

这个 Modifier 为所有 HTTP Delete 类型的 Definition 加上空的 Producer，即允许响应体为空。

### LastErrorResult

这个 Modifier 为所有的 Definition 的最后一个返回值加上 `definition.Error`。


