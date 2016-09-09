SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

client_bin=bin/tunnelc
server_bin=bin/tunnels

all: $(client_bin) $(server_bin)

$(client_bin): $(SOURCES)
	go build -o ${client_bin} ./tunnelc

$(server_bin): $(SOURCES)
	go build -o ${server_bin} ./tunnels

.PHONY: clean install
clean:
	rm -fv bin/*

install:
	install -d /opt/tools/tunnel/bin/
	install ${client_bin} ${server_bin} -t /opt/tools/tunnel/bin/
