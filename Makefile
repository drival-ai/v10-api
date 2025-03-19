VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || cat $(CURDIR)/.version 2> /dev/null || echo v0)
BLDVER = module:$(MODULE),version:$(VERSION),build:$(CIRCLE_BUILD_NUM)
BASE = $(CURDIR)
MODULE = v10-mvp-api

.PHONY: all $(MODULE) install
all: version $(MODULE)

$(MODULE):| $(BASE)
	@GO111MODULE=on GOFLAGS=-mod=vendor go build -v -trimpath -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o $(BASE)/bin/$@

$(BASE):
	@mkdir -p $(dir $@)

install:
	@GO111MODULE=on GOFLAGS=-mod=vendor go install -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -v

.PHONY: version list
version:
	@echo "Version: $(VERSION)"

list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs
