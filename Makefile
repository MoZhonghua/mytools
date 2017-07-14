SUBDIRS=mhttp tcpproxy upnp
all:
	 for dir in $(SUBDIRS); do \
	   $(MAKE) -j -C $$dir; \
	 done

clean:
	 for dir in $(SUBDIRS); do \
	   $(MAKE) -j -C $$dir clean; \
	 done
