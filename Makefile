include $(GOROOT)/src/Make.inc

YAML=yaml-0.1.3
LIBYAML=$(PWD)/$(YAML)/src/.libs/libyaml.a

TARG=goyaml

GOFILES=\
	goyaml.go\
	resolve.go\

CGOFILES=\
	decode.go\

CGO_LDFLAGS+=-lm -lpthread
CGO_CFLAGS+=-I$(YAML)/include
CGO_OFILES+=_lib/*.o


all: package

$(CGO_OFILES): $(LIBYAML)
	@mkdir -p _lib
	cd _lib && ar x $(LIBYAML)

$(LIBYAML):
	cd $(YAML) && CFLAGS=-fpic ./configure && make

CLEANFILES=_lib

include $(GOROOT)/src/Make.pkg

_cgo_defun.c: helpers.c
