
# Go parameters
GO=go
GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 
BUILDDIR = build

# Commands to build
COMMANDS = $(wildcard ./cmd/*)

# Rules for building
.PHONY: commands $(COMMANDS)
commands: builddir $(COMMANDS)

$(COMMANDS): 
	@PKG_CONFIG_PATH="$(PKG_CONFIG_PATH)" $(GO) build -o ${BUILDDIR}/$@ -tags "$(TAGS)" ${GOFLAGS} ./$@

.PHONY: test
test:
	@PKG_CONFIG_PATH="$(PKG_CONFIG_PATH)" $(GO) test -tags "$(TAGS)" ./pkg/...

# The next example rule adds tag into compile given existence of pkg-config for cgo
# using mmal as the example
.PHONY: lib
lib:
	$(eval EXISTS = $(shell PKG_CONFIG_PATH="$(PKG_CONFIG_PATH)" pkg-config --silence-errors --modversion mmal))
ifneq ($strip $(MMAL)),)
	@echo "Targetting mmal"
	$(eval TAGS += mmal)
endif

.PHONY: builddir
builddir:
	install -d $(BUILDDIR)

.PHONY: clean
clean: 
	rm -fr $(BUILDDIR)
	$(GO) clean
	$(GO) mod tidy

