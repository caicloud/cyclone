# Caicloud.yml 简介

当给定 `caicloud.yml` 作为配置文件后，Cyclone 会根据配置文件来执行创建版本的过程。

下面的代码是一个 `caicloud.yml` 的例子：

```yml
integration:
  image: golang:1.5
  environment:
    - GO15VENDOREXPERIMENT=1
    - GOOS=linux
    - GOARCH=amd64
    - CGO_ENABLED=0
  commands:
    - go run main.go
```

`caicloud.yml` 配置文件最多有六个部分，分别是 integration，pre\_build，build，post_build，和 deploy。需要注意的是这些部分都不是必须要被定义在 `caicloud.yml` 里的，在配置的时候可以根据需要来决定实现哪些部分。

## Pre Build

传统的构建方式，会使得代码遗留在容器镜像中，这样会导致一些安全问题。所以 Cyclone 引入了构建前这样一个环节。它支持在构建前事先根据代码构建出二进制文件，这个过程也是在容器中完成的。在构建完成后会把得到的二进制文件从容器中拷贝出来，供之后构建镜像的环节使用。

```yml
pre_build:
  image: golang:1.6
  volumes:
    - .:/go/src/github.com/caicloud/ci-demo-go
  commands:
    - echo "compile executable files"
    - cd /go/src/github.com/caicloud/ci-demo-go/code
    - go build -v -o app
  outputs: # copy out publish executable files from prebuild container
    - /go/src/github.com/caicloud/ci-demo-go/code/app
```

## Build

在构建环节，可以指定构建时的工作目录，和 Dockerfile。默认情况下是使用根目录和 Dockerfile 来进行构建。在构建结束后 Cyclone 会将构建好的镜像推送到镜像仓库中。这是建立在在 Build 之前的两个环节都成功结束的基础上的。

```yml
build:
  context_dir: publish
  dockerfile_name: Dockerfile_publish
```

## Integration

Integration 部分，是集成部分。Cyclone 首先会使用Build段构建的镜像启动一个容器，然后根据配置文件的定义，先进行持续集成。持续集成过程运行在一个容器中，代码是以 Volume 的形式被挂载到持续集成容器中的。如果持续集成失败，那么就不会执行后面的逻辑，版本创建失败。

```yml
integration:
  environment:
    - GO15VENDOREXPERIMENT=1
    - GOOS=linux
    - GOARCH=amd64
    - CGO_ENABLED=0
  commands:
    - go run main.go
```

## Post Build

Cyclone 支持构建结束后的 hook。在一些用例中，构建结束后会有做一些联合发布等等的需求。而构建结束后的 hook 可以满足这样的需求。

```yml
post_build:
  image: golang:1.6
  commands:
    - echo "Built Successfully."
```

## Deploy

Cyclone 可以将镜像部署到 [Caicloud Cubernetes](https://caicloud.io/products/cubernetes) 和 [Google Kubernetes](http://kubernetes.io/) 上。

### Caicloud Cubernetes

```yml
deploy:
  - deployment: mongo-server
    cluster: 1e73520d-f7ab-4998-b169-41b8c122342b
    namespace: test
    containers:
      - mongo-server
```

### Google Kubernetes

```yml
deploy:
  - type: kubernetes 
    host: <cluster host>
    token: <cluster access token>
    cluster: <cluster name>
    namespace: <namespace name>
	deployment: <deployment name>
    containers:
      - mongo-server
```

Cyclone 还可以支持部署多个应用到多个集群。你只需要跟上面做法一样，把相关信息加到“deploy”段。

更多 `caicloud.yml` 使用说明详见 [caicloud.yml 参考文档](./caicloud-yml-reference_zh-CN.md)。
