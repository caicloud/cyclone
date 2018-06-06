<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick Start](#quick-start)
  - [Overview](#overview)
  - [Register a cloud](#register-a-cloud)
  - [Create a project](#create-a-project)
  - [Create a pipeline](#create-a-pipeline)
  - [Create a pipeline record](#create-a-pipeline-record)
  - [Check out logs](#check-out-logs)
    - [At the end of the pipeline run](#at-the-end-of-the-pipeline-run)
    - [While the pipeline is running](#while-the-pipeline-is-running)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Quick Start

## Overview

 Once you have [set up Cyclone](), you have a CI/CD platform to ship your code from local development all the way to container engine. This section provides a brief overview of Cyclone operational processes.

 ## Register a cloud

 Cyclone runs pipelines in cyclone-worker container,
 In order to build your pipeline, a cloud where cyclone-worker container runs on must be registered to cyclone-server.


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

 The script above would create a cloud at your local computer, then you could set up the project.

 ## Create a project

 A project represents a group to manage a set of related pipelines. project must has a credential to access your VCS repository, and now Cyclone supports three types of VCS: GitHub, GitLab and SVN.

 you can create a project by typing these following instructions:

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
		"password": "'PASSWORD'"
	}
 }' ''$HOST_URL'/api/v1/projects'
 ```

 ## Create a pipeline

 Pipeline is a workflow contained by a series of stages which make the CI/CD work. There are five stages for you to configure your pipeline, and we will introduce the required two of them, which are CodeCheckout and Package, and you can refer to [Pipelines  APIs]() to get more detailed infomations.

 ```
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
 That means checkout code from `$YOUR_GITHUB_REPO_ADDRESS` firstly, and then running a container which used busbox image to perform the commands we defined, which are `echo hello` and `echo word`.

 ## Create a pipeline record

After the pipeline is created successfully, you could create a recod to perfom it.

```shell
# create a record.
 curl -X POST -H "Content-Type: application/json" -d '{
	"name": "v0.0.1",
	"description": "first pipeline record",
	"stages": ["codeCheckout", "package"]
 }' ''$HOST_URL'/api/v1/projects/'$PROJECT_NAME'/pipelines/'$PIPELINE_NAME'/records'
```

When creating the record, you could choose stages you have defined in pipeline to perform, but `codeCheckout` and `package` are required. If you have defined the optional stages, which are `imageBuild` `integrationTest` and `imageRelease`, of course you can choose them as well.

## Check out logs

After the record is created successfully, you could get the log generated when the version is being created.

### At the end of the pipeline run
At the end of the pipeline run successfully, you could get the logs by a HTTP GET request.

```shell
# get record logs.
 curl -v -S ''$HOST_URL'/api/v1/projects/'PROJECT_NAME'/pipelines/'$PIPELINE_NAME'/records/'$RECORD_ID'/logs?stage=package'
```

The log would looks like:
```
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

### While the pipeline is running

If you want to get logs while the pipeline is running, please typing following instructions to send a websocket connection to cyclone-server:

```
# get the real time record log stream.
 curl -v -S --no-buffer \
     -H "Connection: Upgrade" \
     -H "Upgrade: websocket" \
     -H "Host: localhost:8080" \
     -H "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     -H "Sec-WebSocket-Version: 13" \
 ''$HOST_URL'/api/v1/projects/'PROJECT_NAME'/pipelines/'$PIPELINE_NAME'/records/'$RECORD_ID'/logstream?stage=package'
```
