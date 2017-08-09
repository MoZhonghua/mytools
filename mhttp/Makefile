SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all


mhttp=bin/mhttp
$(mhttp): $(SOURCES)
	go build -o ${mhttp} .


all: $(mhttp)
	
.PHONY: clean
clean:
	rm -fv bin/mhttp
