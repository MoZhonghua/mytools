SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


upnp=bin/upnp
$(upnp): $(SOURCES)
	go build -o ${upnp} ./cmd/upnp


all: $(upnp) 
	
.PHONY: clean
clean:
	rm -fv bin/*
