# Current version of the project.
VERSION ?= v0.3.17

# This repo's root import path (under GOPATH).
ROOT := github.com/caicloud/cargo-admin

# Target binaries. You can build multiple binaries for a single project.
TARGETS := admin token

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip cargo-)
IMAGE_SUFFIX ?= $(strip )

# Registry where to pull library image, for example, golang
BASE_REGISTRY ?= cargo.caicloudprivatetest.com/library

# Registries to push the images. Multiple registries supported
REGISTRIES ?= cargo.caicloudprivatetest.com/release


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

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: test build container push

build: build-local

test:
	go test $(PKGS)

# Need go version 1.10 or higher
test-cover:
	go version ;                                                                       \
	go test $(PKGS) -coverprofile cover.out ;                                          \
	go tool cover -func cover.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }';\
	rm cover.out

build-local:
	@for target in $(TARGETS); do                                                      \
	  go build -i -v -o $(OUTPUT_DIR)/cargo-$${target}                                 \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                      \
	              -X $(ROOT)/pkg/version.COMMIT=$(COMMIT)                              \
	              -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                             \
	    $(CMD_DIR)/$${target};                                                         \
	done

build-linux:
	@for target in $(TARGETS); do                                                      \
	  docker run --rm                                                                  \
	    -v $(PWD):/go/src/$(ROOT)                                                      \
	    -w /go/src/$(ROOT)                                                             \
	    -e GOOS=linux                                                                  \
	    -e GOARCH=amd64                                                                \
	    -e GOPATH=/go                                                                  \
	    $(BASE_REGISTRY)/golang:1.9.2                                                  \
	    go build -i -v -o $(OUTPUT_DIR)/cargo-$${target}                               \
	      -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                    \
	      -X $(ROOT)/pkg/version.COMMIT=$(COMMIT)                                      \
	      -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                                     \
	      $(CMD_DIR)/$${target};                                                       \
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
