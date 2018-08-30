# 多客户端整合

默认情况下，我们会为每个服务生成如下结构的客户端：
```go
client
├── client.go
├── v1
│   ├── client.go
│   └── types.go
└── v2
```

在微服务的场景下，会生成大量的客户端，直接导致用户在使用客户端时引用各种各样的项目。为了解决这个问题，可以建立一个综合项目，然后将所有服务的客户端生成到该项目中：
```
clientset
├── svca
│   ├── client.go
│   ├── v1
│   │   ├── client.go
│   │   └── types.go
│   └── v2
└── svcb
    ├── client.go
    ├── v1
    │   ├── client.go
    │   └── types.go
    └── v2 
```

## 整合客户端

为了演示这个过程，我们逐步构建这个项目。

### 创建项目

首先创建 `clientset` 项目，用于保存所有服务的客户端：
```
$ cd $GOPATH/src/
$ mkdir clientset
```
然后创建两个服务项目（仅用于演示）：
```
$ nirvana init svca
$ nirvana init svcb
```
即创建了 svca 和 svcb 两个服务项目。


### 生成客户端

生成 `svca` 的客户端：
```
$ cd $GOPATH/src/svca
$ nirvana client --output ../clientset/svca
```

生成 `svcb` 的客户端：
```
$ cd $GOPATH/src/svcb
$ nirvana client --output ../clientset/svcb
```

此时 `clientset` 的项目结构如下：
```
clientset
├── svca
│   ├── client.go
│   └── v1
│       ├── client.go
│       └── types.go
└── svcb
    ├── client.go
    └── v1
        ├── client.go
        └── types.go
```

这样所有客户端都在一个项目中，不需要依赖其他服务项目。


## 统一网关访问


在某些场景下，所有的微服务会通过一个公共的网关进行暴露。这样就需要再进一步对客户端进行整合。

### 创建 ClientSet

在 `clientset` 中创建 `clientset.go`：
```
$ cd $GOPATH/src/clientset
$ touch clientset.go
```

`clientset.go` 代码如下：
```go
package clientset

import (
	svca "clientset/svca"
	svcb "clientset/svcb"

	rest "github.com/caicloud/nirvana/rest"
)

// Interface describes a clientset.
type Interface interface {
	// SvcA returns a client for svc a.
	SvcA() svca.Interface
	// SvcB returns a client for svc b.
	SvcB() svcb.Interface
}

// ClientSet contains multiple clients.
type ClientSet struct {
	svcA svca.Interface
	svcB svcb.Interface
}

// NewClientSet creates a new client set.
func NewClientSet(cfg *rest.Config) (Interface, error) {
	c := &ClientSet{}
	var err error

	c.svcA, err = svca.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	c.svcB, err = svcb.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// MustNewClientSet creates a new client set or panic if an error occurs.
func MustNewClientSet(cfg *rest.Config) Interface {
	c, err := NewClientSet(cfg)
	if err != nil {
		panic(err)
	}
	return c
}

// SvcA returns a client for svc a.
func (c *ClientSet) SvcA() svca.Interface {
	return c.svcA
}

// SvcB returns a client for svc b.
func (c *ClientSet) SvcB() svcb.Interface {
	return c.svcB
}
```

### 使用 ClientSet

`ClientSet` 的使用方法和普通 `Client` 没有太大区别：
```go
func main() {
	cli := clientset.MustNewClientSet(&rest.Config{
		Scheme: "http",
		Host:   "localhost:8080",
	})
	msgs, err := cli.SvcA().V1().ListMessages(context.Background(), 10)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(msgs)
}
```

