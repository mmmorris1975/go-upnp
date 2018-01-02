MODULES := discovery description control eventing

.PHONY: all
all: $(MODULES)

.PHONY: $(MODULES)
$(MODULES):
	go install -v -x ./$@
