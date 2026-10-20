# Executables
GO ?= $(shell which go 2>/dev/null)

# Locations
BUILD_DIR ?= build
CMD_DIR := $(wildcard cmd/*)
PLUGIN_DIR := $(wildcard plugin/*)

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
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}" 

###############################################################################
# ALL

.PHONY: all
all: build

###############################################################################
# BUILD

# Compile and build the commands in the cmd directory, and the plugins
.PHONY: build
build: tidy $(PLUGIN_DIR) $(CMD_DIR)

# Build the commands
$(CMD_DIR): go-dep mkdir
	@echo Build command $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

# Build the plugins
.PHONY: plugins
plugins: $(PLUGIN_DIR) 

# Build the plugins
$(PLUGIN_DIR): go-dep mkdir
	@echo Build plugin $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build -buildmode=plugin ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@).plugin ./$@

###############################################################################
# TEST

.PHONY: test
test: vet unit-test coverage-test

.PHONY: vet
vet: go-dep
	@echo Vet
	@${GO} vet ./...

.PHONY: unit-test
unit-test: go-dep
	@echo Unit Tests
	@${GO} test ${VERBOSE_FLAG} ./pkg/...

.PHONY: race-test
race-test: go-dep
	@echo Race Tests
	@${GO} test -race ${VERBOSE_FLAG} ./pkg/...

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
	@${GO} clean -modcache

###############################################################################
# DEPENDENCIES

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)
