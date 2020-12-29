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

# This repo's root import path (under GOPATH).
ROOT := github.com/caicloud/cyclone

# Target binaries. You can build multiple binaries for a single project.
TARGETS ?= server workflow/controller workflow/coordinator cicd/cd toolbox/fstream
IMAGES ?= server workflow/controller workflow/coordinator resolver/git resolver/svn resolver/image resolver/http watcher cicd/cd cicd/sonarqube toolbox
BASE_IMAGES := base/alpine base/openjdk

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip cyclone-)
IMAGE_SUFFIX ?= $(strip )

# Container registry for target images.
REGISTRY ?= docker.io/caicloud

# Container registry for images used to complile, like building golang binaries and building webs, e.g. docker.io/library/golang.
BASE_REGISTRY ?= docker.io/library

# Container registry for cyclone target base images.
CYCLONE_BASE_REGISTRY ?= docker.io/caicloud

# Version of the cyclone base images.
CYCLONE_BASE_VERSION ?= v1.1.0

# Example scene
SCENE ?= cicd

# Go build GOARCH, you can choose to build amd64 or arm64
ARCH ?= amd64

# Change Dockerfile name and registry project name for arm64
ifeq ($(ARCH),arm64)
DOCKERFILE ?= Dockerfile.arm64
REGISTRY ?= cargo.dev.caicloud.xyz/arm64v8
CYCLONE_BASE_VERSION := $(CYCLONE_BASE_VERSION)-arm64v8
else
DOCKERFILE ?= Dockerfile
REGISTRY ?= cargo.dev.caicloud.xyz/release
endif

#
# These variables should not need tweaking.
#

# It's necessary to set this because some environments don't link sh -> bash.
export SHELL := /bin/bash

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

# This will force go to use the vendor files instead of using the `$GOPATH/pkg/mod`. (vendor mode)
export GOFLAGS := -mod=vendor

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

# Available cpus for compiling, please refer to https://github.com/caicloud/engineering/issues/8186#issuecomment-518656946 for more information.
CPUS ?= $(shell /bin/bash hack/read_cpus_available.sh)

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

# Golang standard bin directory.
GOPATH ?= $(shell go env GOPATH)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: lint test build container push

lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

build: build-local

$(GOLANGCI_LINT):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_DIR) v1.20.1

test:
	@go test -race -p $(CPUS) $$(go list ./... | grep -v /vendor | grep -v /test) -coverprofile=coverage.out
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

build-local:
	@for target in $(TARGETS); do                                                      \
	  CGO_ENABLED=0   GOOS=linux   GOARCH=$(ARCH)                                      \
	  go build -trimpath -i -v -o $(OUTPUT_DIR)/$${target} -p $(CPUS)                  \
	    -ldflags "-s -w -X $(ROOT)/pkg/server/version.VERSION=$(VERSION)               \
	              -X $(ROOT)/pkg/server/version.COMMIT=$(COMMIT)                       \
	              -X $(ROOT)/pkg/server/version.REPOROOT=$(ROOT)"                      \
	    $(CMD_DIR)/$${target};                                                         \
	done

build-linux:
	@docker run --rm -it                                                               \
	  -v $(PWD):/go/src/$(ROOT)                                                        \
	  -w /go/src/$(ROOT)                                                               \
	  -e GOOS=linux                                                                    \
	  -e GOARCH=$(ARCH)                                                                \
	  -e GOPATH=/go                                                                    \
	  -e CGO_ENABLED=0                                                                 \
	  -e GOFLAGS=$(GOFLAGS)                                                            \
	  -e SHELLOPTS=$(SHELLOPTS)                                                        \
	  $(BASE_REGISTRY)/golang:1.13.9-stretch                                           \
	    /bin/bash -c 'for target in $(TARGETS); do                                     \
	      go build -trimpath -i -v -o $(OUTPUT_DIR)/$${target} -p $(CPUS)              \
	        -ldflags "-s -w -X $(ROOT)/pkg/server/version.VERSION=$(VERSION)           \
	          -X $(ROOT)/pkg/server/version.COMMIT=$(COMMIT)                           \
	          -X $(ROOT)/pkg/server/version.REPOROOT=$(ROOT)"                          \
	        $(CMD_DIR)/$${target};                                                     \
	    done'

build-web:
	docker run --rm -it                                                                \
	  -v $(PWD)/web/:/app                                                              \
	  -w /app                                                                          \
	  -e SHELLOPTS=$(SHELLOPTS)                                                        \
	    $(BASE_REGISTRY)/node:10.16-stretch                                            \
	      sh -c '                                                                      \
	        yarn;                                                                      \
	        yarn build'

build-web-local:
	sh -c '                                                                            \
	  cd web;                                                                          \
	  yarn;                                                                            \
	  yarn build'

container-web: build-web
	imageName=$(IMAGE_PREFIX)web$(IMAGE_SUFFIX);                                       \
	docker build -t ${REGISTRY}/$${imageName}:$(VERSION)                               \
	  -f $(BUILD_DIR)/web/$(DOCKERFILE).local .

container-web-local: build-web-local
	imageName=$(IMAGE_PREFIX)web$(IMAGE_SUFFIX);                                       \
	docker build -t ${REGISTRY}/$${imageName}:$(VERSION)                               \
	  -f $(BUILD_DIR)/web/$(DOCKERFILE).local .

container: build-linux
	@for image in $(IMAGES); do                                                        \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                          \
	  docker build -t ${REGISTRY}/$${imageName}:$(VERSION)                             \
	    -f $(BUILD_DIR)/$${image}/$(DOCKERFILE) .;                                     \
	done

container-local: build-local
	@for image in $(IMAGES); do                                                        \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                          \
	  docker build -t ${REGISTRY}/$${imageName}:$(VERSION)                             \
	    -f $(BUILD_DIR)/$${image}/$(DOCKERFILE) .;                                     \
	done

container-base:
	@for image in $(BASE_IMAGES); do                                                   \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                          \
	  docker build -t ${CYCLONE_BASE_REGISTRY}/$${imageName}:${CYCLONE_BASE_VERSION}   \
	    -f $(BUILD_DIR)/$${image}/$(DOCKERFILE) .;                                     \
	done

push:
	@for image in $(IMAGES); do                                                        \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX);                          \
	  docker push ${REGISTRY}/$${imageName}:$(VERSION);                                \
	done

push-local: container-local
	@for image in $(IMAGES); do \
	  imageName=$(IMAGE_PREFIX)$${image/\//-}$(IMAGE_SUFFIX); \
	  docker push ${REGISTRY}/$${imageName}:$(VERSION); \
	done

gen: clean-generated
	@./hack/update-codegen.sh
	sed -i 's|v1alpha1.Resource(|v1alpha1.GroupResource(|' ./pkg/k8s/listers/cyclone/v1alpha1/*.go

swagger-local:
	nirvana api --output web/public pkg/server/apis

swagger:
	docker run --rm -it                                                               \
	  -v $(PWD):/go/src/$(ROOT)                                                       \
	  -w /go/src/$(ROOT)                                                              \
	  -e GOOS=linux                                                                   \
	  -e GOARCH=$(ARCH)                                                               \
	  -e GOPATH=/go                                                                   \
	  -e CGO_ENABLED=0                                                                \
	  -e GOFLAGS=$(GOFLAGS)                                                           \
	  $(BASE_REGISTRY)/golang:1.13.9-stretch                                          \
	  sh -c "go get -u github.com/caicloud/nirvana/cmd/nirvana &&                     \
	  go get -u github.com/golang/dep/cmd/dep &&                                      \
	  nirvana api --output web/public pkg/server/apis"

run_examples:
	./examples/${SCENE}/generate.sh --registry=${REGISTRIY}
	kubectl create -f ./examples/${SCENE}/.generated

remove_examples:
	./examples/${SCENE}/generate.sh --registry=${REGISTRIY}
	kubectl delete -f ./examples/${SCENE}/.generated

.PHONY: clean
clean:
	-rm -vrf ${OUTPUT_DIR}
clean-generated:
	-rm -rf ./pkg/k8s/informers
	-rm -rf ./pkg/k8s/clientset
	-rm -rf ./pkg/k8s/listers
