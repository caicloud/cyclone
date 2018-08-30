# 项目结构和初始化

## 创建项目
Nirvana 创建项目非常简单，不过在创建项目之前，首先需要下载安装 Nirvana 的命令行和其他依赖工具：
```
$ go get -u github.com/caicloud/nirvana/cmd/nirvana
$ go get -u github.com/golang/dep/cmd/dep
```
然后就可以直接使用命令创建项目（请确保 `$GOPATH/bin` 在 `$PATH` 中）：
```
$ cd $GOPATH/src/
$ nirvana init ./myproject
$ cd ./myproject
```

此时在 `$GOPATH/src/myproject` 会生成一个完整的 Nirvana 项目。项目结构如下：
```
 .
├── bin                                # 存放编译后的二进制文件
├── build                              # 存放项目的 Dockerfile 以及与构建相关的文件
│   └── myproject                      # 
│       └── Dockerfile                 #
├── cmd                                # 存放项目的启动命令
│   └── myproject                      #
│       └── main.go                    # 
├── Gopkg.toml                         #
├── Makefile                           #
├── nirvana.yaml                       #
├── pkg                                # 所有的业务逻辑都应该在这个目录中
│   ├── apis                           # 所有与 Nirvana API 定义相关的代码都在这个目录中
│   │   ├── api.go                     #
│   │   ├── filters                    # 存放 HTTP Request 过滤器
│   │   │   └── filters.go             #
│   │   ├── middlewares                # 存放 Nirvana 中间件
│   │   │   └── middlewares.go         #
│   │   ├── modifiers                  # 存放 Nirvana Definition 修改器
│   │   │   └── modifiers.go           #
│   │   └── v1                         # 存放项目 v1 版本所有的 API 定义
│   │       ├── converters             # 存放 v1 版本需要用到的类型转换器
│   │       ├── descriptors            # 存放 v1 版本的 API 描述符
│   │       │   ├── descriptors.go     #
│   │       │   └── message.go         # 对应 message 业务逻辑的 API 定义，可以修改或删除
│   │       └── middlewares            # 存放 v1 版本需要用到的中间件
│   │           └── middlewares.go     #
│   ├── message                        # 业务逻辑目录，这个目录是生成的样板，可以修改或删除
│   │   └── message.go                 #
│   └── version                        # 项目版本信息目录
│       └── version.go                 #
├── README.md                          #
└── vendor                             #
```

这个项目中包含了编译和构建容器的基本工具(Makefile 和 Dockefile)，还有一个 `golang/dep` 的需要的包定义文件 `Gopkg.toml`。通过如下命令即可完成依赖包的安装：
```
$ dep ensure -v
```
到这里就完成了整个项目的创建和依赖安装工作，默认的项目结构中自带了一个 API 范例，因此可以直接运行查看效果。


## 编译运行

### 直接编译运行
Nirvana 创建项目时自动生成了 Makefile，只需要使用简单的 `make` 命令就可以完成编译工作：
```
$ make
```
在项目的 `bin` 目录中就能看到编译后的二进制文件，直接运行：
```
$ ./bin/myproject
```
启动时会打印出 Nirvana 的 Logo 以及当前项目的版本信息以及监听的端口，默认端口是 8080。

在服务启动之后，可以通过浏览器或者命令行访问 `http://localhost:8080/apis/v1/messages`：
```
$ curl http://localhost:8080/apis/v1/messages
```
就能够看到 API 的返回结果。


### 编译并打包成 Docker 镜像
在需要发布的时候，通常需要打包成镜像的形式，在 Makefile 中也提供了直接打包成镜像的命令：
```
$ make container
```
就会自动开始在容器内编译和打包镜像。不过这个过程中需要 `golang:latest` 和 `debian:jessie` 这两个镜像。如果本地没有这两个镜像，或者希望使用其他镜像进行编译和构建工作，请修改 `./build/myproject/Dockerfile` 。

打包完成后，可以通过 Docker 命令启动容器，然后进行访问：
```
$ docker run -p 8080:8080 myproject:v0.1.0
```

## Nirvana 项目配置
每个 Nirvana 项目都有一个 `nirvana.yaml` 配置文件，用于描述项目的基本信息和结构。
```yaml
# 项目名称
project: myproject
# 项目描述
description: This project uses nirvana as API framework
# 服务使用的协议，只能填写 http 和 https
schemes:
- http
# 访问 IP 或域名，可以有多个
hosts:
- localhost:8080
# 项目负责人
contacts:
- name: nobody
  email: nobody@nobody.io
  description: Maintain this project
# 项目 API 版本信息，用于区分不同版本的 API
# 用于文档和客户端生成
versions:
  # 版本名称
- name: v1
  # 版本描述
  description: The v1 version is the first version of this project
  # 版本规则
  rules:
    # 路径前缀，匹配前缀为 "/apis/v1" 的 API
  - prefix: /apis/v1
    # 正则表达式，用于匹配路径
    # 如果设置了 prefix，那么 regexp 字段无效
    regexp: ""
    # 这个字段仅用于在生成文档和客户端的时候，替换匹配的 API 路径。为空时不会进行替换。
    # 比如设置 replacement = "/apis/myproject/v1"
    # 那么 "/apis/v1/someapi" 为被替换为 "/apis/myproject/v1/someapi"
    replacement: ""
```
这个配置文件不会影响 Server 的运行，只用于描述项目的信息以及区分不同版本的 API。API 文档生成和客户端生成会依赖这个配置文件进行 API 版本识别和 API 路径替换，因此需要正确设置版本规则。

