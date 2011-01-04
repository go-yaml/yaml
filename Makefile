include $(GOROOT)/src/Make.inc

TARG=goyaml

GOFILES=\
	goyaml.go\

CGOFILES=\
	parser.go\

LIBYAML=/usr/lib/libyaml.a
LIBYAML_OFILES=$(shell ar t $(LIBYAML))

CGO_LDFLAGS+=-lm -lpthread
CGO_OFILES+=$(LIBYAML_OFILES:%=_lib/%)

$(CGO_OFILES): $(LIBYAML)
	@mkdir _lib
	cd _lib && ar x $(LIBYAML)

CLEANFILES=_lib

include $(GOROOT)/src/Make.pkg

#_cgo_defun.c: helpers.c
