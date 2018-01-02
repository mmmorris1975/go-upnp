PKG := github.com/mmmorris1975/upnp
MODULES := $(shell go list ${PKG}/... | grep -v /vendor/ | xargs basename)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all
all: $(MODULES)

.PHONY: $(MODULES)
$(MODULES):
	go install -v ./$@

.PHONY: release
release:
	ls -l $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(PKG)/
