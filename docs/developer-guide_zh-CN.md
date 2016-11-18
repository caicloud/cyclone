# 开发者指南

这份文档提供给想为 Cyclone 贡献代码或者文档的用户。

## 向 Cyclone 进行贡献的工作流

因为 Cyclone 是在 Github 上开源的，因此我们使用 [Github Flow](https://guides.github.com/introduction/flow/) 作为协作的工作流，你可以花五分钟的时间来了解下它 :)

## 搭建你的开发环境

我们已经写了一些 bash 脚本来帮助你搭建开发环境。如果你想在本地运行一个 Cyclone 服务，可以通过：

```shell
./scripts/local-up.sh
```

来实现，这段脚本会在容器中启动所有的依赖服务，在**本地**编译和运行 Cyclone server，相比于把全部服务都运行在容器中的方式，在本地运行 Cyclone server 更容易开发与调试。

请注意，如果你的 docker daemon 是运行在一个 docker machine 中的，那么你可能需要做一些额外的工作，比如端口映射等等。

## 测试 Cyclone

我们有单元测试和端到端的测试用例，当你想为 Cyclone 贡献时可以使用它们来进行代码的测试。

### 单元测试

你可以通过：

```shell
./tests/run-unit-test.sh
```

来进行单元测试，与此同时目前我们的 Travis 也会进行这样的单元测试，因此你也可以在 Travis 的构建日志中查看测试结果。在之后我们会使用 Cyclone 来对 Cyclone 进行单元测试。

### 端到端测试

目前 Cyclone 的端到端测试会先启动一个 Cyclone server，然后在另外一个独立的进程中通过发送 RESTful 请求的方式对其进行测试，并验证结果。

你可以通过：

```shell
./tests/run-e2e.sh
```

来进行端到端的测试，在之后我们会使用 Cyclone 来对 Cyclone 进行端到端的测试。

## API 文档

我们使用 [swagger ui](https://github.com/swagger-api/swagger-ui) 来生成 API 文档，如果你的工作影响了 Cyclone 的 API，你可以在 `http://<your cyclone server host>:7099/apidocs` 查看最新的 API 文档，或者你可以通过我们的[在线文档](http://118.193.142.27:7099/apidocs/)来进行开发与贡献。

## Cyclone 架构以及工作流

### 工作流

![flow](flow.png)

- Cyclone提供了丰富的[API](http://118.193.142.27:7099/apidocs/)供web应用调用（详见API说明）
- 通过API建立版本控制系统中代码库与Cyclone服务关联关系后，版本控制系统的提交、发布等动作会通过webhook通知到Cyclone-Server
- Cyclone-Server启动一个基于Docker in Docker技术的Cyclone-Worker容器，在该容器中从代码库中拉取源码，按照源码中caicloud.yml配置文件，依次执行：
  - PreBuild：在指定编译环境中编译可执行文件
  - Build：将可执行文件拷到运行环境容器中，打成镜像发布到镜像仓库中
  - Integration：运行持续集成所依赖的微服务，启动一个容器执行集成测试，如果微服务镜像名配置为“BUILT_IAMGE”，则使用Build阶段新构建的镜像
  - PostBuild：启动一个容器执行一些脚本命令，实现镜像发布后的一些关联操作
  - Deploy：使用发布的镜像部署应用到kubernetes等容器集群Paas平台
- 构建过程日志可以通过Websocket从Cyclone-Server拉取
- 构建结束后Cyclone-Server将构建结果和完整构建日志通过邮件通知用户

### 软件架构

![architecture](architecture.png)

每个立方体代表一个容器

- Cyclone-Server中Api-Server组件提供Restful API服务，被调用后需要较长时间处理的任务生成一个待处理事件写入etcd
- EventManager加载etcd中未完成事件，监视事件变化，发送新增待处理事件到WorkerManager中
- WorkerManager调用Docker API启一个Cyclone－Worker容器，通过环境变量传入需要处理的事件ID等信息
- Cyclone-Worker使用事件ID作为token（有效期2小时）调用API，拉取事件信息依次启容器执行integration、prebuild、build、post build，完成后反馈事件执行结果，构建过程日志推送到Log-Server，转存到Kafka
- Log-Server组件从Kafka拉取日志推送给用户
- 需要持久化的数据存入mongo
