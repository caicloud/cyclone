# Caicloud.yml Reference

## Pre Build

Pre build container would run the commands and copy out outputs.

```yml
pre_build:
  context_dir: publish
  dockerfile_name: Dockerfile_prebuild
  image: golang:v1.5.3
    volumes:
      - .:/root
    environment:
      - key = value
    commands:
      - ls
      - pwd
    outputs:
        - file1
        - dir2
```

## Build

You must define context\_dir or dockerfile\_name to make Cyclone run build step from the configuration file.

```yml
build:
  context_dir: publish
  dockerfile_name: Dockerfile_publish
```

Cyclone would build a docker image from `<root_dir>/<context_dir>/<dockerfile_name>`, and the working directory is `<context_dir>`. If the image is built successfully, Cyclone would push it to the docker registry defined before running.

## Integration

Cyclone supports launching separate, ephemeral Docker containers as part of the integration process. This is useful, for example, if you require a database for running your tests.

The service conatienrs and the integration container are in one network mode, so integration container could resolve the service names to the service containers. If you want to communicate with these containers, just use their names and the ports as endpoints. **If the image of service configurated as "BUILT_IMAGE", it will use the newly built image.**

```yml
integration:
  services:
    postgres:
      image: postgres:9.4.5
      enviroment:
        - key = value
      commands:
        - cmd1
        - cmd2
    app:
      image: BUILT_IMAGE
      enviroment:
        - key = value
      commands:
        - cmd1
        - cmd2
  environment:
    - key = value
  commands:
    - ls
    - pwd
    - ping postgress
```

## Post Build

Cyclone supports post build hook to complete the build process. Post build hook would be run in a isolated container, the root directory is mounted into the container as a volume. We suggest users to add necessary binaries to image, and implement the post build logic in `commands` field.

```yml
post_build:
  image: golang:v1.5.3
  environment:
    - key = value
  commands:
    - ls
    - pwd
```

## Depoly

Cyclone supports deploy the newest image for applications. It also support deploy the image to [Caicloud Cubernetes](https://caicloud.io/products/cubernetes) and [Google Kubernetes](http://kubernetes.io/).

#### Caicloud Cubernetes

```yml
#deploy
deploy:
  - deployment: redis-master
    cluster: cluster1_id
    namespace: namespace1_id
    containers:
      - container1
      - container2
  - deployment: redis-slave
    cluster: cluster2_id
    namespace: namespace2_id
    containers:
      - container1
      - container2
```

#### Google Kubernetes

```yml
#deploy
deploy:
  - type: kubernetes
    host: <cluster host>
    token: <cluster access token>
    cluster: cluster1_id
    namespace: namespace1_id
    deployment: redis-master
    containers:
      - container1
      - container2
  - type: kubernetes
    host: <cluster host>
    token: <cluster access token>
    cluster: cluster2_id
    namespace: namespace2_id
    deployment: redis-slave
    containers:
      - container1
      - container2
```
