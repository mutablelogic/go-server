# Paths to tools needed in dependencies
GO := $(shell which go)
DOCKER := $(shell which docker)

# nginx image and version
NGINX := library/nginx
VERSION := 1.23.1

# Target image name, architecture and version
IMAGE := go-server
ARCH := $(shell dpkg --print-architecture)
PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]')
TAG := $(shell git describe --tags)

# target architectures: linux/amd64 linux/arm linux/arm64

# Build flags
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

# Paths to locations, etc
BUILD_DIR := "build"
PLUGIN_DIR := $(wildcard plugin/*)
CMD_DIR := $(wildcard cmd/*)

# Targets
all: clean cmd plugins

cmd: $(filter-out cmd/README.md, $(wildcard cmd/*))

plugins: $(filter-out $(wildcard plugin/*.go), $(wildcard plugin/*))

test:
	@${GO} mod tidy
	@${GO} test -v ./pkg/...

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

$(PLUGIN_DIR): dependencies mkdir
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

# Docker target will build the docker image for amd64, arm and arm64
docker: dependencies docker-dependencies
	@echo Build ${IMAGE}-${ARCH}:${TAG} for platform ${PLATFORM}
	@${DOCKER} build --tag ${IMAGE}-${ARCH}:${TAG} --build-arg VERSION=${VERSION} --build-arg ARCH=${ARCH} --build-arg PLATFORM=${PLATFORM} -f etc/docker/Dockerfile .

FORCE:

docker-dependencies:
	@test -x ${DOCKER} || (echo "Docker not found" && exit 1)

dependencies:
	@test -x ${GO} || (echo "Missing go binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean

