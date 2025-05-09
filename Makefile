# Executables
GO ?= $(shell which go 2>/dev/null)
DOCKER ?= $(shell which docker 2>/dev/null)
NPM ?= $(shell which npm 2>/dev/null)

# Locations
BUILD_DIR ?= build
CMD_DIR := $(wildcard cmd/*)
PLUGIN_DIR := $(wildcard plugin/*)
NPM_DIR := $(wildcard npm/*)

# VERBOSE=1
ifneq ($(VERBOSE),)
  VERBOSE_FLAG = -v
else
  VERBOSE_FLAG =
endif

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')

# Set build flags
BUILD_MODULE = $(shell cat go.mod | head -1 | cut -d ' ' -f 2)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}" 

# Docker
DOCKER_REPO ?= ghcr.io/mutablelogic/go-server
DOCKER_SOURCE ?= ${BUILD_MODULE}
DOCKER_TAG = ${DOCKER_REPO}-${OS}-${ARCH}:${VERSION}

###############################################################################
# ALL

.PHONY: all
all: clean build

###############################################################################
# BUILD

# Compile NPM and build the commands in the cmd directory
.PHONY: build
build: $(NPM_DIR) build-docker

# Build the commands in the cmd directory
.PHONY: build-docker
build-docker: tidy $(PLUGIN_DIR) $(CMD_DIR)

# Build the commands
$(CMD_DIR): go-dep mkdir
	@echo Build command $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

# Build the plugins
.PHONY: plugins
plugins: $(NPM_DIR) $(PLUGIN_DIR) 

# Build the plugins
$(PLUGIN_DIR): go-dep mkdir
	@echo Build plugin $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

# Build the NPM packages
$(NPM_DIR): npm-dep
	@echo Build npm $(notdir $@)
	@cd $@ && npm install && npm run dist	

# Build the docker image
.PHONY: docker
docker: docker-dep ${NPM_DIR}
	@echo build docker image ${DOCKER_TAG} OS=${OS} ARCH=${ARCH} SOURCE=${DOCKER_SOURCE} VERSION=${VERSION}
	@${DOCKER} build \
		--tag ${DOCKER_TAG} \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${DOCKER_SOURCE} \
		--build-arg VERSION=${VERSION} \
		-f etc/docker/Dockerfile .

# Push docker container
.PHONY: docker-push
docker-push: docker-dep 
	@echo push docker image: ${DOCKER_TAG}
	@${DOCKER} push ${DOCKER_TAG}

# Print out the version
.PHONY: docker-version
docker-version: docker-dep 
	@echo "tag=${VERSION}"

###############################################################################
# TEST

.PHONY: test
test: unit-test coverage-test

.PHONY: unit-test
unit-test: go-dep
	@echo Unit Tests
	@${GO} test ${VERBOSE_FLAG} ./pkg/...

.PHONY: coverage-test
coverage-test: go-dep mkdir
	@echo Test Coverage
	@${GO} test -coverprofile ${BUILD_DIR}/coverprofile.out ./pkg/...

###############################################################################
# CLEAN

.PHONY: tidy
tidy:
	@echo Running go mod tidy
	@${GO} mod tidy

.PHONY: mkdir
mkdir:
	@install -d ${BUILD_DIR}

.PHONY: clean
clean:
	@echo Clean $(BUILD_DIR)
	@rm -fr $(BUILD_DIR)
	@${GO} clean
	@for dir in $(NPM_DIR); do \
		echo "Cleaning npm $$dir"; \
		cd $$dir && ${NPM} run clean; \
	done

###############################################################################
# DEPENDENCIES

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

.PHONY: docker-dep
docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

.PHONY: npm-dep
npm-dep:
	@test -f "${NPM}" && test -x "${NPM}"  || (echo "Missing npm binary" && exit 1)
