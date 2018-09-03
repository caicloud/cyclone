# 生成客户端


## 生成代码

API 文档通常足够使用者使用。但是在微服务场景下，服务之间也会存在调用关系。因此需要使调用者方便快速的进行 API 调用，可以生成客户端以供使用：
```
$ nirvana client
```
该命令默认会在 `./client` 目录下生成 golang 客户端代码，可以通过 `--output` 指定其他输出目录（目前仅支持生成 golang 客户端，其他语言客户端尚不支持）。

输出日志如下（项目路径和日志时间会有所不同）：
```
INFO  0702-15:50:54.156+08 client.go:73 | No packages are specified, defaults to pkg/apis
INFO  0702-15:50:55.609+08 client.go:81 | Project root directory is /home/go/src/myproject
INFO  0702-15:50:55.610+08 client.go:106 | Generated golang client package myproject/client
```

生成的客户端代码如下：
```
client
├── client.go
└── v1
    ├── client.go
    └── types.go
```

`./client.go` 生成代码:
```go
package client

import (
	v1 "myproject/client/v1"

	rest "github.com/caicloud/nirvana/rest"
)

// Interface describes a versioned client.
type Interface interface {
	// V1 returns v1 client.
	V1() v1.Interface
}

// Client contains versioned clients.
type Client struct {
	v1 *v1.Client
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (Interface, error) {
	c := &Client{}
	var err error

	c.v1, err = v1.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) Interface {
	return &Client{
		v1: v1.MustNewClient(cfg),
	}
}

// V1 returns a versioned client.
func (c *Client) V1() v1.Interface {
	return c.v1
}
```

`./v1/client.go` 生成代码
```go
package v1

import (
	"context"

	rest "github.com/caicloud/nirvana/rest"
)

// Interface describes v1 client.
type Interface interface {
	// GetMessage return a message by id.
	GetMessage(ctx context.Context, message int) (message1 *Message, err error)
	// ListMessages returns all messages.
	ListMessages(ctx context.Context, count int) (messages []Message, err error)
}

// Client for version v1.
type Client struct {
	rest *rest.Client
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (*Client, error) {
	client, err := rest.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) *Client {
	client, err := NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

// GetMessage return a message by id.
func (c *Client) GetMessage(ctx context.Context, message int) (message1 *Message, err error) {
	message1 = new(Message)
	err = c.rest.Request("GET", 200, "/apis/v1/messages/{message}").
		Path("message", message).
		Data(message1).
		Do(ctx)
	return
}

// ListMessages returns all messages.
func (c *Client) ListMessages(ctx context.Context, count int) (messages []Message, err error) {
	err = c.rest.Request("GET", 200, "/apis/v1/messages").
		Query("count", count).
		Data(&messages).
		Do(ctx)
	return
}
```
 
`./v1/types.go` 生成代码：
```go
package v1

// Message describes a message entry.
type Message struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
```

生成的客户端是版本化的，版本在 nirvana.yaml 中定义。API 依赖的结构体都会被提取出来并生成到 `types.go` 文件中，方便客户端使用。

每个服务都可以生成一份客户端，为了方便使用，可以将客户端整合在一起。具体内容请参考 [多客户端整合](../concepts/clients.md)。

## 使用客户端

客户端的使用非常简单，只需要创建客户端，然后调用相应的 API 函数即可：
```go
func main() {
	cli := client.MustNewClient(&rest.Config{
		Scheme: "http",
		Host:   "localhost:8080",
	})
	msgs, err := cli.V1().ListMessages(context.Background(), 10)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(msgs)
}
```

## 限制

目前客户端生成对接口的支持还不完善，仅仅对 `io.Reader` 和 `io.ReadCloser` 支持良好，因此在 API 的参数和返回值中，尽量避免使用其他接口。
