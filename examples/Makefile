PKG := github.com/mmmorris1975/upnp
MODULES := $(shell go list ${PKG}/examples/... | grep -v /vendor/ | xargs basename)

.PHONY: all
all: $(MODULES)

.PHONY: $(MODULES)
$(MODULES):
	go build -v -o $@.cmd ./$@

.PHONY: clean
clean:
	rm -f *.cmd
