# Paths to packages
GO=$(shell which go)
ARCH=$(shell which arch)
UNAME=$(shell which uname)
DOCKER=$(shell which docker)

# Go Modules
GOMODULE := "github.com/mutablelogic/go-server"

# Docker
DOCKER_CONTAINER := "ghcr.io/mutablelogic/go-server"

# Paths to locations, etc
BUILD_DIR := "build"
BUILD_ARCH := $(shell ${ARCH} | tr A-Z a-z)
BUILD_PLATFORM := $(shell ${UNAME} | tr A-Z a-z)
BUILD_VERSION := $(shell git describe --tags | sed 's/^v//')
PLUGIN_DIR := $(filter-out $(wildcard plugin/*.go), $(wildcard plugin/*))
CMD_DIR := $(wildcard cmd/*)

# Add linker flags
BUILD_LD_FLAGS += -X $(GOMODULE)/pkg/version.GitSource=${GOMODULE}
BUILD_LD_FLAGS += -X $(GOMODULE)/pkg/version.GitTag=${BUILD_VERSION}
BUILD_LD_FLAGS += -X $(GOMODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(GOMODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(GOMODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

all: clean cmd plugins

cmd: $(CMD_DIR)

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

plugins: $(PLUGIN_DIR)

$(PLUGIN_DIR): dependencies mkdir
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

test: dependencies
	@echo Running tests
	@${GO} test ./pkg/...

docker: docker-dep
	@echo Building docker image
	@${DOCKER} build \
		--tag ${DOCKER_CONTAINER} \
		--build-arg VERSION=${BUILD_VERSION} \
		--build-arg ARCH=${BUILD_ARCH} \
		--build-arg PLATFORM=${BUILD_PLATFORM} \
		-f etc/docker/Dockerfile .

FORCE:

docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

dependencies:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)
	@test -f "${ARCH}" && test -x "${ARCH}"  || (echo "Missing arch binary" && exit 1)
	@test -f "${UNAME}" && test -x "${UNAME}"  || (echo "Missing uname binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
