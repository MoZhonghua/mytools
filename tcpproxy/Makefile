SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


tcpproxy-cli=bin/tcpproxy-cli
$(tcpproxy-cli): $(SOURCES)
	go build -o ${tcpproxy-cli} ./cmd/tcpproxy-cli


tcpproxy-server=bin/tcpproxy-server
$(tcpproxy-server): $(SOURCES)
	go build -o ${tcpproxy-server} ./cmd/tcpproxy-server


all: $(tcpproxy-cli) $(tcpproxy-server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
