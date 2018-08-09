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

- Using Kubectl: this approach requires a few yaml files. You can read related files in cyclone/scripts/k8s for more details, and adjust the parameters before executing these instructions:

```
git clone https://github.com/caicloud/cyclone
cd cyclone/scripts/k8s
kubectl create namespace cyclone
kubectl --namespace=cyclone create -f mongo.yaml
kubectl --namespace=cyclone create -f mongo-svc.yaml
kubectl --namespace=cyclone create -f cyclone.yaml
kubectl --namespace=cyclone create -f cyclone-svc.yaml
```

Then Cyclone is started.

## Other

Environment variables:

| ENV                       | Description                              |
| ------------------------- | ---------------------------------------- |
| MONGODB_HOST              | The IP of mongodb, default is localhost. |
| REGISTRY_LOCATION         | The registry to push images, default is cargo.caicloud.io. |
| REGISTRY_USERNAME         | The default username for docker registry, default is null. |
| REGISTRY_PASSWORD         | The default password for docker registry, default is null. |
| LIMIT_MEMORY              | Same concept as [kubernetes limits.memory](https://kubernetes.io/docs/concepts/policy/resource-quotas), used for cyclone-worker, default is 1Gi     |
| LIMIT_CPU                 | Same concept as [kubernetes limits.cpu](https://kubernetes.io/docs/concepts/policy/resource-quotas), used for cyclone-worker, default is1           |
| REQUEST_MEMORY            | Same concept as [kubernetes requests.memory](https://kubernetes.io/docs/concepts/policy/resource-quotas), used for cyclone-worker, default is 0.5Gi |
| REQUEST_CPU               | Same concept as [kubernetes requests.cpu](https://kubernetes.io/docs/concepts/policy/resource-quotas), used for cyclone-worker, default is 0.5      |
| RECORD_ROTATION_THRESHOLD | The number of pipeline records cyclone preserved, default is 50      |
| CALLBACK_URL              | The URL used for webhook to callback, default is http://127.0.0.1:7099/v1/pipelines       |
| CYCLONE_SERVER            | The host of Cyclone-Server, default is http://localhost:7099. |
| WORKER_IMAGE              | The image name of Cyclone-Worker container, default is cargo.caicloud.io/caicloud/cyclone-worker:latest. |
