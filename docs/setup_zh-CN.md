# 安装

## 版本依赖

```
golang：1.6+
docker：建议使用 1.10.1
kubernetes：1.2+
```

## 介绍

这篇文档会对 Cyclone 的安装进行简单的介绍。这里有两种方式，一种是使用 docker compose，一种是使用 kubectl。

在此之前，你需要先编译生成 Cyclone-server 和 Cyclone-worker 镜像，操作命令如下：

```
git clone https://github.com/caicloud/cyclone
cd cyclone
./scripts/build-image.sh
./scripts/build-worker-image.sh
```

- 使用 docker compose 方式。 你可以去查看 docker-compose.yml 来了解运行过程。具体的操作指令如下：

```
docker-compose -f docker-compose.yml up -d
```

这样 Cyclone 已经启动了。使用 docker compose 的方式来启动时，clair 可能会在 postgres 之前启动，这样会出现错误，因为 clair 是依赖于 postgres。如果出现这种错误，需要手动执行 `docker start clair_clair`。

- 使用 kubectl 方式，需要一些 YAML 文件。你可以在 cirlce/scripts/k8s 目录下查看这些文件。然后执行下面这些命令，

```
git clone https://github.com/caicloud/cyclone
cd cyclone/scripts/k8s
kubectl create namespace cyclone
kubectl --namespace=cyclone create -f zookeeper.yaml
kubectl --namespace=cyclone create -f zookeeper-svc.yaml
kubectl --namespace=cyclone create -f kafka.yaml
kubectl --namespace=cyclone create -f kafka-svc.yaml
kubectl --namespace=cyclone create -f mongo.yaml
kubectl --namespace=cyclone create -f mongo-svc.yaml
kubectl --namespace=cyclone create -f etcd.yaml
kubectl --namespace=cyclone create -f etcd-svc.yaml
kubectl --namespace=cyclone create secret generic clairsecret --from-file=config.yaml
kubectl --namespace=cyclone create -f postgres.yaml
kubectl --namespace=cyclone create -f postgres-svc.yaml
kubectl --namespace=cyclone create -f clair.yaml
kubectl --namespace=cyclone create -f clair-svc.yaml
kubectl --namespace=cyclone create -f cyclone.yaml
kubectl --namespace=cyclone create -f cyclone-svc.yaml
```

这样 Cyclone 就启动了。

## 其它

环境变量解释：

```
MONGO_DB_IP             mongo db的地址, 默认是localhost
KAFKA_SERVER_IP         kafka 服务的地址, 默认是127.0.0.1:9092
LOG_SERVER              日志服务的地址, 默认是127.0.0.1:8000
WORK_REGISTRY_LOCATION  镜像仓地址, 默认是cargo.caicloud.io
REGISTRY_USERNAME       镜像仓用户名, 默认是null
REGISTRY_PASSWORD       镜像仓用户密码, 默认是null
CLIENTID                用于github oauth授权的clienID, 默认是null
CLIENTIDSECRET          用于github oauth授权的secret, 默认是null
CONSOLE_WEB_ENDPOINT    caicloud网页访问地址, 默认是http://localhost:8000
CLIENTID_GITLAB         用于gitlab oauth授权的clienID, 默认是null
CLIENTIDSECRET_GITLAB   用于gitlab oauth授权的secret, 默认是null
SERVER_GITLAB           gitlab的服务器地址, 默认是https://gitlab.com
ETCD_SERVER_IP          etcd的服务器地址, 默认是127.0.0.1:2379
CYCLONE_SERVER_HOST      cirlce的访问地址, 默认是http://localhost:709
WORKER_IMAGE            worker容器的镜像名称，默认是cargo.caicloud.io/caicloud/cyclone-worker
CLAIR_SERVER_IP         clair的服务器地址, 默认是127.0.0.1:6060
```
