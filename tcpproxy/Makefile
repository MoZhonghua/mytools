SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


cli=bin/cli
$(cli): $(SOURCES)
	go build -o ${cli} ./cmd/cli


server=bin/server
$(server): $(SOURCES)
	go build -o ${server} ./cmd/server


all: $(cli) $(server) 
	
.PHONY: clean
clean:
	rm -fv bin/*
