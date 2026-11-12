# Build the commands by default
.DEFAULT_GOAL := all

# Executables
GO ?= $(shell which go 2>/dev/null)
NPM ?= $(shell which npm 2>/dev/null)
GOWORK ?= off

# Locations
BUILD_DIR ?= build
CMD_DIR := $(wildcard cmd/*)
TS_DIR := frontend

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')

# Set test flags, such as -v, or -run
TEST_FLAGS ?=

# Set build flags
VERSION_PKG = github.com/mutablelogic/go-server/pkg/version
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}"

###############################################################################
# ALL

.PHONY: all
all: build

###############################################################################
# BUILD

# Build the commands in the cmd directory
.PHONY: build
build: tidy $(CMD_DIR)

$(CMD_DIR): go-dep mkdir tidy
	@echo Build $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOWORK=${GOWORK} GOOS=${OS} GOARCH=${ARCH} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

# Build the TypeScript package into frontend/dist, from package-lock.json,
# once it's type checked
.PHONY: frontend
frontend: npm-dep
	@echo Build frontend
	@cd $(TS_DIR) && ${NPM} ci --silent && ${NPM} run --silent check && ${NPM} run --silent build

###############################################################################
# TEST

.PHONY: test
test: vet unit-test coverage-test

.PHONY: vet
vet: go-dep
	@echo Vet
	@GOWORK=${GOWORK} ${GO} vet ./...

.PHONY: unit-test
unit-test: go-dep
	@echo Unit Tests
	@GOWORK=${GOWORK} ${GO} test ${TEST_FLAGS} ./...

.PHONY: race-test
race-test: go-dep
	@echo Race Tests
	@GOWORK=${GOWORK} ${GO} test -race ${TEST_FLAGS} ./...

# Test the TypeScript package
.PHONY: frontend-test
frontend-test: npm-dep
	@echo Frontend Tests
	@cd $(TS_DIR) && ${NPM} ci --silent && ${NPM} run --silent check && ${NPM} test --silent

.PHONY: coverage-test
coverage-test: go-dep mkdir
	@echo Test Coverage
	@GOWORK=${GOWORK} ${GO} test ${TEST_FLAGS} -coverprofile ${BUILD_DIR}/coverprofile.out ./...
	@${GO} tool cover -func ${BUILD_DIR}/coverprofile.out > ${BUILD_DIR}/coverage.txt

###############################################################################
# CLEAN

.PHONY: mkdir
mkdir:
	@install -d $(BUILD_DIR)

.PHONY: tidy
tidy: go-dep
	@echo 'go tidy'
	@GOWORK=${GOWORK} $(GO) mod tidy

.PHONY: clean
clean:
	@echo 'clean'
	@rm -fr $(BUILD_DIR) $(TS_DIR)/dist
	@$(GO) clean

###############################################################################
# DEPENDENCIES

.PHONY: npm-dep
npm-dep:
	@test -f "${NPM}" && test -x "${NPM}"  || (echo "Missing npm binary, which builds the frontend" && exit 1)

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)
