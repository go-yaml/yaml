package goyaml

// #include "helpers.h"
import "C"

import (
    "unsafe"
    "reflect"
    "strconv"
)

const (
    documentNode = 1 << iota
    mappingNode
    sequenceNode
    scalarNode
    aliasNode
)

type node struct {
    kind int
    line, column int
    tag string
    value string
    implicit bool
    children []*node
    anchors map[string]*node
}

func GoYString(s *C.yaml_char_t) string {
    return C.GoString((*C.char)(unsafe.Pointer(s)))
}


// ----------------------------------------------------------------------------
// Parser, produces a node tree out of a libyaml event stream.

type parser struct {
    parser C.yaml_parser_t
    event C.yaml_event_t
    doc *node
}

func newParser(b []byte) *parser {
    p := parser{}
    if C.yaml_parser_initialize(&p.parser) == 0 {
        panic("Failed to initialize YAML emitter")
    }

    if len(b) == 0 {
        b = []byte{'\n'}
    }

    // How unsafe is this really?  Will this break if the GC becomes compacting?
    // Probably not, otherwise that would likely break &parse below as well.
    input := (*C.uchar)(unsafe.Pointer(&b[0]))
    C.yaml_parser_set_input_string(&p.parser, input, (C.size_t)(len(b)))

    p.skip()
    if p.event._type != C.YAML_STREAM_START_EVENT {
        panic("Expected stream start event, got " +
              strconv.Itoa(int(p.event._type)))
    }
    p.skip()
    return &p
}

func (p *parser) destroy() {
    if p.event._type != C.YAML_NO_EVENT {
        C.yaml_event_delete(&p.event)
    }
    C.yaml_parser_delete(&p.parser)
}

func (p *parser) skip() {
    if p.event._type != C.YAML_NO_EVENT {
        if p.event._type == C.YAML_STREAM_END_EVENT {
            panic("Attempted to go past the end of stream. Corrupted value?")
        }
        C.yaml_event_delete(&p.event)
    }
    if C.yaml_parser_parse(&p.parser, &p.event) == 0 {
        p.fail()
    }
}

func (p *parser) fail() {
    var where string
    var line int
    if p.parser.problem_mark.line != 0 {
        line = int(C.int(p.parser.problem_mark.line))
    } else if p.parser.context_mark.line != 0 {
        line = int(C.int(p.parser.context_mark.line))
    }
    if line != 0 {
        where = "line " + strconv.Itoa(line) + ": "
    }
    var msg string
    if p.parser.problem != nil {
        msg = C.GoString(p.parser.problem)
    } else {
        msg = "Unknown problem parsing YAML content"
    }
    panic(where + msg)
}

func (p *parser) anchor(n *node, anchor *C.yaml_char_t) {
    if anchor != nil {
        p.doc.anchors[GoYString(anchor)] = n
    }
}

func (p *parser) parse() *node {
    switch p.event._type {
    case C.YAML_SCALAR_EVENT:
        return p.scalar()
    case C.YAML_ALIAS_EVENT:
        return p.alias()
    case C.YAML_MAPPING_START_EVENT:
        return p.mapping()
    case C.YAML_SEQUENCE_START_EVENT:
        return p.sequence()
    case C.YAML_DOCUMENT_START_EVENT:
        return p.document()
    case C.YAML_STREAM_END_EVENT:
        // Happens when attempting to decode an empty buffer.
        return nil
    default:
        panic("Attempted to parse unknown event: " +
              strconv.Itoa(int(p.event._type)))
    }
    panic("Unreachable")
}

func (p *parser) node(kind int) *node {
    return &node{kind: kind,
                 line: int(C.int(p.event.start_mark.line)),
                 column: int(C.int(p.event.start_mark.column))}
}

func (p *parser) document() *node {
    n := p.node(documentNode)
    n.anchors = make(map[string]*node)
    p.doc = n
    p.skip()
    n.children = append(n.children, p.parse())
    if p.event._type != C.YAML_DOCUMENT_END_EVENT {
        panic("Expected end of document event but got " +
              strconv.Itoa(int(p.event._type)))
    }
    p.skip()
    return n
}

func (p *parser) alias() *node {
    alias := C.event_alias(&p.event)
    n := p.node(aliasNode)
    n.value = GoYString(alias.anchor)
    p.skip()
    return n
}

func (p *parser) scalar() *node {
    scalar := C.event_scalar(&p.event)
    n := p.node(scalarNode)
    n.value = GoYString(scalar.value)
    n.tag = GoYString(scalar.tag)
    n.implicit = (scalar.plain_implicit != 0)
    p.anchor(n, scalar.anchor)
    p.skip()
    return n
}

func (p *parser) sequence() *node {
    n := p.node(sequenceNode)
    p.anchor(n, C.event_sequence_start(&p.event).anchor)
    p.skip()
    for p.event._type != C.YAML_SEQUENCE_END_EVENT {
        n.children = append(n.children, p.parse())
    }
    p.skip()
    return n
}

func (p *parser) mapping() *node {
    n := p.node(mappingNode)
    p.anchor(n, C.event_mapping_start(&p.event).anchor)
    p.skip()
    for p.event._type != C.YAML_MAPPING_END_EVENT {
        n.children = append(n.children, p.parse(), p.parse())
    }
    p.skip()
    return n
}


// ----------------------------------------------------------------------------
// Decoder, unmarshals a node into a provided value.

type decoder struct {
    doc *node
    aliases map[string]bool
}

func newDecoder() *decoder {
    d := &decoder{}
    d.aliases = make(map[string]bool)
    return d
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

func (d *decoder) unmarshal(n *node, out reflect.Value) (good bool) {
    switch n.kind {
    case documentNode:
        good = d.document(n, out)
    case scalarNode:
        good = d.scalar(n, out)
    case aliasNode:
        good = d.alias(n, out)
    case mappingNode:
        good = d.mapping(n, out)
    case sequenceNode:
        good = d.sequence(n, out)
    default:
        panic("Internal error: unknown node kind: " + strconv.Itoa(n.kind))
    }
    return
}

func (d *decoder) document(n *node, out reflect.Value) (good bool) {
    if len(n.children) == 1 {
        d.doc = n
        d.unmarshal(n.children[0], out)
        return true
    }
    return false
}

func (d *decoder) alias(n *node, out reflect.Value) (good bool) {
    an, ok := d.doc.anchors[n.value]
    if !ok {
        panic("Unknown anchor '" + n.value + "' referenced")
    }
    if d.aliases[n.value] {
        panic("Anchor '" + n.value + "' value contains itself")
    }
    d.aliases[n.value] = true
    good = d.unmarshal(an, out)
    d.aliases[n.value] = false, false
    return good
}

func (d *decoder) scalar(n *node, out reflect.Value) (good bool) {
    var tag string
    var resolved interface{}
    if n.tag == "" && !n.implicit {
        resolved = n.value
    } else {
        tag, resolved = resolve(n.tag, n.value)
        if set := d.setter(tag, &out, &good); set != nil {
            defer set()
        }
    }
    switch out := out.(type) {
    case *reflect.StringValue:
        out.Set(n.value)
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
        panic("Can't handle type yet: " + out.Type().String())
    }
    return good
}

func (d *decoder) sequence(n *node, out reflect.Value) (good bool) {
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
        return false
    }
    st := sv.Type().(*reflect.SliceType)
    et := st.Elem()

    l := len(n.children)
    for i := 0; i < l; i++ {
        e := reflect.MakeZero(et)
        if ok := d.unmarshal(n.children[i], e); ok {
            sv.SetValue(reflect.Append(sv, e))
        }
    }
    return true
}

func (d *decoder) mapping(n *node, out reflect.Value) (good bool) {
    if set := d.setter("!!map", &out, &good); set != nil {
        defer set()
    }
    if s, ok := out.(*reflect.StructValue); ok {
        return d.mappingStruct(n, s)
    }

    if iface, ok := out.(*reflect.InterfaceValue); ok {
        // No type hints. Will have to use a generic map.
        out = reflect.NewValue(make(map[interface{}]interface{}))
        iface.SetValue(out)
    }

    mv, ok := out.(*reflect.MapValue)
    if !ok {
        return false
    }
    mt := mv.Type().(*reflect.MapType)
    kt := mt.Key()
    et := mt.Elem()

    l := len(n.children)
    for i := 0; i < l; i += 2 {
        k := reflect.MakeZero(kt)
        if d.unmarshal(n.children[i], k) {
            e := reflect.MakeZero(et)
            if d.unmarshal(n.children[i+1], e) {
                mv.SetElem(k, e)
            }
        }
    }
    return true
}

func (d *decoder) mappingStruct(n *node, out *reflect.StructValue) (good bool) {
    fields, err := getStructFields(out.Type().(*reflect.StructType))
    if err != nil {
        panic(err)
    }
    name := reflect.NewValue("").(*reflect.StringValue)
    fieldsMap := fields.Map
    l := len(n.children)
    for i := 0; i < l; i += 2 {
        if !d.unmarshal(n.children[i], name) {
            continue
        }
        if info, ok := fieldsMap[name.Get()]; ok {
            d.unmarshal(n.children[i+1], out.Field(info.Num))
        }
    }
    return true
}
