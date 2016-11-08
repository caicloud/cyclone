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

You could define the directory and dockerfile's name in build section. If you don't, Cyclone would use the root directory and `Dockerfile` by default. If the image is built successfully, Cyclone would push it to the docker registry.

```yml
build:
  context_dir: publish
  dockerfile_name: Dockerfile_publish
```

### Integration

Cyclone use the image  which built during the build step to run a container . Then Cyclone would run integration step defined in `caicloud.yml` as a docker container. If the integration is failed, version creation fails.

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

Cyclone supports post build hook to complete the build process. In some cases, you could do some clean-ups or have some joint products to publish. So the post build hook could handle these cases by adding a new step to build process, and is run after the build task is over.

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
  - deployment: mongo-server
    cluster: 1e73520d-f7ab-4998-b169-41b8c122342b
    namespace: test
    containers:
      - mongo-server
```

#### Google Kubernetes

```yml
deploy:
  - type: kubernetes
    host: <cluster host>
    token: <cluster access token>
    cluster: <cluster name>
    namespace: <namespace name>
    deployment: <deployment name>
    containers:
      - mongo-server
```

Cyclone also supports deploying a few applications into clusters one by one. You only need to write the info in "deploy" section, as the above do.

See [caicloud.yml Reference](./caicloud-yml-reference.md) for more information.
