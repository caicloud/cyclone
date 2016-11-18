# Setup

## Prerequisites

```
golang：1.6+
docker：(Suggested)1.10.0+(Has tested and verified with 1.10~1.11)
kubernetes：1.2+
```

## Setup Instruction

There are two ways to set up Cyclone, either via Docker Compose or kubectl.

At first, you need to build Cyclone-server and Cyclone-worker docker images, by following these instructions: 

```
git clone https://github.com/caicloud/cyclone
cd cyclone
./scripts/build-image.sh
./scripts/build-worker-image.sh
```

- Using Docker Compose: Follow these instructions to bring up Cyclone with docker-compose (you can checkout the compose file for more details):

```
docker-compose -f docker-compose.yml up -d
```

Then Cyclone is started. Docker compose will start Clair before Postgres which may raise an error. If this error is raised, manually execute `docker start clair_clair`.

- Using Kubectl: this approach requires a few yaml files. You can read related files in cirlce/scripts/k8s for more detial. Follow these instructions:

```
git clone https://github.com/caicloud/cirlce
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

Then Cyclone is started.

## Other

Environment variables: 

| ENV                    | Description                              |
| ---------------------- | ---------------------------------------- |
| MONGO_DB_IP            | The IP of mongodb, default is localhost. |
| KAFKA_SERVER_IP        | The address of kafka, default is 127.0.0.1:9092. |
| WORK_REGISTRY_LOCATION | The registry to push images, default is cargo.caicloud.io. |
| REGISTRY_USERNAME      | The username in docker registry, default is null. |
| REGISTRY_PASSWORD      | The password in docker registry, default is null. |
| CLIENTID               | The client ID from Github for oauth, default is null. |
| CLIENTIDSECRET         | The client secret from Github for oauth, default is null. |
| CONSOLE_WEB_ENDPOINT   | The address of caicloud access point, default is http://localhost:8000. |
| CLIENTID_GITLAB        | The client ID from Gitlab for oauth, default is null. |
| CLIENTIDSECRET_GITLAB  | The client secret from Gitlab for oauth, default is null. |
| SERVER_GITLAB          | The address of gitlab, default is https://gitlab.com. |
| ETCD_SERVER_IP         | The address of etcd, default is 127.0.0.1:2379. |
| CYCLONE_SERVER_HOST    | The host of Cyclone-Server, default is http://localhost:7099. |
| WORKER_IMAGE           | The image name of Cyclone-Worker container, default is cargo.caicloud.io/caicloud/cyclone-worker:latest. |
| CLAIR_SERVER_IP        | The address of clair, default is 127.0.0.1:6060. |
