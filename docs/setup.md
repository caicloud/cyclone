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
./scripts/build-server.sh
./scripts/build-worker.sh
```

- Using Docker Compose: Follow these instructions to bring up Cyclone with docker-compose (you can checkout the compose file for more details):

```
docker-compose -f docker-compose.yml up -d
```

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
kubectl --namespace=cyclone create -f cyclone.yaml
kubectl --namespace=cyclone create -f cyclone-svc.yaml
```

Then Cyclone is started.

## Other

Environment variables: 

| ENV                  | Description                              |
| -------------------- | ---------------------------------------- |
| MONGODB_HOST         | The IP of mongodb, default is localhost. |
| KAFKA_HOST           | The address of kafka, default is 127.0.0.1:9092. |
| REGISTRY_LOCATION    | The registry to push images, default is cargo.caicloud.io. |
| REGISTRY_USERNAME    | The username in docker registry, default is null. |
| REGISTRY_PASSWORD    | The password in docker registry, default is null. |
| GITHUB_CLIENT        | The client ID from Github for oauth, default is null. |
| GITHUB_SECRET        | The client secret from Github for oauth, default is null. |
| GITLAB_CLIENT        | The client ID from Gitlab for oauth, default is null. |
| GITLAB_SECRET        | The client secret from Gitlab for oauth, default is null. |
| GITLAB_URL           | The address of gitlab, default is https://gitlab.com. |
| CYCLONE_SERVER       | The host of Cyclone-Server, default is http://localhost:7099. |
| WORKER_IMAGE         | The image name of Cyclone-Worker container, default is cargo.caicloud.io/caicloud/cyclone-worker:latest. |
