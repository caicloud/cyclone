# Setup 

## Describe

There are two ways to set up Cyclone, the one is using Docker Compose, the another is using Kubectl. 

At first, you need build Cyclone-server and Cyclone-worker docker image, do the following these instructions,
```
	git clone https://github.com/caicloud/cyclone
	cd cyclone
	./scripts/build-image.sh
	./scripts/build-worker-image.sh
```

- Using Docker Compose，it need docker-compose.yaml file. You can read docker－compose.yaml file to learn. The detail instructions，
```
	git clone https://github.com/caicloud/cyclone
	cd cyclone
	docker-compose -f  docker－compose.yaml up -d
```
Then Cyclone is started. Docker compose may start Clair before Postgres which will raise an error. If this error is raised, manually execute ```docker start clair_clair```.

- Using Kubectl，it need a few yaml files. You can read related files in cirlce/scripts/k8s to learn. Then do the following these instructions,
```
	git clone https://github.com/caicloud/cirlce
	cd cyclone/scripts/k8s
	kubectl create -f zookeeper.yaml
	kubectl create -f zookeeper-svc.yaml
	kubectl create -f kafka.yaml
	kubectl create -f kafka-svc.yaml
	kubectl create -f mongo.yaml
	kubectl create -f mongo-svc.yaml
	kubectl create -f etcd.yaml
	kubectl create -f etcd-svc.yaml
	kubectl create secret generic clairsecret --from-file=config.yaml
	kubectl create -f postgres.yaml
	kubectl create -f postgres-svc.yaml
	kubectl create -f clair.yaml
	kubectl create -f clair-svc.yaml
	kubectl create -f cyclone.yaml
	kubectl create -f cyclone-svc.yaml
```
Then Cyclone is started.

## Other

Describe environment variables：
- MONGO_DB_IP             The IP of mongodb, default is localhost
- KAFKA_SERVER_IP         The address of kafka, default to 127.0.0.1:9092
- LOG_SERVER              The address of log server, default to 127.0.0.1:8000
- WORK_REGISTRY_LOCATION  The registry to push image, default is index.caicloud.io
- REGISTRY_USERNAME       The username in docker registry, default is null
- REGISTRY_PASSWORD       The password in docker registry, default is null
- CLIENTID                The client ID from Github to for oauth, default is null
- CLIENTIDSECRET          The client secret from Github to for oauth, default is null
- CONSOLE_WEB_ENDPOINT    The address of caicloud access point, default to http://localhost:8000
- CLIENTID_GITLAB         The client ID from Gitlab to for oauth, default is null
- CLIENTIDSECRET_GITLAB   The client secret from Gitlab to for oauth, default is null
- SERVER_GITLAB           The address of gitlab, default to https://gitlab.com
- ETCD_SERVER_IP          The address of etcd, default to 127.0.0.1:2379
- CYCLONE_SERVER_HOST     The address where cyclone is running, default to http://localhost:709
- WORKER_IMAGE            The image name for worker container, default to index.caicloud.io/caicloud/cyclone-worker
- CLAIR_SERVER_IP         The address of clair, default to 127.0.0.1:6060
