# Caicloud.yml 参考文档

## Pre Build

在执行构建前阶段时，Cyclone 会根据定义的镜像和命令去执行，最后，与其他环节不同的是，构建前环节需要指定 `outputs` 字段，在构建结束后，被定义在 `outputs` 字段中的文件或目录会被拷贝出容器，供构建环节构建镜像时使用。

```yml
pre_build:
  image: golang:v1.5.3
    volumes:
      - .:/root
    environment:
      - key = value
    commands:
      - ls
      - pwd
    outputs:
        - file1
        - dir2
```

## Build

如果要启用 `caicloud.yml` 中的构建环节，需要定义 `context\_dir` 和 `dockerfile_name` 其中的至少一个。

```yml
build:
  context_dir: publish
  dockerfile_name: Dockerfile_publish
```

Cyclone会在 `<root_dir>/<context_dir>` 目录下，根据文件 `<root_dir>/<context_dir>/<dockerfile_name>` 来进行镜像的构建。如果镜像被成功构建，就会将其推送到镜像仓库中。

## Integration

Cyclone 在进行持续集成时，会使用Build阶段构建的镜像运行一个 docker 容器。与此同时，Cyclone 支持在执行持续集成时同时运行多个 service。这些 service 是以独立容器的方式运行的，与持续集成容器之间可以进行直接地通信。比如，当持续集成需要进行数据库访问时，可以将数据库以 service 的方式启动，持续集成容器可以访问到该容器。

Service 容器和持续集成容器是在同一个docker network mode 下的，所以在持续集成容器中可以通过服务名和端口，就可以直接访问到服务容器。**当service的镜像名配置为“BUILT_IMAGE”时，会使用Build阶段新构建的镜像。**

```yml
integration:
  services:
    postgres:
      image: postgres:9.4.5
      enviroment:
        - key = value
      commands:
        - cmd1
        - cmd2
    app:
      image: BUILT_IMAGE
      enviroment:
        - key = value
      commands:
        - cmd1
        - cmd2
  environment:
    - key = value
  commands:
    - ls
    - pwd
    - ping postgress
```

## Post Build

Cyclone 支持构建后阶段，构建后阶段会运行在一个单独的容器中，代码会被以 volume 的形式挂载在容器中。我们推荐你可以将需要用到的二进制等等都打包在镜像中，而把构建后需要执行的逻辑写在 `commands` 字段中。

```yml
post_build:
  image: golang:v1.5.3
  environment:
    - key = value
  commands:
    - ls
    - pwd
```

## Depoly
Cyclone 可以把最新发布到镜像仓的镜像部署更新到指定的多个应用上。Cyclone 还可以将镜像部署到 [Caicloud Cubernetes](https://caicloud.io/products/cubernetes) 和 [Google Kubernetes](http://kubernetes.io/) 上。

#### Caicloud Cubernetes

```yml
#deploy段
deploy:
  - deployment: redis-master
    cluster: cluster1_id
    namespace: namespace1_id
    containers:
      - container1
      - container2
  - deployment: redis-slave
    cluster: cluster2_id
    namespace: namespace2_id
    containers:
      - container1
      - container2
```

#### Google Kubernetes

```yml
#deploy段
deploy:
  - type: kubernetes 
    host: <cluster host>
    token: <cluster access token>
    cluster: cluster1_id
    namespace: namespace1_id
    deployment: redis-master
    containers:
      - container1
      - container2
  - type: kubernetes 
    host: <cluster host>
    token: <cluster access token>
    cluster: cluster2_id
    namespace: namespace2_id
    deployment: redis-slave
    containers:
      - container1
      - container2
```
