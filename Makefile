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

CGO_OFILES+=\
	helpers.o\
	api.o\
	scanner.o\
	reader.o\
	parser.o\
	writer.o\
	emitter.o\

CGO_LDFLAGS+=-lm -lpthread
CGO_CFLAGS+=-I$(PWD) -DHAVE_CONFIG_H=1

include $(GOROOT)/src/Make.pkg
