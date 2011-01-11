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
CGO_CFLAGS+=\
	-I$(PWD) \
	-DYAML_VERSION_STRING='"0.1.3"' \
	-DYAML_VERSION_MAJOR=0 \
	-DYAML_VERSION_MINOR=1 \
	-DYAML_VERSION_PATCH=3 \

include $(GOROOT)/src/Make.pkg
