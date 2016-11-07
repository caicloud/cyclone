English | [中文](README_zh-CN.md)
# Cyclone
![logo](docs/logo.jpeg)

[![GoDoc](https://godoc.org/github.com/caicloud/cyclone?status.svg)](https://godoc.org/github.com/caicloud/cyclone)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/cyclone)](https://goreportcard.com/report/github.com/caicloud/cyclone)


Cyclone is a cloud native CI/CD platform built for container workflow.

The primary directive of cyclone is to ship code from local development all the way to container engine of choice, either running in test or production environment. Features of cyclone includes:

- **Container Native**: every build, integration and deployment runs in container, completely excludes inconsistency between runtime environment
- **Dependency Aware**: define dependency rules, or simply component relationship, cyclone takes care of execution order as well as rollout strategy
- **Version Control**: cyclone is built with version control in mind; retrieving image/pipeline history is as simple as querying its version management interface
- **Two-way Binding**: cyclone records every CI/CD operation and its effect to answer questions like "how various container images are deployed across the fleet?"
- **Security First**: security is an essential part of cyclone; barriers can be setup to prevent insecure images from launching into production

## Documentation
* [Setup Guide](./docs/setup.md)
* [Quick Start](./docs/quick-start.md)
* [caicloud.yml Reference](./docs/caicloud-yml-reference.md)
* [API Guide](http://118.193.142.27:7099/apidocs/)
* [Principle](./docs/principle.md)

## Caicloud.yml Introduction

When `caicloud.yml` is given, version creation would include what you defined in the configuration file.

Example `caicloud.yml` configuration:

```yml
integration:
  image: golang:1.5
  environment:
    - GO15VENDOREXPERIMENT=1
    - GOOS=linux
    - GOARCH=amd64
    - CGO_ENABLED=0
  commands:
    - go run main.go
```

The configuration file has at most six sections: integration, pre\_build, build, post_build, and deploy. And all of the six sections are optional, you could choose some of them to construct your version creation process.

### Pre Build

It's not elegent to leave the code in docker image, so you could compile the source code, and copy resulting artifacts out of the build container, then build a new image which contains only artifacts.

```yml
pre_build:
  image: golang:1.6
  volumes:
    - .:/go/src/github.com/caicloud/ci-demo-go
  commands:
    - echo "compile executable files"
    - cd /go/src/github.com/caicloud/ci-demo-go/code
    - go build -v -o app
  outputs: # copy out publish executable files from prebuild container
    - /go/src/github.com/caicloud/ci-demo-go/code/app
```

### Build

You could define the directory and dockerfile's name in build section. If you don't, Cyclone uses the root directory and `Dockerfile` by default. If the image is built successfully, Cyclone would push it to the docker registry.

```yml
build:
  context_dir: publish
  dockerfile_name: Dockerfile_publish
```

### Integration

Cyclone use the image  which built durning the build step to run a container . Then Cyclone would run integration step defined in `caicloud.yml` as a docker container. If the integration is down, version creation fails.

```yml
integration:
  environment:
    - GO15VENDOREXPERIMENT=1
    - GOOS=linux
    - GOARCH=amd64
    - CGO_ENABLED=0
  commands:
    - go run main.go
```

### Post Build

Cyclone supports post build hook to complete the build process. In some cases, you could do some clean-ups or have some joint products to publish. So the post build hook could handle these cases by add a new step to build process, and is run after the build task is over.

```yml
post_build:
  image: golang:1.6
  commands:
    - echo "Built Successfully."
```

### Deploy

Cyclone supports to deploy the image to [Caicloud Cubernetes](https://caicloud.io/products/cubernetes) and [Google Kubernetes](http://kubernetes.io/).

#### Caicloud Cubernetes

```yml
deploy:
  - application: mongo-server
    cluster: 1e73520d-f7ab-4998-b169-41b8c122342b
    partition: test
    containers:
      - mongo-server
```

#### Google Kubernetes

```yml
deploy:
  - type: kubernetes 
    host: <cluster host>
    token: <cluster access token>
    application: mongo-server
    cluster: 1e73520d-f7ab-4998-b169-41b8c122342b
    partition: test
    containers:
      - mongo-server
```

Cirlce also supports deploy a few applications into clusters one by one. You need only write the info in "deploy" section, as the above do.
