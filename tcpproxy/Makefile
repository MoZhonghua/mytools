SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


cli=bin/tcpproxy-cli
$(cli): $(SOURCES)
	go build -o ${cli} ./cmd/cli


server=bin/tcpproxy-server
$(server): $(SOURCES)
	go build -o ${server} ./cmd/server


all: $(cli) $(server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
