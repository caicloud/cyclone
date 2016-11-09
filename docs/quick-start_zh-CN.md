# 快速开始

* [介绍](#介绍)
* [节点注册](#节点注册)
* [创建服务](#创建服务)
* [创建版本](#创建版本)
* [查看日志](#查看日志)

## 介绍

这篇文档会对 Cyclone 的使用进行简单的介绍。首先需要向 Cyclone server 注册一个 worker 节点。随后可以通过 RESTful API 来与 Cyclone server 进行交互，完成项目的持续集成与持续部署。

## 节点注册

Cyclone 是 Master/Slave 的架构，因此为了能够在 worker 节点上运行构建任务，需要先向 Cyclone server 注册不少于一个的 worker 节点，worker 节点可以与 Cyclone server 在同一台机器上。

```shell
# Create a worker node.
curl -X POST -H "Content-Type: application/json" -d '{
         "name": "First Node",
         "description": "Just for test",
         "ip": "127.0.0.1",
         "docker_host": "'$DOCKER_HOST'",
         "type": "system",
         "total_resource": {
             "memory": 1024,
             "cpu": 1
         }
}' ''$HOST_URL'/api/v0.1/system_worker_nodes'
```

上述的脚本会在本地计算机上，创建一个 worker 节点。一个 worker 节点是一个 docker host。接下来就可以去创建需要进行持续集成与持续部署的项目。

## 创建服务

一个 service（服务）代表了一个被版本管理工具所管理的repo（代码库）。Cyclone 对于每个项目而言，都需要通过一个POST请求来进行创建，在创建的过程中会进行仓库地址的校验等等。在创建成功后就可以进行对版本的构建。

```shell
# Create a service from a github repo.
curl -sS -X POST -H "Content-Type:application/json" -d '{
   "name": "Test-Service",
   "username": "'$USERNAME'",
   "description": "This is a test-service",
   "repository": {
     "url": "https://github.com/caicloud/toy-dockerfile",
     "vcs": "git"
   }
}' ''$HOST_URL'/api/v0.1/'$USER_ID'/services'
```

## 创建版本

如果在你的代码的根目录中有 `caicloud.yml` 文件，那么 Cyclone 会根据该配置文件来进行版本的创建。如果该文件不存在，Cyclone 会按照默认的步骤来进行版本的构建与发布。

```shell
# Create a version manually.
curl -sS -X POST -H "Content-Type:application/json" -d '{
   "name": "v0.1.0",
   "description": "just for test",
   "service_id": "'$SERVICE_ID'"
}' ''$HOST_URL'/api/v0.1/'$USER_ID'/versions'

```

为了能够执行镜像的构建，代码的根目录下必须有 `Dockerfile` 文件。在 Caicloud.yml 参考文档中我们将提供更多的自定义的功能，其中包括使用自定义名的文件作为 dockerfile 进行镜像创建。

在得到 dockerfile 后，Cyclone 会根据根目录下的 dockerfile 去进行镜像的构建，随后会将发布后的镜像发布到 registry 上。

## 查看日志

在版本创建过程中和结束后，可以通过 API 请求得到构建过程的日志。如果版本仍然在构建，日志可以通过 websocket 请求来得到。

Websocket 日志服务器默认在 8001 端口，通过建立 websocket 链接并发送包来接收日志：

```
{
    "action" : "watch_log", 
    "api" : "create_version", 
    "user_id" : "${USER_ID}", 
    "service_id" : "${SERVICE_ID}", 
    "version_id" : "${VERSION_ID}",
    "operation" : "start", 
    "id" : "${UUID}"
}
```

随后 Cyclone 会发送日志包：

```
{
    "action" : "push_log", 
    "api" : "create_version", 
    "user_id" : "${USER_ID}", 
    "service_id" : "${SERVICE_ID}", 
    "version_id" : "${VERSION_ID}",
    "log" : "...", 
    "id" : "${UUID}"
}
```

Cyclone 超过 120 秒收不到客户端的任何合法数据包，会主动断开链路，客户端如需维持链路需定时发送心跳：

```
{
    "action" : "heart_beat", 
    "id" : "${UUID}"
}
```

在版本创建结束后，可以通过 HTTP GET 请求得到日志。

```shell
# Get log from Cyclone server.
curl -v -sS ''$HOST_URL'/api/v0.1/'$USER_ID'/versions/'$VERSION_ID'/logs'
```

得到的日志会是如下格式：

```text
 step: clone repository state: start 
Cloning into 'b05a2eeb-6665-41fb-8c55-e594384981c5'...
 step: clone repository state: finish 
Step 1 : FROM busybox
Pulling from library/busybox
Pulling fs layer
Downloading [==================================================>] 668.2 kB/668.2 kB
Verifying Checksum
Download complete
Extracting [==================================================>] 668.2 kB/668.2 kB
Pull complete
Digest: sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912
Status: Downloaded newer image for busybox:latest
 ---> e02e811dd08f
Step 2 : COPY ./echo.sh /echo.sh
 ---> 9d839b8ec299
Removing intermediate container 5b23ff6df449
Step 3 : CMD /echo.sh
 ---> Running in 35de934148fb
 ---> 207a093246c2
Removing intermediate container 35de934148fb
Successfully built 207a093246c2
The push refers to a repository [cargo.caicloud.io/<user>/toy-dockerfile]
Preparing
Pushing [==================================================>]    512 B
Pushing [==================================================>]    544 B
Pushing [==================================================>] 2.048 kB
Pushing [==================================================>] 1.294 MB
Pushed
11: digest: sha256:37e5b97bc527717d2d133d84e83d73d39b4a2da6535aa6ce29e623c44bdeffc9 size: 712
Clair analysis images got vulnerabilits: 0
```
