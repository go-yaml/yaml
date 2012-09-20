package goyaml

// #include "helpers.h"
import "C"

import (
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

type encoder struct {
	emitter C.yaml_emitter_t
	event   C.yaml_event_t
	out     []byte
	tmp     []byte
	tmph    *reflect.SliceHeader
	flow    bool
}

//export outputHandler
func outputHandler(data unsafe.Pointer, buffer *C.uchar, size C.size_t) C.int {
	e := (*encoder)(data)
	e.tmph.Data = uintptr(unsafe.Pointer(buffer))
	e.tmph.Len = int(size)
	e.tmph.Cap = int(size)
	e.out = append(e.out, e.tmp...)
	return 1
}

func newEncoder() (e *encoder) {
	e = &encoder{}
	e.tmph = (*reflect.SliceHeader)(unsafe.Pointer(&e.tmp))
	if C.yaml_emitter_initialize(&e.emitter) == 0 {
		panic("Failed to initialize YAML emitter")
	}
	C.set_output_handler(&e.emitter)
	C.yaml_stream_start_event_initialize(&e.event, C.YAML_UTF8_ENCODING)
	e.emit()
	C.yaml_document_start_event_initialize(&e.event, nil, nil, nil, 1)
	e.emit()
	return e
}

func (e *encoder) finish() {
	C.yaml_document_end_event_initialize(&e.event, 1)
	e.emit()
	e.emitter.open_ended = 0
	C.yaml_stream_end_event_initialize(&e.event)
	e.emit()
}

func (e *encoder) destroy() {
	C.yaml_emitter_delete(&e.emitter)
}

func (e *encoder) emit() {
	// This will internally delete the e.event value.
	if C.yaml_emitter_emit(&e.emitter, &e.event) == 0 &&
		e.event._type != C.YAML_DOCUMENT_END_EVENT &&
		e.event._type != C.YAML_STREAM_END_EVENT {
		if e.emitter.error == C.YAML_EMITTER_ERROR {
			// XXX TESTME
			panic("YAML emitter error: " + C.GoString(e.emitter.problem))
		} else {
			// XXX TESTME
			panic("Unknown YAML emitter error")
		}
	}
}

func (e *encoder) fail(msg string) {
	if msg == "" {
		if e.emitter.problem != nil {
			msg = C.GoString(e.emitter.problem)
		} else {
			msg = "Unknown problem generating YAML content"
		}
	}
	panic(msg)
}

func (e *encoder) marshal(tag string, in reflect.Value) {
	var value interface{}
	if getter, ok := in.Interface().(Getter); ok {
		tag, value = getter.GetYAML()
		if value == nil {
			e.nilv()
			return
		}
		in = reflect.ValueOf(value)
	}
	switch in.Kind() {
	case reflect.Interface:
		if in.IsNil() {
			e.nilv()
		} else {
			e.marshal(tag, in.Elem())
		}
	case reflect.Map:
		e.mapv(tag, in)
	case reflect.Ptr:
		if in.IsNil() {
			e.nilv()
		} else {
			e.marshal(tag, in.Elem())
		}
	case reflect.Struct:
		e.structv(tag, in)
	case reflect.Slice:
		e.slicev(tag, in)
	case reflect.String:
		e.stringv(tag, in)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.intv(tag, in)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.uintv(tag, in)
	case reflect.Float32, reflect.Float64:
		e.floatv(tag, in)
	case reflect.Bool:
		e.boolv(tag, in)
	default:
		panic("Can't marshal type yet: " + in.Type().String())
	}
}

func (e *encoder) mapv(tag string, in reflect.Value) {
	e.mappingv(tag, func() {
		keys := keyList(in.MapKeys())
		sort.Sort(keys)
		for _, k := range keys {
			e.marshal("", k)
			e.marshal("", in.MapIndex(k))
		}
	})
}

func (e *encoder) structv(tag string, in reflect.Value) {
	fields, err := getStructFields(in.Type())
	if err != nil {
		panic(err)
	}
	e.mappingv(tag, func() {
		for i, info := range fields.List {
			value := in.Field(i)
			if info.OmitEmpty && isZero(value) {
				continue
			}
			e.marshal("", reflect.ValueOf(info.Key))
			e.flow = info.Flow
			e.marshal("", value)
		}
	})
}

func (e *encoder) mappingv(tag string, f func()) {
	var ctag *C.yaml_char_t
	var free func()
	cimplicit := C.int(1)
	if tag != "" {
		ctag, free = ystr(tag)
		defer free()
		cimplicit = 0
	}
	cstyle := C.yaml_mapping_style_t(C.YAML_BLOCK_MAPPING_STYLE)
	if e.flow {
		e.flow = false
		cstyle = C.YAML_FLOW_MAPPING_STYLE
	}
	C.yaml_mapping_start_event_initialize(&e.event, nil, ctag, cimplicit,
		cstyle)
	e.emit()
	f()
	C.yaml_mapping_end_event_initialize(&e.event)
	e.emit()
}

func (e *encoder) slicev(tag string, in reflect.Value) {
	var ctag *C.yaml_char_t
	var free func()
	var cimplicit C.int
	if tag != "" {
		ctag, free = ystr(tag)
		defer free()
		cimplicit = 0
	} else {
		cimplicit = 1
	}

	cstyle := C.yaml_sequence_style_t(C.YAML_BLOCK_SEQUENCE_STYLE)
	if e.flow {
		e.flow = false
		cstyle = C.YAML_FLOW_SEQUENCE_STYLE
	}
	C.yaml_sequence_start_event_initialize(&e.event, nil, ctag, cimplicit,
		cstyle)
	e.emit()
	n := in.Len()
	for i := 0; i < n; i++ {
		e.marshal("", in.Index(i))
	}
	C.yaml_sequence_end_event_initialize(&e.event)
	e.emit()
}

func (e *encoder) stringv(tag string, in reflect.Value) {
	var style C.yaml_scalar_style_t
	s := in.String()
	if rtag, _ := resolve("", s); rtag != "!!str" {
		style = C.YAML_DOUBLE_QUOTED_SCALAR_STYLE
	} else {
		style = C.YAML_PLAIN_SCALAR_STYLE
	}
	e.emitScalar(s, "", tag, style)
}

func (e *encoder) boolv(tag string, in reflect.Value) {
	var s string
	if in.Bool() {
		s = "true"
	} else {
		s = "false"
	}
	e.emitScalar(s, "", tag, C.YAML_PLAIN_SCALAR_STYLE)
}

func (e *encoder) intv(tag string, in reflect.Value) {
	s := strconv.FormatInt(in.Int(), 10)
	e.emitScalar(s, "", tag, C.YAML_PLAIN_SCALAR_STYLE)
}

func (e *encoder) uintv(tag string, in reflect.Value) {
	s := strconv.FormatUint(in.Uint(), 10)
	e.emitScalar(s, "", tag, C.YAML_PLAIN_SCALAR_STYLE)
}

func (e *encoder) floatv(tag string, in reflect.Value) {
	// FIXME: Handle 64 bits here.
	s := strconv.FormatFloat(float64(in.Float()), 'g', -1, 32)
	switch s {
	case "+Inf":
		s = ".inf"
	case "-Inf":
		s = "-.inf"
	case "NaN":
		s = ".nan"
	}
	e.emitScalar(s, "", tag, C.YAML_PLAIN_SCALAR_STYLE)
}

func (e *encoder) nilv() {
	e.emitScalar("null", "", "", C.YAML_PLAIN_SCALAR_STYLE)
}

func (e *encoder) emitScalar(value, anchor, tag string, style C.yaml_scalar_style_t) {
	var canchor, ctag, cvalue *C.yaml_char_t
	var cimplicit C.int
	var free func()
	if anchor != "" {
		canchor, free = ystr(anchor)
		defer free()
	}
	if tag != "" {
		ctag, free = ystr(tag)
		defer free()
		cimplicit = 0
		style = C.YAML_PLAIN_SCALAR_STYLE
	} else {
		cimplicit = 1
	}
	cvalue, free = ystr(value)
	defer free()
	size := C.int(len(value))
	if C.yaml_scalar_event_initialize(&e.event, canchor, ctag, cvalue, size,
		cimplicit, cimplicit, style) == 0 {
		e.fail("")
	}
	e.emit()
}

func ystr(s string) (ys *C.yaml_char_t, free func()) {
	up := unsafe.Pointer(C.CString(s))
	ys = (*C.yaml_char_t)(up)
	free = func() { C.free(up) }
	return ys, free
}
