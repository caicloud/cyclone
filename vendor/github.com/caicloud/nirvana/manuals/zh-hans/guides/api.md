# API

## 概念

### Nirvana Definition
在 Nirvana 中，所有的 API 都是通过 Descriptor 和 Definition 进行描述的。首先看一个 `List Messages` 的 API 定义：
```go
// 在使用 nirvana init 创建的标准项目结构中，这个文件位于 pkg/apis/v1/descriptors/message.go

func init() {
	register([]def.Descriptor{{
		// Path 定义了 API 路径
		Path:        "/messages",
		// Definitions 数组包含了这个路径下的所有定义。
		Definitions: []def.Definition{listMessages},
	},
	}...)
}

// listMessages 定义了一个返回 Message 列表的 API
var listMessages = def.Definition{
	// 这个 API 返回的是资源数组，所以使用 List 方法。
	Method:      def.List,
	// Summary 是一个短语，用于描述这个 API 的用途。这个短语在生成文档和客户端的时候用于区分 API。
	// 这个字符串去掉空格后会作为生成客户端时的函数名，因此请确保这个字符串是有意义的。
	Summary:     "List Messages",
	// 详细描述这个 API 的用途。
	Description: "Query a specified number of messages and returns an array",
	// 业务函数
	Function:    message.ListMessages,
	// 对应业务函数的参数信息。用于告知 Nirvana 从请求的那一部分取得数据，然后传递给业务函数。
	Parameters: []def.Parameter{
		{
			// 参数来源
			Source:      def.Query,
			// 参数名称，作为 key 从 Source 里取值。
			// 与业务函数的参数名称无关。
			Name:        "count",
			// 默认值
			Default:     10,
			// 参数描述 
			Description: "Number of messages",
		},
	},
	// 对应业务函数的返回结果。用于告知 Nirvana 业务函数返回结果如何放到请求的响应中。
	Results: def.DataErrorResults("A list of messages"),
}
```
根据上面的 API 定义，再对应业务函数：
```go
// 在使用 nirvana init 创建的标准项目结构中，这个文件位于 pkg/message/message.go

// Message describes a message entry.
type Message struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ListMessages returns all messages.
func ListMessages(ctx context.Context, count int) ([]Message, error) {
	messages := make([]Message, count)
	for i := 0; i < count; i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("Example %d", i)
		messages[i].Content = fmt.Sprintf("Content of example %d", i)
	}
	return messages, nil
}
```
可以看到，业务函数既不关心参数的来源和类型转换，也不关心如何将返回值写到响应里，只是按照业务需求实现逻辑。

### Definition Method
在 Nirvana 中，我们建议所有的 API 都遵守 RESTful 风格，并且在 URL 中携带 API 的版本号。下表中展示了 Nirvana 中定义的动作以及对应的 API 定义。

| Nirvana 方法 | HTTP 方法 | HTTP 成功状态码 | URL                           | 描述                          |
|--------------|-----------|-----------------|-------------------------------|-------------------------------|
| List         | GET       | 200             | /apis/v1/resources            | 获取资源列表                  |
| Get          | GET       | 200             | /apis/v1/resources/{resource} | 根据资源唯一 ID/Name 获取资源 |
| Create       | POST      | 201             | /apis/v1/resources            | 创建一个资源（非幂等）        |
| Update       | PUT       | 200             | /apis/v1/resources/{resource} | 更新一个资源（幂等）          |
| Patch        | PATCH     | 200             | /apis/v1/resources/{resource} | 修改一个资源的部分内容        |
| Delete       | DELETE    | 204             | /apis/v1/resources/{resource} | 删除一个资源                  |
| AsyncCreate  | POST      | 202             | /apis/v1/resources            | 异步创建资源                  |
| AsyncUpdate  | PUT       | 202             | /apis/v1/resources/{resource} | 异步更新资源                  |
| AsyncPatch   | PATCH     | 202             | /apis/v1/resources/{resource} | 异步修改资源部分内容          |
| AsyncDelete  | DELETE    | 202             | /apis/v1/resources/{resource} | 异步删除资源                  |

对于 Nirvana 异步方法，发出后服务端应当只是将请求加入执行队列，然后立刻返回一个关联的对象或者链接，供客户端后续查询请求执行状态。

所有的 Nirvana 方法都是语义层面的，为的是提高 API 定义的可读性。也就是说 List 和 Get 在一个 HTTP 请求中使用的都是 GET，两者没有区别。
但是为了使 API 定义更加明确，我们应该根据场景确定使用哪个 Nirvana 方法。比如某个 API 是返回一个资源列表的，那么 Nirvana 方法就应该是 List 而不是 Get。
 
### Definition Source
Definition Source 用于描述一个业务函数的参数的来源和默认值。

| 参数来源 | 名称 | 描述                                                                                                         |
|----------|------|--------------------------------------------------------------------------------------------------------------|
| Path     | 有   | 参数值来源于 API Path                                                                                        |
| Query    | 有   | 参数值来源于 URL Query                                                                                       |
| Header   | 有   | 参数值来源于 Request Header                                                                                  |
| Form     | 有   | 参数值来源于 Request Body，但是 Content-Type 必须是 application/x-www-form-urlencoded 或 multipart/form-data |
| File     | 有   | 参数值来源于 Request Body，但是 Content-Type 必须是 multipart/form-data                                      |
| Body     | 无   | 参数值来源于 Request Body                                                                                    |
| Auto     | 无   | Auto 类型对应的参数必须是一个结构体，通过结构体的 tag 定义来确定每个字段的来源                               |
| Prefab   | 有   | 参数值来源于当前 server 内部，比如一个 DB 链接                                                               |

Auto 类型的 tag 范例如下：

```go
type Example struct {
	Start       int    `source:"Query,start,default=100"`
	ContentType string `source:"Header,Content-Type"`
}
```
tag 名称为 `source`。值使用逗号分隔，第一个参数表示参数来源，第二个表示名称。如果是 Body 类型名称可以为空。
如果需要给字段设置默认值，则需要使用 `default={value}` 的形式。 

如果有多个 Auto 结构体，可以组合成一个：
```go
type AnotherAutoStruct struct {
	...
}

type Example struct {
	Start       int    `source:"Query,start,default=100"`
	ContentType string `source:"Header,Content-Type"`
	AnotherAutoStruct
}
```
对于没有 `source` 的结构体类型，会递归遍历以寻找带有 `source` 的字段。忽略所有没有 `source` 的字段。



### Definition Destination
Definition Destination 用于描述一个业务函数的参数的来源和默认值。

| 返回值目标 | 描述                                                           |
|------------|----------------------------------------------------------------|
| Meta       | 这个返回值类型必须是 map[string]string，会写入 Response Header |
| Data       | 返回值可以是任意结构，自动转换并写入到 Response Body           |
| Error      | 错误类型，必须是 error                                         |

## 给项目添加一个 API
接下来我们给项目增加一个 API，用于获取一条消息：
```go
func init() {
	register([]def.Descriptor{{
		Path:        "/messages",
		Definitions: []def.Definition{listMessages},
	}, {
		// 获取一条消息的 Descriptor。
		Path:        "/messages/{message}",
		Definitions: []def.Definition{getMessage},
	},
	}...)
}

// 获取一条消息的 API 定义。
var getMessage = def.Definition{
	// 因为只获取一条消息，此处为 Get。
	Method:      def.Get,
	Summary:     "Get Message",
	Description: "Get a message by id",
	// 业务函数
	Function:    message.GetMessage,
	Parameters: []def.Parameter{
		// 这是一个工具方法，用于快速生成一个参数结构。
		// message 是从 API Path 里获取的。
		def.PathParameterFor("message", "Message id"),
	},
	Results: def.DataErrorResults("A message"),
}
```

对应的业务函数如下：
```go
// GetMessage return a message by id.
func GetMessage(ctx context.Context, id int) (*Message, error) {
	return &Message{
		ID:      id,
		Title:   "This is an example",
		Content: "Example content",
	}, nil
}
```

添加 API 之后，编译运行。然后访问 `http://localhost:8080/apis/v1/messages/100`，即可获得结果（默认情况下都是 json 类型）：
```
{"id":100,"title":"This is an example","content":"Example content"}
```

这里添加的业务函数都是以 Golang 函数的方式呈现的。如果希望使用实例方法作为业务的处理函数，请参考 [方法包](../concepts/method.md)。
