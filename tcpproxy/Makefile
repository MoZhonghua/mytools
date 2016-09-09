SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

bin=bin/tcpproxy

all: $(bin) 

$(bin): $(SOURCES)
	go build -o ${bin} ./

.PHONY: clean install
clean:
	rm -fv bin/*

install:
	install -d /opt/tools/tcpproxy/bin/
	install ${bin} -t /opt/tools/tcpproxy/bin/
