# Paths to packages
GO=$(shell which go)

# Paths to locations, etc
BUILD_DIR = "build"
PLUGIN_DIR = $(wildcard plugin/*)
BUILD_MODULE = "github.com/mutablelogic/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
BUILD_ARCH = $(shell $(GO) env GOARCH)
BUILD_PLATFORM = $(shell $(GO) env GOOS)

all: clean server plugins

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} ./cmd/server

plugins: $(PLUGIN_DIR)

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
#@${GO} mod tidy
	@${GO} clean
