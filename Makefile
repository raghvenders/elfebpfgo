GO=go
CLANG=clang-14
CC=gcc

ARCH ?= $(shell uname -m | sed 's/x86_64/amd64/g; s/aarch64/arm64/g')



EXTERNAL=./external

#cgo flags
CGO_CFLAGS_ = "-I$(abspath $(EXTERNAL))"
CGO_LDFLAGS = "-lelf -lz $(LIBBPF_OBJ)"
CGO_EXTLDFLAGS = '-w -extldflags "-static"'

CGO_CFLAGS_STATIC = "-I$(abspath $(EXTERNAL)) -I$(abspath ./vmlinux.h)"


#libbpf src and Objects
LIBBPF_SRC = $(abspath ./libbpf/src)
LIBBPF_OBJ = $(abspath ./$(EXTERNAL)/libbpf.a)
LIBBPF_OBJDIR = $(abspath ./$(EXTERNAL)/libbpf)
LIBBPF_DESTDIR = $(abspath ./$(EXTERNAL))

CFLAGS = -g -O2 -Wall -fpie


libbpf-uapi: $(LIBBPF_SRC)
# UAPI headers can be installed by a different package so they're not installed
# in by (libbpf) install rule.
	UAPIDIR=$(LIBBPF_DESTDIR) \
		$(MAKE) -C $(LIBBPF_SRC) install_uapi_headers


build-libbpf: libbpf-uapi $(LIBBPF_OBJ) 
	CC=$(CLANG) \
		CGO_CFLAGS=$(CGO_CFLAGS) \
		CGO_LDFLAGS=$(CGO_LDFLAGS) \
		GOOS=linux GOARCH=$(ARCH) \
		$(GO) build \
		-ldflags $(CGO_EXTLDFLAGS) \
		.

$(LIBBPF_OBJ): $(LIBBPF_SRC) $(wildcard $(LIBBPF_SRC)/*.[ch]) | $(EXTERNAL)/libbpf
	CC="$(CC)" CFLAGS="$(CFLAGS)" LD_FLAGS="$(LDFLAGS)" \
	   $(MAKE) -C $(LIBBPF_SRC) \
		BUILD_STATIC_ONLY=1 \
		OBJDIR=$(LIBBPF_OBJDIR) \
		DESTDIR=$(LIBBPF_DESTDIR) \
		INCLUDEDIR= LIBDIR= UAPIDIR= install

$(LIBBPF_SRC):
ifeq ($(wildcard $@), )
	echo "INFO: updating submodule 'libbpf'"
	$(GIT) submodule update --init --recursive
endif


$(EXTERNAL):
	mkdir -p $(EXTERNAL)

$(EXTERNAL)/libbpf:
	mkdir -p $(EXTERNAL)/libbpf


# cleanup

clean: 	
	rm -rf $(EXTERNAL)

# Examples

MAIN=main


$(MAIN).bpf.o: kill.bpf.c
	$(CLANG) $(CFLAGS) -target bpf -D__TARGET_ARCH_$(ARCH) -I$(EXTERNAL) -I$(abspath ./vmlinux.h) -c $< -o $@

.PHONY: run-ebpf

run-ebpf: build-libbpf | $(MAIN).bpf.o
	CC=$(CLANG) \
		CGO_CFLAGS=$(CGO_CFLAGS_STATIC) \
		CGO_LDFLAGS=$(CGO_LDFLAGS) \
		GOOS=linux GOARCH=$(ARCH) \
		$(GO) build \
	  -tags org -v -ldflags $(CGO_EXTLDFLAGS) \
		-o ebpf-pro ./examples/$(MAIN).go

