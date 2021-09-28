# Paths to packages
GO=$(shell which go)
SED=$(shell which sed)
NFPM=$(shell which nfpm)

# Paths to locations, etc
BUILD_DIR = "build"
PLUGIN_DIR = $(wildcard plugin/*)
BUILD_MODULE = "github.com/djthorpe/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
BUILD_ARCH = $(shell $(GO) env GOARCH)
BUILD_PLATFORM = $(shell $(GO) env GOOS)

all: clean server plugins

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} ./cmd/server

plugins: $(PLUGIN_DIR)
	@echo Build plugin media 
	@${GO} get github.com/djthorpe/go-media
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/media.plugin ${BUILD_FLAGS} github.com/djthorpe/go-media/plugin/media

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

deb: nfpm go-server-deb

go-server-deb: server plugin/httpserver plugin/log plugin/basicauth plugin/static
	@echo Build go-server deb package
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server/nfpm.yaml > $(BUILD_DIR)/go-server-nfpm.yaml
	@${NFPM} pkg -f $(BUILD_DIR)/go-server-nfpm.yaml --packager deb --target $(BUILD_DIR)

nfpm:
	@echo Installing nfpm
	@${GO} mod tidy
	@${GO} install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.3.1	

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif
ifeq (,${SED})
        $(error "Missing sed binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
