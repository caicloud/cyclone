# Setup

## Version Dependence

```
golang：1.6+
docker：(Suggested)1.10.1
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

```
MONGO_DB_IP             The IP of mongodb, default is localhost
KAFKA_SERVER_IP         The address of kafka, defaults to 127.0.0.1:9092
LOG_SERVER              The address of log server, defaults to 127.0.0.1:8000
WORK_REGISTRY_LOCATION  The registry to push images, default is cargo.caicloud.io
REGISTRY_USERNAME       The username in docker registry, default is null
REGISTRY_PASSWORD       The password in docker registry, default is null
CLIENTID                The client ID from Github for oauth, default is null
CLIENTIDSECRET          The client secret from Github for oauth, default is null
CONSOLE_WEB_ENDPOINT    The address of caicloud access point, defaults to http://localhost:8000
CLIENTID_GITLAB         The client ID from Gitlab for oauth, default is null
CLIENTIDSECRET_GITLAB   The client secret from Gitlab for oauth, default is null
SERVER_GITLAB           The address of gitlab, defaults to https://gitlab.com
ETCD_SERVER_IP          The address of etcd, defaults to 127.0.0.1:2379
CYCLONE_SERVER_HOST     The address where cyclone is running, defaults to http://localhost:709
WORKER_IMAGE            The image name for worker container, defaults to cargo.caicloud.io/caicloud/cyclone-worker
CLAIR_SERVER_IP         The address of clair, defaults to 127.0.0.1:6060
```
