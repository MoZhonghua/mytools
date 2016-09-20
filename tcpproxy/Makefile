SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


tcpproxy-admin=bin/tcpproxy-admin
$(tcpproxy-admin): $(SOURCES)
	go build -o ${tcpproxy-admin} ./cmd/tcpproxy-admin


tcpproxy-server=bin/tcpproxy-server
$(tcpproxy-server): $(SOURCES)
	go build -o ${tcpproxy-server} ./cmd/tcpproxy-server


all: $(tcpproxy-admin) $(tcpproxy-server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
