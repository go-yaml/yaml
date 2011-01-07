package goyaml

/* #include "helpers.c" */
import "C"

import (
    "unsafe"
    "reflect"
    "strconv"
)


type decoder struct {
    parser *C.yaml_parser_t
    event *C.yaml_event_t
}

func newDecoder(b []byte) *decoder {
    if len(b) == 0 {
        panic("Can't handle empty buffers yet") // XXX Fix this.
    }

    d := decoder{}
    d.event = &C.yaml_event_t{}
    d.parser = &C.yaml_parser_t{}
    C.yaml_parser_initialize(d.parser)

    // How unsafe is this really?  Will this break if the GC becomes compacting?
    // Probably not, otherwise that would likely break &parse below as well.
    input := (*C.uchar)(unsafe.Pointer(&b[0]))
    C.yaml_parser_set_input_string(d.parser, input, (C.size_t)(len(b)))

    d.next()
    if d.event._type != C.YAML_STREAM_START_EVENT {
        panic("Expected stream start event, got " +
              strconv.Itoa(int(d.event._type)))
    }
    d.next()
    return &d
}

func (d *decoder) destroy() {
    if d.event._type != C.YAML_NO_EVENT {
        C.yaml_event_delete(d.event)
    }
    C.yaml_parser_delete(d.parser)
}

func (d *decoder) next() {
    if d.event._type != C.YAML_NO_EVENT {
        if d.event._type == C.YAML_STREAM_END_EVENT {
            panic("Attempted to go past the end of stream. Corrupted value?")
        }
        C.yaml_event_delete(d.event)
    }
    if C.yaml_parser_parse(d.parser, d.event) == 0 {
        panic("Parsing failed.") // XXX Need better error handling here.
    }
}

func (d *decoder) skip(_type C.yaml_event_type_t) {
    for d.event._type != _type {
        d.next()
    }
    d.next()
}

var blackHole = reflect.NewValue(true)

func (d *decoder) drop() {
    d.unmarshal(blackHole)
}

// d.setter deals with setters and pointer dereferencing and initialization.
//
// It's a slightly convoluted case to handle properly:
//
// - Nil pointers should be zeroed out, unless being set to nil
// - We don't know at this point yet what's the value to SetYAML() with.
// - We can't separate pointer deref/init and setter checking, because
//   a setter may be found while going down a pointer chain.
//
// Thus, here is how it takes care of it:
//
// - out is provided as a pointer, so that it can be replaced.
// - when looking at a non-setter ptr, *out=ptr.Elem(), unless tag=!!null
// - when a setter is found, *out=interface{}, and a set() function is
//   returned to call SetYAML() with the value of *out once it's defined.
//
func (d *decoder) setter(tag string, out *reflect.Value, good *bool) (set func()) {
    again := true
    for again {
        again = false
        setter, _ := (*out).Interface().(Setter)
        if tag != "!!null" || setter != nil {
            if pv, ok := (*out).(*reflect.PtrValue); ok {
                if pv.IsNil() {
                    *out = reflect.MakeZero(pv.Type().(*reflect.PtrType).Elem())
                    pv.PointTo(*out)
                } else {
                    *out = pv.Elem()
                }
                setter, _ = pv.Interface().(Setter)
                again = true
            }
        }
        if setter != nil {
            var arg interface{}
            *out = reflect.NewValue(&arg).(*reflect.PtrValue).Elem()
            return func() {
                *good = setter.SetYAML(tag, arg)
            }
        }
    }
    return nil
}

func (d *decoder) unmarshal(out reflect.Value) (good bool) {
    switch d.event._type {
    case C.YAML_SCALAR_EVENT:
        good = d.scalar(out)
    case C.YAML_MAPPING_START_EVENT:
        good = d.mapping(out)
    case C.YAML_SEQUENCE_START_EVENT:
        good = d.sequence(out)
    case C.YAML_DOCUMENT_START_EVENT:
        good = d.document(out)
    default:
        panic("Attempted to unmarshal unexpected event: " +
              strconv.Itoa(int(d.event._type)))
    }
    return
}

func (d *decoder) document(out reflect.Value) bool {
    d.next()
    result := d.unmarshal(out)
    if d.event._type != C.YAML_DOCUMENT_END_EVENT {
        panic("Expected end of document event but got " +
              strconv.Itoa(int(d.event._type)))
    }
    d.next()
    return result
}

func (d *decoder) scalar(out reflect.Value) (good bool) {
    scalar := C.event_scalar(d.event)
    str := GoYString(scalar.value)
    tag := GoYString(scalar.tag)
    var resolved interface{}
    if tag == "" && scalar.plain_implicit == 0 {
        resolved = str
    } else {
        tag, resolved = resolve(tag, str)
        if set := d.setter(tag, &out, &good); set != nil {
            defer set()
        }
    }
    switch out := out.(type) {
    case *reflect.StringValue:
        out.Set(str)
        good = true
    case *reflect.InterfaceValue:
        out.Set(reflect.NewValue(resolved))
        good = true
    case *reflect.IntValue:
        switch resolved := resolved.(type) {
        case int:
            if !out.Overflow(int64(resolved)) {
                out.Set(int64(resolved))
                good = true
            }
        case int64:
            if !out.Overflow(resolved) {
                out.Set(resolved)
                good = true
            }
        }
    case *reflect.UintValue:
        switch resolved := resolved.(type) {
        case int:
            if resolved >= 0 {
                out.Set(uint64(resolved))
                good = true
            }
        case int64:
            if resolved >= 0 {
                out.Set(uint64(resolved))
                good = true
            }
        }
    case *reflect.BoolValue:
        switch resolved := resolved.(type) {
        case bool:
            out.Set(resolved)
            good = true
        }
    case *reflect.FloatValue:
        switch resolved := resolved.(type) {
        case float:
            out.Set(float64(resolved))
            good = true
        }
    case *reflect.PtrValue:
        switch resolved := resolved.(type) {
        case nil:
            out.PointTo(nil)
            good = true
        }
    default:
        panic("Can't handle scalar type yet: " + out.Type().String())
    }
    d.next()
    return good
}

func (d *decoder) sequence(out reflect.Value) (good bool) {
    if set := d.setter("!!seq", &out, &good); set != nil {
        defer set()
    }
    if iface, ok := out.(*reflect.InterfaceValue); ok {
        // No type hints. Will have to use a generic sequence.
        out = reflect.NewValue(make([]interface{}, 0))
        iface.SetValue(out)
    }

    sv, ok := out.(*reflect.SliceValue)
    if !ok {
        d.skip(C.YAML_SEQUENCE_END_EVENT)
        return false
    }
    st := sv.Type().(*reflect.SliceType)
    et := st.Elem()

    d.next()
    for d.event._type != C.YAML_SEQUENCE_END_EVENT {
        e := reflect.MakeZero(et)
        if ok := d.unmarshal(e); ok {
            sv.SetValue(reflect.Append(sv, e))
        }
    }
    d.next()
    return true
}

func (d *decoder) mapping(out reflect.Value) (good bool) {
    if set := d.setter("!!map", &out, &good); set != nil {
        defer set()
    }
    if s, ok := out.(*reflect.StructValue); ok {
        return d.mappingStruct(s)
    }

    if iface, ok := out.(*reflect.InterfaceValue); ok {
        // No type hints. Will have to use a generic map.
        out = reflect.NewValue(make(map[interface{}]interface{}))
        iface.SetValue(out)
    }

    mv, ok := out.(*reflect.MapValue)
    if !ok {
        d.skip(C.YAML_MAPPING_END_EVENT)
        return false
    }
    mt := mv.Type().(*reflect.MapType)
    kt := mt.Key()
    et := mt.Elem()

    d.next()
    for d.event._type != C.YAML_MAPPING_END_EVENT {
        k := reflect.MakeZero(kt)
        kok := d.unmarshal(k)
        e := reflect.MakeZero(et)
        eok := d.unmarshal(e)
        if kok && eok {
            mv.SetElem(k, e)
        }
    }
    d.next()
    return true
}

func (d *decoder) mappingStruct(out *reflect.StructValue) bool {
    fields, err := getStructFields(out.Type().(*reflect.StructType))
    if err != nil {
        panic(err)
    }
    name := reflect.NewValue("").(*reflect.StringValue)
    fieldsMap := fields.Map
    d.next()
    for d.event._type != C.YAML_MAPPING_END_EVENT {
        if d.unmarshal(name) {
            if info, ok := fieldsMap[name.Get()]; ok {
                d.unmarshal(out.Field(info.Num))
                continue
            }
        }
        // Can't unmarshal name, or it's not present in struct.
        d.drop()
    }
    d.next()
    return true
}

func GoYString(s *C.yaml_char_t) string {
    return C.GoString((*C.char)(unsafe.Pointer(s)))
}
