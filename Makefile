include $(GOROOT)/src/Make.inc

YAML=yaml-0.1.3
LIBYAML=$(PWD)/$(YAML)/src/.libs/libyaml.a

TARG=goyaml

GOFILES=\
	goyaml.go\
	resolve.go\

CGOFILES=\
	decode.go\
	encode.go\

CGO_LDFLAGS+=-lm -lpthread
CGO_CFLAGS+=-I$(YAML)/include
#CGO_OFILES+=_lib/*.o
CGO_OFILES+=\
	helpers.o\
	_lib/api.o\
	_lib/scanner.o\
	_lib/reader.o\
	_lib/parser.o\
	_lib/writer.o\
	_lib/emitter.o\


all: package

_lib/api.o: $(LIBYAML)
	@mkdir -p _lib
	cd _lib && ar x $(LIBYAML)

$(LIBYAML):
	cd $(YAML) && CFLAGS=-fpic ./configure && make

CLEANFILES=_lib

include $(GOROOT)/src/Make.pkg
