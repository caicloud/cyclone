# Consumer and Producer

包路径: `github.com/caicloud/nirvana/service`

在 HTTP 请求中，Content-Type 说明了请求和响应中的数据类型。为了根据 Content-Type 自动处理数据转换，Nirvana 提供了 Consumer 和 Producer 接口。其中 Consumer 用于将请求体中的数据转换为业务函数需要的类型，而 Producer 则负责将业务函数的返回结果写入到响应体中。 

Nirvana 默认提供的 Consumers：

| Content-Type                      | 描述                                                                                                              |
|-----------------------------------|-------------------------------------------------------------------------------------------------------------------|
|                                   | 空的 Content-Type 通常对应于 GET 之类的请求，因此不能转换为任何数据类型。                                         |
| text/plain                        | 只能生成 string 和 []byte 类型                                                                                    |
| application/json                  | 如果接收类型是 string 和 []byte，则直接将数据转换为这两个类型。对于其他类型，使用 json.Unmarshal 进行解析。       |
| application/xml                   | 如果接收类型是 string 和 []byte，则直接将数据转换为这两个类型。对于其他类型，使用 xml.Unmarshal 进行解析。        |
| application/octet-stream          | 只能生成 string 和 []byte 类型                                                                                    |
| application/x-www-form-urlencoded | 只能生成 string 和 []byte 类型，这种类型的请求通常会被 Parse 并成为 Form 类型，因此一般不转换为具体类型。         |
| multipart/form-data               | 只能生成 string 和 []byte 类型，这种类型的请求通常会被 Parse 并成为 Form 或 File 类型，因此一般不转换为具体类型。 |

Nirvana 默认提供的 Producers：

| Content-Type             | 描述                                                                                                                               |
|--------------------------|------------------------------------------------------------------------------------------------------------------------------------|
|                          | 空的 Content-Type 通常对应于 204 之类的响应，没有响应体，不需要写入。                                                              |
| text/plain               | 如果类型符合 io.Reader 接口或者是 string 和 []byte，则直接将数据写入到响应。                                                       |
| application/json         | 如果类型符合 io.Reader 接口或者是 string 和 []byte，则直接将数据写入到响应。如果是其他类型，则使用 json.Marshal 将数据写入到响应。 |
| application/xml          | 如果类型符合 io.Reader 接口或者是 string 和 []byte，则直接将数据写入到响应。如果是其他类型，则使用 xml.Marshal 将数据写入到响应。  |
| application/octet-stream | 如果类型符合 io.Reader 接口或者是 string 和 []byte，则直接将数据写入到响应。                                                       |


## 添加 Consumer 和 Producer

在业务的实际场景中，默认提供的 Consumers 和 Producers 可能不能满足实际使用需求。因此 Nirvana 的 service 包提供了相应的工具用于注册用户自己的 Consumer 和 Producer。


### 注册 Consumer

Consumer 需要实现接口：
```go
// Consumer handles specifically typed data from a reader and unmarshals it into an object.
type Consumer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Consume unmarshals data from r into v.
	Consume(r io.Reader, v interface{}) error
}
```

实现了这个接口后，通过 service 的注册方法即可注册 Consumer：

```go
if err := service.RegisterConsumer(consumer); err != nil {
	log.Fatal(err)
}
```

在接收到 Content-Type 与 consumer 一致的请求，并且业务函数需要从请求中取得请求体的时候，就会调用这个 Consumer 去读取数据并进行类型转换。


### 注册 Producer

Producer 需要实现接口：
```go
// Producer marshals an object to specifically typed data and write it into a writer.
type Producer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Produce marshals v to data and write to w.
	Produce(w io.Writer, v interface{}) error
}
```

实现了这个接口后，通过 service 的注册方法即可注册 Producer：
```go
if err := service.RegisterProducer(producer); err != nil {
	log.Fatal(err)
}
```

在需要生成 Conetent-Type 于 producer 一致的响应，并且业务函数需要返回数据的时候，就会调用这个 producer 将类型转换为字节数据写入到响应体中。


### 快速生成 Consumer 和 Producer 的工具

通常情况下，我们需要快速添加一些 Consumers 和 Producers，并且他们的行为和 `application/octet-stream` 一致的时候，那么可以直接使用工具方法：
```go
serializer := NewSimpleSerializer(contentType)

if err := service.RegisterConsumer(serializer); err != nil {
	log.Fatal(err)
}

if err := service.RegisterProducer(serializer); err != nil {
	log.Fatal(err)
}
```



