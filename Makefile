PROJECT_ROOT := $(CURDIR)
BUILD_BASE ?= build
CMD_BASE ?= cmd
GO ?= go
GO_TEST ?= $(GO) test
GO_FMT ?= gofmt
DOCKER ?= docker

GO_FILES := $(wildcard *.go)
COMMANDS := $(shell find $(CMD_BASE) -mindepth 1 -maxdepth 1 | sed 's,^$(CMD_BASE)/,,')

BINARIES := $(addprefix $(BUILD_BASE)/, $(COMMANDS))

.PHONY: build
build: $(BINARIES)
	@echo '# build complete'

.PHONY: test
test:
	@echo -n '# testing' in && pwd && $(GO_TEST) -v
	@for cmd in $(COMMANDS); do \
		cd $(PROJECT_ROOT)/$(CMD_BASE)/$${cmd}; \
		echo -n '# testing' && pwd; \
		$(GO_TEST) -v || exit 1; \
	done
	@echo '# all tests passed--'

.PHONY: lint
lint:
	@echo running golangci-lint
	@$(DOCKER) run --rm -v $(PROJECT_ROOT):/app -w /app golangci/golangci-lint:v1.22.0 golangci-lint run
	@echo running go vet
	@$(GO) vet
	@echo running gofmt
	@test -z "`$(GO_FMT) -s -l .`"
	@echo running hadolint
	@$(DOCKER) run --rm -i hadolint/hadolint < Dockerfile

.PHONY: install
install: 
	@for cmd in $(COMMANDS); do \
		cd $(PROJECT_ROOT)/$(CMD_BASE)/$${cmd}; \
		echo -n '# installing from ' && pwd; \
		$(GO) install -v || exit 1; \
	done
	@echo '# install completed --'

.PHONY: uninstall
uninstall: 
	@for cmd in $(COMMANDS); do \
		cd $(PROJECT_ROOT)/$(CMD_BASE)/$${cmd}; \
		echo -n '# uninstalling from ' && pwd; \
		$(GO) clean -x -i || exit 1; \
	done
	@echo '# uninstall completed --'

.PHONY: clean
clean:
	@rm -rf $(BUILD_BASE)

$(BUILD_BASE)/%: cmd/% $(GO_FILES)
	@mkdir -p $(BUILD_BASE)
	@export BUILD_OUTPUT=`realpath "$(BUILD_BASE)"` && \
		cd $< && $(GO) build -v -o $$BUILD_OUTPUT/$(@F)
	@chmod +x $@ $(BUILD_BASE)/$(@F)
