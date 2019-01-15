# Contributing to Cyclone

Welcome to Cyclone! This guide will provide information on how to contribute to Cyclone. It covers how to build Cyclone, how to commit and what conventions and practices contributors should follow.

If you have any problems about Cyclone, welcome to join [Cyclone Slack](https://caicloud-cyclone.slack.com) to discuss.

## Get Start

### Fork Repository

Fork Cyclone repository to your own Github account, and then:
```
$ mkdir -p $GOPATH/src/github.com/caicloud
$ go get github.com/caicloud/cyclone

# Avoid push to caicloud/cyclone
$ git config push.default nothing

$ git remove add $USER git@github.com:$USER/cyclone.git
$ git fetch $USER
```

### Project Structure

```
.
...
├── Makefile  # Make file to build and package
├── build     # Contains files to build images, e.g. Dockerfiles
├── cmd       # Entrypoints (main.go) of components
├── manifests # Kubernetes manifests for deployment and examples
├── pkg       # Source code for Cyclone server, workflow engine
├── tools     # Tools used by Cyclone, like Kubernetes client generator
└── web       # Web UI source code for Cyclone
```

### Build & Start

About how to build and start Cyclone, please refer to [Build & Run](./docs/build-guide.md).

## Contribute Workflow

Checkout a feature branch
```
$ git checkout -b my_feature
```

Add your commits,
```
$ git commit -a -s -m "Implement an awesome feature"
```

Push to your own repository
```
$ git push $USER my_feature
```

Create a merge request in Github and wait for review.

For a more detailed workflow, you can refer to [Kubernetes Github Workflow](https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md)

## Code Styles

### Golang

Cyclone is written with Golang, before you commit your code, please check [Golang Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

Cyclone uses [Gometalinter](https://github.com/alecthomas/gometalinter) to statically check source code for errors and warnings. Please ensure your commits have passed the check before creating PR.

```bash
$ make lint
``` 

### Javascript

Cyclone's web UI is developed with React, Ant Design, please read [Airbnb Javascript Style Guide](https://github.com/airbnb/javascript) if you want to contribute to frontend. Cyclone also makes use of linter tools like Prettier, ESLint, Lint-staged to enforce its code quality.
