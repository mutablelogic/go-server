# Paths to tools needed in dependencies
GO := $(shell which go)
DOCKER := $(shell which docker)

# nginx image and version
NGINX := "library/nginx"
VERSION := "1.23.1"
IMAGE := "nginx-gateway"

# target architectures: linux/amd64 linux/arm/v7 linux/arm64/v8

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

# Docker target will build the docker image for amd64, arm and arm66. The output image is
# nginx-gateway:1.23.1 (or whatever is in the IMAGE and VERSION variables)
# which can then be pushed to ghcr.io
docker: dependencies docker-dependencies
	@${DOCKER} build --tag ${IMAGE}-arm:${VERSION} --build-arg VERSION=${VERSION} --build-arg PLATFORM=linux/arm/v7 etc/docker
	@${DOCKER} build --tag ${IMAGE}-arm64:${VERSION} --build-arg VERSION=${VERSION} --build-arg PLATFORM=linux/arm64/v8 etc/docker
	@${DOCKER} build --tag ${IMAGE}-amd64:${VERSION} --build-arg VERSION=${VERSION} --build-arg PLATFORM=linux/amd64 etc/docker
	@${DOCKER} manifest create ${IMAGE}:${VERSION} --amend ${IMAGE}-arm:${VERSION} --amend ${IMAGE}-arm64:${VERSION} --amend ${IMAGE}-amd64:${VERSION}
	@${DOCKER} manifest annotate ${IMAGE}:${VERSION} ${IMAGE}-arm:${VERSION} --arch arm --os linux --variant v7
	@${DOCKER} manifest annotate ${IMAGE}:${VERSION} ${IMAGE}-arm64:${VERSION} --arch arm64 --os linux --variant v8
	@${DOCKER} manifest annotate ${IMAGE}:${VERSION} ${IMAGE}-amd64:${VERSION} --arch amd64 --os linux

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

