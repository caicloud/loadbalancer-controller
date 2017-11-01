
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

# TODO git describe --tags --abbrev=0
# get release from tags
RELEASE?=v0.2.2
GOOS?=linux
PREFIX?=cargo.caicloudprivatetest.com/caicloud/loadbalancer-controller

# Current version of the project.
VERSION ?= v0.3.0

#
# These variables should not need tweaking.
#
 
# A list of all packages.
PKGS := $(shell go list ./... | grep -v /vendor | grep -v /tests)
 
# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Deployment direcotory.
BUILD_DIR := ./build

# Git commit sha.
COMMIT := $(shell git rev-parse --short HEAD)

# Golang standard bin directory.
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

#
# Tweak the variables based on your project.
#

# this pkg 
PKG := github.com/caicloud/loadbalancer-controller

# Target binaries. You can build multiple binaries for a single project.
TARGETS := controller
 
# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# Container registries.
REGISTRIES ?= cargo.caicloudprivatetest.com/caicloud

#
# Define all targets. At least the following commands are required:
#
 
# All targets.
.PHONY: lint test build container push build-local build-linux 

build: build-local

lint: $(GOMETALINTER)
	cat .gofmt | xargs -I {} gofmt -w -s -d -r {}  $$(find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*")
	gosimple $$(go list ./... | grep -v vendor)
	gometalinter ./... --vendor
 
$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

test:
	 @go test $(PKGS)

build-local: test
	go build -i -v -o $(OUTPUT_DIR)/$(TARGETS) \
	-ldflags "-s -w -X $(PKG)/pkg/version.RELEASE=$(VERSION) -X $(PKG)/pkg/version.COMMIT=$(COMMIT) -X $(PKG)/pkg/version.REPO=$(PKG)" \
	$(PKG)/cmd/$(TARGETS)

build-linux: test
	GOOS=linux GOARCH=amd64 go build -i -v -o $(OUTPUT_DIR)/$(TARGETS) \
	-ldflags "-s -w -X $(PKG)/pkg/version.RELEASE=$(VERSION) -X $(PKG)/pkg/version.COMMIT=$(COMMIT) -X $(PKG)/pkg/version.REPO=$(PKG)" \
	$(PKG)/cmd/$(TARGETS)

container: build-linux
	@for registry in $(REGISTRIES); do \
		image=$(IMAGE_PREFIX)$(TARGETS)$(IMAGE_SUFFIX); \
		docker build -t $${registry}/$${image}:$(VERSION) -f $(BUILD_DIR)/Dockerfile .; \
	done

push: container
	@for registry in $(REGISTRIES); do \
		image=$(IMAGE_PREFIX)$(TARGETS)$(IMAGE_SUFFIX); \
		docker push $${registry}/$${image}:$(VERSION); \
	done

debug: build-local
	$(OUTPUT_DIR)/$(TARGETS) --kubeconfig=$${HOME}/.kube/config --debug --log-force-color; 
