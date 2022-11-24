# Paths to tools needed in dependencies
GO := $(shell which go)
NPM := $(shell which npm)

# Paths to locations, etc
BUILD_DIR := "build"
PLUGIN_DIR := $(wildcard plugin/*)
NPM_DIR := $(wildcard npm/*)
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))
BUILD_MODULE = "github.com/mutablelogic/go-server"

# Build flags
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

# Targets
all: clean cmd npm plugins

cmd: $(CMD_DIR)

npm: $(NPM_DIR)

plugins: $(PLUGIN_DIR)

$(CMD_DIR): dependencies mkdir FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): dependencies mkdir FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

$(NPM_DIR): dependencies FORCE
	@echo Build frontend $(notdir $@)
	cd $@ && ${NPM} install
	cd $@ && ${NPM} run build

FORCE:

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif
ifeq (,${NPM})
        $(error "Missing npm binary")
endif

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@find ${NPM_DIR} -name node_modules -type d -prune -exec rm -fr {} \;
	@find ${NPM_DIR} -name dist -type d -prune -exec rm -fr {} \;
	@${GO} mod tidy
	@${GO} clean
