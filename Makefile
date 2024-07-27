# Paths to tools needed in dependencies
GO := $(shell which go)
DOCKER := $(shell which docker)
NPM := $(shell which npm)

# Build flags
BUILD_MODULE := $(shell cat go.mod | head -1 | cut -d ' ' -f 2)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')
DOCKER_REGISTRY ?= ghcr.io/mutablelogic

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(wildcard cmd/*)
PLUGIN_DIR := $(wildcard plugin/*)
NPM_DIR := $(wildcard npm/*)
BUILD_TAG := ${DOCKER_REGISTRY}/go-server-${OS}-${ARCH}:${VERSION}

# Targets
all: clean plugins cmds npm 

cmds: $(CMD_DIR)

plugins: $(PLUGIN_DIR)

npm: $(NPM_DIR)

docker: docker-dep
	@echo build docker image: ${BUILD_TAG} for ${OS}/${ARCH}
	@${DOCKER} build \
		--tag ${BUILD_TAG} \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${BUILD_MODULE} \
		--build-arg VERSION=${VERSION} \
		-f etc/docker/Dockerfile .

docker-tag: docker-dep
	@echo ${BUILD_TAG}

docker-push: docker-dep
	@echo push docker image: ${BUILD_TAG}
	@${DOCKER} push ${BUILD_TAG}

test: go-dep
	@echo Test
	@${GO} mod tidy
	@${GO} test ./pkg/...

$(CMD_DIR): go-dep mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

$(PLUGIN_DIR): go-dep mkdir
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

$(NPM_DIR): npm-dep mkdir
	@echo Build npm $(notdir $@)
	@cd $@ && npm install && npm run build
	@${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).npm.plugin ./$@

FORCE:

npm-dep:
	@test -f "${NPM}" && test -x "${NPM}"  || (echo "Missing nom binary" && exit 1)

go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean

