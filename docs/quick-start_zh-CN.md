<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [快速开始](#%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B)
  - [介绍](#%E4%BB%8B%E7%BB%8D)
  - [容器云注册](#%E5%AE%B9%E5%99%A8%E4%BA%91%E6%B3%A8%E5%86%8C)
  - [创建项目](#%E5%88%9B%E5%BB%BA%E9%A1%B9%E7%9B%AE)
  - [创建流水线](#%E5%88%9B%E5%BB%BA%E6%B5%81%E6%B0%B4%E7%BA%BF)
  - [查看日志](#%E6%9F%A5%E7%9C%8B%E6%97%A5%E5%BF%97)
    - [查看运行中日志](#%E6%9F%A5%E7%9C%8B%E8%BF%90%E8%A1%8C%E4%B8%AD%E6%97%A5%E5%BF%97)
    - [查看运行完日志](#%E6%9F%A5%E7%9C%8B%E8%BF%90%E8%A1%8C%E5%AE%8C%E6%97%A5%E5%BF%97)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# 快速开始

## 介绍

Cyclone [安装](./setup_zh-CN.md)成功后，
这篇文档会对 Cyclone 的使用进行简单的介绍。

首先需要向 Cyclone server 注册一个运行worker的容器云。随后可以通过 RESTful API 来与 Cyclone server 进行交互，完成项目的持续集成与持续部署。

## 容器云注册

Cyclone 是 Master/Slave 的架构，因此为了能够在 worker 节点上运行构建任务，需要先向 Cyclone server 注册不少于一个的容器云，容器云用来启动运行cyclone-worker。

```shell
# create a cloud.
curl -X POST -H "Content-Type: application/json" -d '{
       "name": "myTestCloud",
       "type": "Docker",
       "docker": {
           "host": "'$DOCKER_HOST'"
       }
}' ''$HOST_URL'/api/v1/clouds'
```

上述命令会在本地计算机上，创建一个docker类型的容器云。接下来就可以去创建需要进行持续集成与持续部署的项目。

## 创建项目

一个 project（项目）可以管理基于同一个源码管理系统（SCM,Source Code Management）中的repo（代码库）所创建的一系列流水线，project必须包含SCM的认证信息。Cyclone支持三种SCM：GitHub，GitLab和SVN。

```shell
# create a project.
curl -X POST -H "Content-Type: application/json" -d '{
  "name": "myProject",
  "description": "first test project",
  "owner": "cyclone",
  "scm": {
    "type": "Github",
    "server": "https://github.com/",
    "authType": "Password",
    "username": "'$USERNAME'",
    "password": "'$PASSWORD'"
  }
}' ''$HOST_URL'/api/v1/projects'
```

## 创建流水线

Pipeline（流水线）是真正定义CI/CD如何运行的工作流。Cyclone支持5个阶段的配置来定义流水线，
一个必选阶段：
 - 检出代码（codeCheckout）

四个可选:
 - 构建代码（package）
 - 镜像构建（imageBuild）
 - 集成测试（integrationTest）
 - 发布仓库（imageRelease）

我们将创建一个包含`检出代码`和`构建代码`阶段的流水线作为示例，更多信息请参考[Pipeline APIs](./api/v1/api.md#pipeline-apis)。
```shell
# create a pipeline including CodeCheckout and Package.
curl -X POST -H "Content-Type: application/json" -d '{
  "name": "myPipeline",
  "description": "first test pipeline",
  "owner": "cyclone",
  "build": {
    "builderImage": {
      "image": "busybox"
    },
    "stages": {
      "codeCheckout": {
        "mainRepo": {
          "type": "Github",
          "github": {
            "url": "'$YOUR_GITHUB_REPO_ADDRESS'",
            "ref": "'$YOUR_BRANCH'"
          }
        }
      },
      "package": {
        "command": [
          "'$YOUR_COMMAND_1'",
          "'$YOUR_COMMAND_2'"
        ]
      }
    }
  }
}' ''$HOST_URL'/api/v1/projects/'$PROJECT_NAME'/pipelines'
```

上述命令表示创建一条流水线，先基于`$YOUR_GITHUB_REPO_ADDRESS`检出代码，然后启动一个以`busybox`为对镜像的容器，在容器中执行`YOUR_COMMAND_1`和`$YOUR_COMMAND_2`来构建代码。

## 查看日志

在版本创建过程中和结束后，可以通过 API 请求得到构建过程的日志。

### 查看运行中日志
如果版本仍然在构建，可以通过 websocket 请求得到流水线执行各阶段的日志：

```
# get package stage record log stream in the real time.
 curl -v -S --no-buffer \
     -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Host: localhost:8080" \
     -H "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     -H "Sec-WebSocket-Version: 13" \
 ''$HOST_URL'/api/v1/projects/'PROJECT_NAME'/pipelines/'$PIPELINE_NAME'/records/'$RECORD_ID'/logstream?stage=package'
```

### 查看运行完日志

在流水线执行结束后，可以通过 HTTP GET 请求得到日志。

```shell
# get package stage record logs.
 curl -v -S ''$HOST_URL'/api/v1/projects/'PROJECT_NAME'/pipelines/'$PIPELINE_NAME'/records/'$RECORD_ID'/logs?stage=package'
```

得到的日志会是如下格式：

```text
Stage: Package status: start
$ echo $GOPATH
/go
$ export REPO_NAME=cyclone
$ export WORKDIR=$GOPATH/src/github.com/caicloud/$REPO_NAME
$ mkdir -p $GOPATH/src/github.com/caicloud/
$ ln -s `pwd` $WORKDIR
$ cd $WORKDIR
$ ls -la
total 88
drwxr-xr-x   17 root     root           620 May 29 03:49 .
drwxrwxrwt    3 root     root            18 May 29 04:00 ..
-rw-r--r--    1 root     root            12 May 29 03:49 .dockerignore
drwxr-xr-x    8 root     root           280 May 29 03:49 .git
drwxr-xr-x    2 root     root            80 May 29 03:49 .github
-rw-r--r--    1 root     root           374 May 29 03:49 .gitignore
-rw-r--r--    1 root     root          1338 May 29 03:49 .travis.yml
-rw-r--r--    1 root     root         16843 May 29 03:49 Gopkg.lock
-rw-r--r--    1 root     root          2716 May 29 03:49 Gopkg.toml
-rw-r--r--    1 root     root          7960 May 29 03:49 Jenkinsfile
-rw-r--r--    1 root     root         11357 May 29 03:49 LICENSE
-rw-r--r--    1 root     root          5370 May 29 03:49 Makefile
-rw-r--r--    1 root     root           151 May 29 03:49 OWNERS
-rw-r--r--    1 root     root          3285 May 29 03:49 README.md
-rw-r--r--    1 root     root          3363 May 29 03:49 README_zh-CN.md
drwxr-xr-x    4 root     root            80 May 29 03:49 api
drwxr-xr-x    4 root     root            80 May 29 03:49 build
-rw-r--r--    1 root     root          2140 May 29 03:49 caicloud_e2e-test.yml
-rw-r--r--    1 root     root           314 May 29 03:49 caicloud_unit-test.yml
drwxr-xr-x    2 root     root           200 May 29 03:49 cloud
drwxr-xr-x    4 root     root            80 May 29 03:49 cmd
-rw-r--r--    1 root     root          1313 May 29 03:49 docker-compose.yml
drwxr-xr-x    3 root     root           520 May 29 03:49 docs
drwxr-xr-x    3 root     root           100 May 29 03:49 http
drwxr-xr-x    3 root     root            60 May 29 03:49 node_modules
drwxr-xr-x    3 root     root           120 May 29 03:49 notify
drwxr-xr-x   17 root     root           340 May 29 03:49 pkg
drwxr-xr-x    8 root     root           280 May 29 03:49 scripts
drwxr-xr-x    8 root     root           180 May 29 03:49 tests
drwxr-xr-x    2 root     root            80 May 29 03:49 utils
drwxr-xr-x    7 root     root           140 May 29 03:49 vendor
$ cd $WORKDIR
$ make build-local
github.com/caicloud/cyclone/pkg/register
github.com/caicloud/cyclone/vendor/github.com/docker/docker/api/types/blkiodev
...
...
github.com/caicloud/cyclone/pkg/worker
github.com/caicloud/cyclone/cmd/worker
Stage: Package status: finish
```
