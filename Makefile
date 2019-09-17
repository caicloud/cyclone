# Copyright 2017 The Caicloud Authors.
#
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
#
#   make        - default to 'build' target
#   make lint   - code analysis
#   make test   - run unit test (or plus integration test)
#   make build        - alias to build-local target
#   make build-local  - build local binary targets
#   make build-linux  - build linux binary targets
#   make container    - build containers
#   make push    - push containers
#   make clean   - clean up targets
#
# Not included but recommended targets:
#   make e2e-test
#
# The makefile is also responsible to populate project version information.
#
# Tweak the variables based on your project.

# Set shell to bash
SHELL := /bin/bash

# This repo's root import path (under GOPATH).
ROOT := github.com/caicloud/cyclone

# Target binaries. You can build multiple binaries for a single project.
TARGETS := server workflow/controller workflow/coordinator cicd/cd toolbox/fstream
IMAGES := server workflow/controller workflow/coordinator resolver/git resolver/svn resolver/image watcher cicd/cd cicd/sonarqube toolbox web

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip cyclone-)
IMAGE_SUFFIX ?= $(strip )

# Container registries.
REGISTRIES ?= docker.io/library

# Example scene
SCENE ?= cicd

#
# These variables should not need tweaking.
#

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Git commit sha.
COMMIT := $(shell git rev-parse --short HEAD)

# Git tag describe.
TAG = $(shell git describe --tags --always --dirty)

# Current version of the project.
VERSION ?= $(TAG)

# Golang standard bin directory.
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: lint test build container push

lint: $(GOMETALINTER)
	golangci-lint run --skip-dirs=pkg/k8s/ --deadline=300s ./pkg/... ./cmd/...

build: build-local

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

test:
	@go test $$(go list ./... | grep -v /vendor | grep -v /test) -coverprofile=coverage.out
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

build-local:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED=0   GOOS=linux   GOARCH=amd64                                        \
	  go build -i -v -o $(OUTPUT_DIR)/$${target}                                       \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                      \
	              -X $(ROOT)/pkg/version.COMMIT=$(COMMIT)                              \
	              -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                             \
	    $(CMD_DIR)/$${target};                                                         \
	done

build-linux:
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    docker run --rm                                                                \
	      -v $(PWD):/go/src/$(ROOT)                                                    \
	      -w /go/src/$(ROOT)                                                           \
	      -e GOOS=linux                                                                \
	      -e GOARCH=amd64                                                              \
	      -e GOPATH=/go                                                                \
	      -e CGO_ENABLED=0                                                             \
	        golang:1.10-alpine3.8                                                      \
	          go build -i -v -o $(OUTPUT_DIR)/$${target}                               \
	            -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)              \
	            -X $(ROOT)/pkg/version.COMMIT=$(COMMIT)                                \
	            -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                               \
	            $(CMD_DIR)/$${target};                                                 \
	  done                                                                             \
	done

build-web:
	for registry in $(REGISTRIES); do                                                  \
	  docker run --rm                                                                  \
	    -v $(PWD)/web/:/app                                                            \
	    -w /app                                                                        \
	      node:8.9-alpine                                                              \
	        sh -c '                                                                    \
	          yarn;                                                                    \
	          yarn build';                                                             \
	done

build-web-local:
	sh -c '                                                                            \
	  cd web;                                                                          \
	  yarn;                                                                            \
	  yarn build'

container: build-linux
	@for image in $(IMAGES); do                                                        \
	  for registry in $(REGISTRIES); do                                                \
	    imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                        \
	    docker build -t $${registry}/$${imageName}:$(VERSION)                          \
	      -f $(BUILD_DIR)/$${image}/Dockerfile .;                                      \
	  done                                                                             \
	done

container-web-local: build-web-local
	@for registry in $(REGISTRIES); do                                                 \
	  imageName=$(IMAGE_PREFIX)web$(IMAGE_SUFFIX);                                     \
	  docker build -t $${registry}/$${imageName}:$(VERSION)                            \
	    -f $(BUILD_DIR)/web/Dockerfile.local .;                                        \
	done

container-local: build-local
	@for image in $(IMAGES); do                                                        \
	  for registry in $(REGISTRIES); do                                                \
	    imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                        \
	    docker build -t $${registry}/$${imageName}:$(VERSION)                          \
	      -f $(BUILD_DIR)/$${image}/Dockerfile .;                                      \
	  done                                                                             \
	done

push: container
	@for image in $(IMAGES); do                                                        \
	  for registry in $(REGISTRIES); do                                                \
	    imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                        \
	    docker push $${registry}/$${imageName}:$(VERSION);                             \
	  done                                                                             \
	done

gen: clean-generated
	bash tools/generator/autogenerate.sh

swagger-local:
	nirvana api --output web/public pkg/server/apis

swagger:
	docker run --rm                                                                   \
	  -v $(PWD):/go/src/$(ROOT)                                                       \
	  -w /go/src/$(ROOT)                                                              \
	  -e GOOS=linux                                                                   \
	  -e GOARCH=amd64                                                                 \
	  -e GOPATH=/go                                                                   \
	  -e CGO_ENABLED=0                                                                \
	  golang:1.10-alpine3.8                                                           \
	  sh -c "apk add git &&                                                           \
	  go get -u github.com/caicloud/nirvana/cmd/nirvana &&                            \
	  go get -u github.com/golang/dep/cmd/dep &&                                      \
	  nirvana api --output web/public pkg/server/apis"

run_examples:
	./examples/${SCENE}/generate.sh --registry=${REGISTRIES}
	kubectl create -f ./examples/${SCENE}/.generated

remove_examples:
	./examples/${SCENE}/generate.sh --registry=${REGISTRIES}
	kubectl delete -f ./examples/${SCENE}/.generated

.PHONY: clean
clean:
	-rm -vrf ${OUTPUT_DIR}
clean-generated:
	-rm -rf ./pkg/k8s/informers
	-rm -rf ./pkg/k8s/clientset
	-rm -rf ./pkg/k8s/listers