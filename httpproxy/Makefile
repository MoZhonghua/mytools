SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


httpproxy=bin/httpproxy
$(httpproxy): $(SOURCES)
	go build -o ${httpproxy} ./cmd/httpproxy


all: $(httpproxy) 
	
.PHONY: clean
clean:
	rm -fv bin/*
