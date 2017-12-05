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
# TODO: implement 'make push'

#
# Tweak the variables based on your project.
#

# Current version of the project.
VERSION ?= v0.4.0

# This repo's root import path (under GOPATH).
ROOT := github.com/caicloud/cyclone

# Target binaries. You can build multiple binaries for a single project.
TARGETS := server worker

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip cyclone-)
IMAGE_SUFFIX ?= $(strip )

# Container registries.
REGISTRIES ?= cargo.caicloud.io/caicloud

#
# These variables should not need tweaking.
#

# A list of all packages.
PKGS := $(shell go list ./... | grep -v /vendor | grep -v /test)

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Git commit sha.
COMMIT := $(shell git rev-parse --short HEAD)

# Golang standard bin directory.
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: lint test build container push

build: build-local

lint: $(GOMETALINTER)
	gometalinter ./... --vendor

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

test:
	go test $(PKGS)

build-local:
	@for target in $(TARGETS); do                                                      \
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
	        $${registry}/golang:1.9.2-alpine3.6                                        \
	          go build -i -v -o $(OUTPUT_DIR)/$${target}                               \
	            -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)              \
	            -X $(ROOT)/pkg/version.COMMIT=$(COMMIT)                                \
	            -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                               \
	            $(CMD_DIR)/$${target};                                                 \
	  done                                                                             \
	done

container: build-linux
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker build -t $${registry}/$${image}:$(VERSION)                              \
	      -f $(BUILD_DIR)/$${target}/Dockerfile .;                                     \
	  done                                                                             \
	done

push: container
	@for target in $(TARGETS); do                                                      \
	  for registry in $(REGISTRIES); do                                                \
	    image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                \
	    docker push $${registry}/$${image}:$(VERSION);                                 \
	  done                                                                             \
	done

.PHONY: clean
clean:
	-rm -vrf ${OUTPUT_DIR}
