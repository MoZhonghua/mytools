SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


tcpmux-admin=bin/tcpmux-admin
$(tcpmux-admin): $(SOURCES)
	go build -o ${tcpmux-admin} ./cmd/tcpmux-admin


tcpmux-client=bin/tcpmux-client
$(tcpmux-client): $(SOURCES)
	go build -o ${tcpmux-client} ./cmd/tcpmux-client


tcpmux-server=bin/tcpmux-server
$(tcpmux-server): $(SOURCES)
	go build -o ${tcpmux-server} ./cmd/tcpmux-server


all: $(tcpmux-admin) $(tcpmux-client) $(tcpmux-server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
