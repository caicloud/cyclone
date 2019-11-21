# +--------------------------------------------------------------+
# | * * *                moul.io/rules.mk                        |
# +--------------------------------------------------------------+
# |                                                              |
# |     ++              ______________________________________   |
# |     ++++           /                                      \  |
# |      ++++          |                                      |  |
# |    ++++++++++      |   https://moul.io/rules.mk is a set  |  |
# |   +++       |      |   of common Makefile rules that can  |  |
# |   ++         |     |   be configured from the Makefile    |  |
# |   +  -==   ==|     |   or with environment variables.     |  |
# |  (   <*>   <*>     |                                      |  |
# |   |          |    /|                      Manfred Touron  |  |
# |   |         _)   / |                        manfred.life  |  |
# |   |      +++    /  \______________________________________/  |
# |    \      =+   /                                             |
# |     \      +                                                 |
# |     |\++++++                                                 |
# |     |  ++++      ||//                                        |
# |  ___|   |___    _||/__                                     __|
# | /    ---    \   \|  |||                   __ _  ___  __ __/ /|
# |/  |       |  \    \ /                    /  ' \/ _ \/ // / / |
# ||  |       |  |    | |                   /_/_/_/\___/\_,_/_/  |
# +--------------------------------------------------------------+

##
## rules.mk
##
ifeq ($(RULESMK),1)
.PHONY: rulesmk.bumpdeps
rulesmk.bumpdeps:
	wget -O rules.mk https://raw.githubusercontent.com/moul/rules.mk/master/rules.mk

BUMPDEPS_STEPS += rulesmk.bumpdeps
endif

##
## Maintainer
##

.PHONY: generate.authors
generate.authors:
	echo "# This file lists all individuals having contributed content to the repository." > AUTHORS
	echo "# For how it is generated, see 'https://github.com/moul/rules.mk'" >> AUTHORS
	echo >> AUTHORS
	git log --format='%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf >> AUTHORS
GENERATE_STEPS += generate.authors

##
## Golang
##

ifdef GOPKG
GO ?= go

ifdef GOBINS
.PHONY: go.install
go.install:
	set -e; for dir in $(GOBINS); do ( set -xe; \
	        cd $$dir; \
	        $(GO) install .; \
	); done
INSTALL_STEPS += go.install
endif

.PHONY: go.unittest
go.unittest:
	echo "" > /tmp/coverage.txt
	set -e; for dir in `find . -type f -name "go.mod" | sed 's@/[^/]*$$@@' | sort | uniq`; do ( set -xe; \
	        cd $$dir; \
	        $(GO) test -v -cover -coverprofile=/tmp/profile.out -covermode=atomic -race ./...; \
	        if [ -f /tmp/profile.out ]; then \
	                cat /tmp/profile.out >> /tmp/coverage.txt; \
	                rm -f /tmp/profile.out; \
	        fi); done
	mv /tmp/coverage.txt .

.PHONY: go.lint
go.lint:
	set -e; for dir in `find . -type f -name "go.mod" | sed 's@/[^/]*$$@@' | sort | uniq`; do ( set -xe; \
	        cd $$dir; \
	        golangci-lint run --verbose ./...; \
	); done

.PHONY: go.tidy
go.tidy:
	set -e; for dir in `find . -type f -name "go.mod" | sed 's@/[^/]*$$@@' | sort | uniq`; do ( set -xe; \
	        cd $$dir; \
	        $(GO)	mod tidy; \
	); done

.PHONY: go.build
go.build:
	set -e; for dir in `find . -type f -name "go.mod" | sed 's@/[^/]*$$@@' | sort | uniq`; do ( set -xe; \
	        cd $$dir; \
	        $(GO)	build ./...; \
	); done

.PHONY: go.bump-deps
go.bumpdeps:
	set -e; for dir in `find . -type f -name "go.mod" | sed 's@/[^/]*$$@@' | sort | uniq`; do ( set -xe; \
	        cd $$dir; \
	        $(GO)	get -u ./...; \
	); done

.PHONY: go.release
go.release:
	goreleaser --snapshot --skip-publish --rm-dist
	@echo -n "Do you want to release? [y/N] " && read ans && \
	        if [ $${ans:-N} = y ]; then set -xe; goreleaser --rm-dist; fi

BUILD_STEPS += go.build
RELEASE_STEPS += go.release
BUMPDEPS_STEPS += go.bumpdeps
TIDY_STEPS += go.tidy
LINT_STEPS += go.lint
UNITTEST_STEPS += go.unittest
endif

##
## Node
##

ifdef NPM_PACKAGES
.PHONY: npm.publish
npm.publish:
	@echo -n "Do you want to npm publish? [y/N] " && read ans && \
	if [ $${ans:-N} = y ]; then \
	        set -e; for dir in $(NPM_PACKAGES); do ( set -xe; \
	                cd $$dir; \
	                npm publish --access=public; \
	        ); done; \
	fi
RELEASE_STEPS += npm.publish
endif

##
## Docker
##

ifdef DOCKER_IMAGE
.PHONY: docker.build
docker.build:
	docker build \
	        --build-arg VCS_REF=`git rev-parse --short HEAD` \
	        --build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
	        --build-arg VERSION=`git describe --tags --always` \
	        -t $(DOCKER_IMAGE) .

BUILD_STEPS += docker.build
endif

##
## Common
##

TEST_STEPS += $(UNITTEST_STEPS)
TEST_STEPS += $(LINT_STEPS)
TEST_STEPS += $(TIDY_STEPS)
ALL_STEPS += $(TEST_STEPS)

all: $(ALL_STEPS)

ifneq ($(strip $(TEST_STEPS)),)
.PHONY: test
test: $(TEST_STEPS)
endif

ifdef INSTALL_STEPS
.PHONY: install
install: $(INSTALL_STEPS)
endif

ifdef UNITTEST_STEPS
.PHONY: unittest
unittest: $(UNITTEST_STEPS)
endif

ifdef LINT_STEPS
.PHONY: lint
lint: $(LINT_STEPS)
endif

ifdef TIDY_STEPS
.PHONY: tidy
tidy: $(TIDY_STEPS)
endif

ifdef BUILD_STEPS
.PHONY: build
build: $(BUILD_STEPS)
endif

ifdef RELEASE_STEPS
.PHONY: release
release: $(RELEASE_STEPS)
endif

ifdef BUMPDEPS_STEPS
.PHONY: bumpdeps
bumpdeps: $(BUMPDEPS_STEPS)
endif

ifdef GENERATE_STEPS
.PHONY: generate
generate: $(GENERATE_STEPS)
endif
