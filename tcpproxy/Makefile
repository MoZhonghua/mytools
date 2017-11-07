SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all

LDFLAGS="-linkmode external -extldflags -static"

tcpproxy-admin=bin/tcpproxy-admin
$(tcpproxy-admin): $(SOURCES)
	go build -ldflags $(LDFLAGS) -o ${tcpproxy-admin} ./cmd/tcpproxy-admin


tcpproxy-server=bin/tcpproxy-server
$(tcpproxy-server): $(SOURCES)
	go build -ldflags $(LDFLAGS) -o ${tcpproxy-server} ./cmd/tcpproxy-server


all: $(tcpproxy-admin) $(tcpproxy-server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
