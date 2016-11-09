# Quick Start

* [Overview](#overview)
* [Node Registration](#node-registration)
* [Setting up a service](#setting-up-a-service)
* [Create a version](#create-a-version)
* [Check log file](#check-log-file)

## Overview

In order to configure your build, you must register a node to Cyclone server, and include a caicloud.yml file in the root of your repository. This section provides a brief overview of the configuration file and build process.

## Node Registration

Cyclone runs your builds in worker nodes. After Cyclone server starts, at least one worker node should be registered.

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

The script above would create a worker node at your local computer, then you could set up the project.

## Setting up a service

A service represents a VCS repository, you need to create a new service, Then Cyclone would run your pre-defined tasks automatically.

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

After you created the service, then Cyclone would run your CI or CD tasks automatically, or you can trigger these tasks manually.

## Create a version

If you include a `caicloud.yml` file in the repository's root directory, Cyclone would create a version from the configuration file, otherwise Cyclone would follow default processes to complete version creation.

```shell
# Create a version manually.
curl -sS -X POST -H "Content-Type:application/json" -d '{
  "name": "v0.1.0",
  "description": "just for test",
  "service_id": "'$SERVICE_ID'"
}' ''$HOST_URL'/api/v0.1/'$USER_ID'/versions'

```

In order to configure your docker image you must include a `Dockerfile` file in the root of your repository. In advance section we would show how to custom the name of `Dockerfile` in configuration file.

Now you have a `Dockerfile`, and version creation would run `docker build` and `docker push`, these operations proceed sequentially.

When `caicloud.yml` is given, version creation would include what you defined in the configuration file.

## Check log file

After the version is created successfully, you could get the log generated when the version is being created.

If version creation is running, you could get log by a websocket connection, And if version creation is over, you would get logs by a HTTP GET request.

```shell
# Get log from Cyclone server.
curl -v -sS ''$HOST_URL'/api/v0.1/'$USER_ID'/versions/'$VERSION_ID'/logs'

```

The log would looks like:

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

Cyclone would build and publish a new version of docker image to custom docker registry, if the version creation is completed successfully.
