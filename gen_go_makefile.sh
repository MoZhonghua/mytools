#! /bin/bash

cat > Makefile << 'EOF'
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.PHONY: default
default: all

EOF

bins=$(/bin/ls cmd)

# write binary targets
echo >> Makefile
for b in $bins
do
cat >> Makefile << EOF
$b=bin/$b
\$($b): \$(SOURCES)
	go build -o \${$b} ./cmd/$b

EOF
echo >> Makefile
done

# write target "all"
echo -n "all: " >> Makefile
for b in $bins
do
    echo -n "\$($b) "  >> Makefile
done
echo >> Makefile
echo -e '\t' >> Makefile


# write target "clean"
cat >> Makefile << "EOF"
.PHONY: clean
clean:
	rm -fv bin/*
EOF
