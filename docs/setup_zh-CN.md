# 安装

## 版本依赖

```
golang：1.6+
docker：建议使用 1.10.0+(已在1.10~1.11版本充分测试验证)
kubernetes：1.2+
```

## 介绍

这篇文档会对 Cyclone 的安装进行简单的介绍。这里有两种方式，一种是使用 docker compose，一种是使用 kubectl。

在此之前，你需要先编译生成 Cyclone-server 和 Cyclone-worker 镜像，操作命令如下：

```
git clone https://github.com/caicloud/cyclone
cd cyclone
./scripts/build-server.sh
./scripts/build-worker.sh
```

- 使用 docker compose 方式。 你可以去查看 docker-compose.yml 来了解运行过程。具体的操作指令如下：

```
docker-compose -f docker-compose.yml up -d
```

- 使用 kubectl 方式，需要一些 YAML 文件。你可以在 cyclone/scripts/k8s 目录下查看这些文件，并根据自己的实际情况对其中的参数做简单的调整，然后执行下面这些命令:

```
git clone https://github.com/caicloud/cyclone
cd cyclone/scripts/k8s
kubectl create namespace cyclone
kubectl --namespace=cyclone create -f mongo.yaml
kubectl --namespace=cyclone create -f mongo-svc.yaml
kubectl --namespace=cyclone create -f cyclone.yaml
kubectl --namespace=cyclone create -f cyclone-svc.yaml
```

这样 Cyclone 就启动了。

## 其它

环境变量表：

| 环境变量                   | 说明                                       |
| ------------------------- | ---------------------------------------- |
| MONGODB_HOST              | mongo db的地址, 默认是localhost                  |
| REGISTRY_LOCATION         | 镜像仓库的地址，默认是cargo.caicloud.io             |
| REGISTRY_USERNAME         | 镜像仓库默认用户名，默认是空                             |
| REGISTRY_PASSWORD         | 镜像仓库默认用户密码，默认是空                           |
| LIMIT_MEMORY              | 与[kubernetes limits.memory](https://kubernetes.io/docs/concepts/policy/resource-quotas)概念相同，用于限制cyclone-worker，默认是1Gi     |
| LIMIT_CPU                 | 与[kubernetes limits.cpu](https://kubernetes.io/docs/concepts/policy/resource-quotas)概念相同，用于限制cyclone-worker，默认是1          |
| REQUEST_MEMORY            | 与[kubernetes requests.memory](https://kubernetes.io/docs/concepts/policy/resource-quotas)概念相同，用于限制cyclone-worker，默认是0.5Gi |
| REQUEST_CPU               | 与[kubernetes requests.cpu](https://kubernetes.io/docs/concepts/policy/resource-quotas)概念相同，用于限制cyclone-worker，默认是0.5      |
| RECORD_ROTATION_THRESHOLD | 流水线执行记录保留数，默认是50      |
| CALLBACK_URL              | webhook回调地址，默认是http://127.0.0.1:7099/v1/pipelines       |
| CYCLONE_SERVER            | Cyclone-Server的访问地址，默认是http://localhost:7099 |
| WORKER_IMAGE              | Cyclone-Worker容器的镜像名，默认是cargo.caicloud.io/caicloud/cyclone-worker:latest |
|NOTIFICATION_URL|通知URL，如果流水线配置了通知策略，流水线执行完会发送通知到改URL|
|RECORD_WEB_URL_TEMPLATE|客户自定义的流水线执行记录前端页面URL地址模板|

PS:
### 关于 `RECORD_WEB_URL_TEMPLATE`
使用 golang text/template 风格, 将 `Event` 作为text/template 的 `Execute` 方法的入参，所以你可以在模板里定义 `{{.Pipeline.Name}}` , `{{.PipelineRecord.ID}}` , `{{.Project.Name}}` 等参数。

`.` 表示 [`Event`](#EventObject) 结构体, `Pipeline` `PipelineRecord` `Project` 是`Event`结构体的成员。

例如：

如果
- `RECORD_WEB_URL_TEMPLATE`=http://127.0.0.1:30000/devops/projects/{{.Project.Name}}/pipelines/{{.Pipeline.Name}}/records/{{.PipelineRecord.ID}}

并且
- `event.Pipeline.Name`=project-test-1,
- `event.Pipeline.Name`=pipeline-1,
- `event.PipelineRecord.ID`=5b98850a1d74bd0001c17dcf,

那么targetURL替换后的结果将是:
http://127.0.0.1:30000/devops/projects/project-test-1/pipelines/pipeline-1/records/5b98850a1d74bd0001c17dcf|

#### EventObject
```
type Event struct {
	Project        *Project
	Pipeline       *Pipeline
	PipelineRecord *PipelineRecord
}
```
