package yaml

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type NodeKind uint32

const (
	DocumentNode NodeKind = 1 << iota
	SequenceNode
	MappingNode
	ScalarNode
	AliasNode
)

type Style uint32

const (
	DoubleQuotedStyle Style = 1 << iota
	SingleQuotedStyle
	LiteralStyle
	FoldedStyle
	FlowStyle
)

type Node struct {
	Kind   NodeKind
	Style  Style
	Line   int
	Column int
	Tag    string
	Value  string
	// TODO Alias should probably be the string, and then perhaps have a hidden cache?
	Alias    *Node // Resolved alias for alias nodes.
	Anchors  map[string]*Node
	Children []*Node
	Header   string
	Inline   string
	Footer   string
}

func (n *Node) implicit() bool {
	return n.Style&(SingleQuotedStyle|DoubleQuotedStyle) == 0 && (n.Tag == "" || n.Tag == "!")
}

func (n *Node) ShortTag() string {
	tag := n.LongTag()
	if strings.HasPrefix(tag, longTagPrefix) {
		return "!!" + tag[len(longTagPrefix):]
	}
	return tag
}

func (n *Node) LongTag() string {
	if n.Tag == "" || n.Tag == "!" {
		switch n.Kind {
		case MappingNode:
			return yaml_MAP_TAG
		case SequenceNode:
			return yaml_SEQ_TAG
		case AliasNode:
			if n.Alias != nil {
				return n.Alias.LongTag()
			}
		case ScalarNode:
			if n.Style&(SingleQuotedStyle|DoubleQuotedStyle) != 0 {
				return yaml_STR_TAG
			}
			tag, _ := resolve("", n.Value)
			return tag
		}
		return ""

	} else if strings.HasPrefix(n.Tag, "!!") {
		return longTagPrefix + n.Tag[2:]
	}
	return n.Tag
}

func (n *Node) SetString(s string) {
	n.Kind = ScalarNode
	n.Value = s
	if strings.Contains(s, "\n") {
		n.Style = LiteralStyle
	} else if n.LongTag() != "tag:yaml.org,2002:str" {
		n.Style = DoubleQuotedStyle
	}
}

// ----------------------------------------------------------------------------
// Parser, produces a node tree out of a libyaml event stream.

type parser struct {
	parser   yaml_parser_t
	event    yaml_event_t
	doc      *Node
	doneInit bool
}

func newParser(b []byte) *parser {
	p := parser{}
	if !yaml_parser_initialize(&p.parser) {
		panic("failed to initialize YAML emitter")
	}
	if len(b) == 0 {
		b = []byte{'\n'}
	}
	yaml_parser_set_input_string(&p.parser, b)
	return &p
}

func newParserFromReader(r io.Reader) *parser {
	p := parser{}
	if !yaml_parser_initialize(&p.parser) {
		panic("failed to initialize YAML emitter")
	}
	yaml_parser_set_input_reader(&p.parser, r)
	return &p
}

func (p *parser) init() {
	if p.doneInit {
		return
	}
	p.expect(yaml_STREAM_START_EVENT)
	p.doneInit = true
}

func (p *parser) destroy() {
	if p.event.typ != yaml_NO_EVENT {
		yaml_event_delete(&p.event)
	}
	yaml_parser_delete(&p.parser)
}

// expect consumes an event from the event stream and
// checks that it's of the expected type.
func (p *parser) expect(e yaml_event_type_t) {
	if p.event.typ == yaml_NO_EVENT {
		if !yaml_parser_parse(&p.parser, &p.event) {
			p.fail()
		}
	}
	if p.event.typ == yaml_STREAM_END_EVENT {
		failf("attempted to go past the end of stream; corrupted value?")
	}
	if p.event.typ != e {
		p.parser.problem = fmt.Sprintf("expected %s event but got %s", e, p.event.typ)
		p.fail()
	}
	yaml_event_delete(&p.event)
	p.event.typ = yaml_NO_EVENT
}

// peek peeks at the next event in the event stream,
// puts the results into p.event and returns the event type.
func (p *parser) peek() yaml_event_type_t {
	if p.event.typ != yaml_NO_EVENT {
		return p.event.typ
	}
	if !yaml_parser_parse(&p.parser, &p.event) {
		p.fail()
	}
	return p.event.typ
}

func (p *parser) fail() {
	var where string
	var line int
	if p.parser.problem_mark.line != 0 {
		line = p.parser.problem_mark.line
		// Scanner errors don't iterate line before returning error
		if p.parser.error == yaml_SCANNER_ERROR {
			line++
		}
	} else if p.parser.context_mark.line != 0 {
		line = p.parser.context_mark.line
	}
	if line != 0 {
		where = "line " + strconv.Itoa(line) + ": "
	}
	var msg string
	if len(p.parser.problem) > 0 {
		msg = p.parser.problem
	} else {
		msg = "unknown problem parsing YAML content"
	}
	failf("%s%s", where, msg)
}

func (p *parser) anchor(n *Node, anchor []byte) {
	if anchor != nil {
		p.doc.Anchors[string(anchor)] = n
	}
}

func (p *parser) parse() *Node {
	p.init()
	switch p.peek() {
	case yaml_SCALAR_EVENT:
		return p.scalar()
	case yaml_ALIAS_EVENT:
		return p.alias()
	case yaml_MAPPING_START_EVENT:
		return p.mapping()
	case yaml_SEQUENCE_START_EVENT:
		return p.sequence()
	case yaml_DOCUMENT_START_EVENT:
		return p.document()
	case yaml_STREAM_END_EVENT:
		// Happens when attempting to decode an empty buffer.
		return nil
	default:
		panic("attempted to parse unknown event: " + p.event.typ.String())
	}
}

func (p *parser) node(kind NodeKind) *Node {
	return &Node{
		Kind:   kind,
		Line:   p.event.start_mark.line + 1,
		Column: p.event.start_mark.column + 1,
		Header: string(p.event.header_comment),
		Inline: string(p.event.inline_comment),
		Footer: string(p.event.footer_comment),
	}
}

func (p *parser) parseChild(parent *Node) *Node {
	child := p.parse()
	parent.Children = append(parent.Children, child)
	return child
}

func (p *parser) document() *Node {
	n := p.node(DocumentNode)
	n.Anchors = make(map[string]*Node)
	p.doc = n
	p.expect(yaml_DOCUMENT_START_EVENT)
	p.parseChild(n)
	if p.peek() == yaml_DOCUMENT_END_EVENT {
		n.Footer = string(p.event.footer_comment)
	}
	p.expect(yaml_DOCUMENT_END_EVENT)
	return n
}

func (p *parser) alias() *Node {
	n := p.node(AliasNode)
	n.Value = string(p.event.anchor)
	n.Alias = p.doc.Anchors[n.Value]
	if n.Alias == nil {
		failf("unknown anchor '%s' referenced", n.Value)
	}
	p.expect(yaml_ALIAS_EVENT)
	return n
}

func (p *parser) scalar() *Node {
	n := p.node(ScalarNode)
	n.Value = string(p.event.value)
	n.Tag = string(p.event.tag)
	style := p.event.scalar_style()
	switch {
	case style&yaml_DOUBLE_QUOTED_SCALAR_STYLE != 0:
		n.Style = DoubleQuotedStyle
	case style&yaml_SINGLE_QUOTED_SCALAR_STYLE != 0:
		n.Style = SingleQuotedStyle
	case style&yaml_LITERAL_SCALAR_STYLE != 0:
		n.Style = LiteralStyle
	case style&yaml_FOLDED_SCALAR_STYLE != 0:
		n.Style = FoldedStyle
	}
	p.anchor(n, p.event.anchor)
	p.expect(yaml_SCALAR_EVENT)
	return n
}

func (p *parser) sequence() *Node {
	n := p.node(SequenceNode)
	n.Tag = string(p.event.tag)
	if p.event.sequence_style()&yaml_FLOW_SEQUENCE_STYLE != 0 {
		n.Style = FlowStyle
	}
	p.anchor(n, p.event.anchor)
	p.expect(yaml_SEQUENCE_START_EVENT)
	for p.peek() != yaml_SEQUENCE_END_EVENT {
		p.parseChild(n)
	}
	n.Inline = string(p.event.inline_comment)
	n.Footer = string(p.event.footer_comment)
	p.expect(yaml_SEQUENCE_END_EVENT)
	return n
}

func (p *parser) mapping() *Node {
	n := p.node(MappingNode)
	n.Tag = string(p.event.tag)
	if p.event.mapping_style()&yaml_FLOW_MAPPING_STYLE != 0 {
		n.Style |= FlowStyle
	}
	p.anchor(n, p.event.anchor)
	p.expect(yaml_MAPPING_START_EVENT)
	for p.peek() != yaml_MAPPING_END_EVENT {
		k := p.parseChild(n)
		v := p.parseChild(n)
		if v.Footer != "" {
			k.Footer = v.Footer
			v.Footer = ""
		}
	}
	n.Inline = string(p.event.inline_comment)
	n.Footer = string(p.event.footer_comment)
	p.expect(yaml_MAPPING_END_EVENT)
	return n
}

// ----------------------------------------------------------------------------
// Decoder, unmarshals a node into a provided value.

type decoder struct {
	doc     *Node
	aliases map[*Node]bool
	mapType reflect.Type
	terrors []string
	strict  bool
}

var (
	nodeType       = reflect.TypeOf(Node{})
	mapItemType    = reflect.TypeOf(MapItem{})
	durationType   = reflect.TypeOf(time.Duration(0))
	defaultMapType = reflect.TypeOf(map[interface{}]interface{}{})
	ifaceType      = defaultMapType.Elem()
	timeType       = reflect.TypeOf(time.Time{})
	ptrTimeType    = reflect.TypeOf(&time.Time{})
)

func newDecoder(strict bool) *decoder {
	d := &decoder{mapType: defaultMapType, strict: strict}
	d.aliases = make(map[*Node]bool)
	return d
}

func (d *decoder) terror(n *Node, tag string, out reflect.Value) {
	if n.Tag != "" {
		tag = n.Tag
	}
	value := n.Value
	if tag != yaml_SEQ_TAG && tag != yaml_MAP_TAG {
		if len(value) > 10 {
			value = " `" + value[:7] + "...`"
		} else {
			value = " `" + value + "`"
		}
	}
	d.terrors = append(d.terrors, fmt.Sprintf("line %d: cannot unmarshal %s%s into %s", n.Line, shortTag(tag), value, out.Type()))
}

func (d *decoder) callUnmarshaler(n *Node, u Unmarshaler) (good bool) {
	terrlen := len(d.terrors)
	err := u.UnmarshalYAML(func(v interface{}) (err error) {
		defer handleErr(&err)
		d.unmarshal(n, reflect.ValueOf(v))
		if len(d.terrors) > terrlen {
			issues := d.terrors[terrlen:]
			d.terrors = d.terrors[:terrlen]
			return &TypeError{issues}
		}
		return nil
	})
	if e, ok := err.(*TypeError); ok {
		d.terrors = append(d.terrors, e.Errors...)
		return false
	}
	if err != nil {
		fail(err)
	}
	return true
}

// d.prepare initializes and dereferences pointers and calls UnmarshalYAML
// if a value is found to implement it.
// It returns the initialized and dereferenced out value, whether
// unmarshalling was already done by UnmarshalYAML, and if so whether
// its types unmarshalled appropriately.
//
// If n holds a null value, prepare returns before doing anything.
func (d *decoder) prepare(n *Node, out reflect.Value) (newout reflect.Value, unmarshaled, good bool) {
	if n.Tag == yaml_NULL_TAG || n.Kind == ScalarNode && n.Tag == "" && (n.Value == "null" || n.Value == "~" || n.Value == "" && n.implicit()) {
		return out, false, false
	}
	again := true
	for again {
		again = false
		if out.Kind() == reflect.Ptr {
			if out.IsNil() {
				out.Set(reflect.New(out.Type().Elem()))
			}
			out = out.Elem()
			again = true
		}
		if out.CanAddr() {
			if u, ok := out.Addr().Interface().(Unmarshaler); ok {
				good = d.callUnmarshaler(n, u)
				return out, true, good
			}
		}
	}
	return out, false, false
}

func (d *decoder) unmarshal(n *Node, out reflect.Value) (good bool) {
	if out.Type() == nodeType {
		out.Set(reflect.ValueOf(n).Elem())
		return true
	}
	switch n.Kind {
	case DocumentNode:
		return d.document(n, out)
	case AliasNode:
		return d.alias(n, out)
	}
	out, unmarshaled, good := d.prepare(n, out)
	if unmarshaled {
		return good
	}
	switch n.Kind {
	case ScalarNode:
		good = d.scalar(n, out)
	case MappingNode:
		good = d.mapping(n, out)
	case SequenceNode:
		good = d.sequence(n, out)
	default:
		panic("internal error: unknown node kind: " + strconv.Itoa(int(n.Kind)))
	}
	return good
}

func (d *decoder) document(n *Node, out reflect.Value) (good bool) {
	if len(n.Children) == 1 {
		d.doc = n
		d.unmarshal(n.Children[0], out)
		return true
	}
	return false
}

func (d *decoder) alias(n *Node, out reflect.Value) (good bool) {
	if d.aliases[n] {
		// TODO this could actually be allowed in some circumstances.
		failf("anchor '%s' value contains itself", n.Value)
	}
	d.aliases[n] = true
	good = d.unmarshal(n.Alias, out)
	delete(d.aliases, n)
	return good
}

var zeroValue reflect.Value

func resetMap(out reflect.Value) {
	for _, k := range out.MapKeys() {
		out.SetMapIndex(k, zeroValue)
	}
}

func (d *decoder) scalar(n *Node, out reflect.Value) bool {
	var tag string
	var resolved interface{}
	if n.Tag == "" && !n.implicit() {
		tag = yaml_STR_TAG
		resolved = n.Value
	} else {
		tag, resolved = resolve(n.Tag, n.Value)
		if tag == yaml_BINARY_TAG {
			data, err := base64.StdEncoding.DecodeString(resolved.(string))
			if err != nil {
				failf("!!binary value contains invalid base64 data")
			}
			resolved = string(data)
		}
	}
	if resolved == nil {
		if out.Kind() == reflect.Map && !out.CanAddr() {
			resetMap(out)
		} else {
			out.Set(reflect.Zero(out.Type()))
		}
		return true
	}
	if resolvedv := reflect.ValueOf(resolved); out.Type() == resolvedv.Type() {
		// We've resolved to exactly the type we want, so use that.
		out.Set(resolvedv)
		return true
	}
	// Perhaps we can use the value as a TextUnmarshaler to
	// set its value.
	if out.CanAddr() {
		u, ok := out.Addr().Interface().(encoding.TextUnmarshaler)
		if ok {
			var text []byte
			if tag == yaml_BINARY_TAG {
				text = []byte(resolved.(string))
			} else {
				// We let any value be unmarshaled into TextUnmarshaler.
				// That might be more lax than we'd like, but the
				// TextUnmarshaler itself should bowl out any dubious values.
				text = []byte(n.Value)
			}
			err := u.UnmarshalText(text)
			if err != nil {
				fail(err)
			}
			return true
		}
	}
	switch out.Kind() {
	case reflect.String:
		if tag == yaml_BINARY_TAG {
			out.SetString(resolved.(string))
			return true
		}
		if resolved != nil {
			out.SetString(n.Value)
			return true
		}
	case reflect.Interface:
		if resolved == nil {
			out.Set(reflect.Zero(out.Type()))
		} else {
			out.Set(reflect.ValueOf(resolved))
		}
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch resolved := resolved.(type) {
		case int:
			if !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				return true
			}
		case int64:
			if !out.OverflowInt(resolved) {
				out.SetInt(resolved)
				return true
			}
		case uint64:
			if resolved <= math.MaxInt64 && !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				return true
			}
		case float64:
			if resolved <= math.MaxInt64 && !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				return true
			}
		case string:
			if out.Type() == durationType {
				d, err := time.ParseDuration(resolved)
				if err == nil {
					out.SetInt(int64(d))
					return true
				}
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch resolved := resolved.(type) {
		case int:
			if resolved >= 0 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				return true
			}
		case int64:
			if resolved >= 0 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				return true
			}
		case uint64:
			if !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				return true
			}
		case float64:
			if resolved <= math.MaxUint64 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				return true
			}
		}
	case reflect.Bool:
		switch resolved := resolved.(type) {
		case bool:
			out.SetBool(resolved)
			return true
		}
	case reflect.Float32, reflect.Float64:
		switch resolved := resolved.(type) {
		case int:
			out.SetFloat(float64(resolved))
			return true
		case int64:
			out.SetFloat(float64(resolved))
			return true
		case uint64:
			out.SetFloat(float64(resolved))
			return true
		case float64:
			out.SetFloat(resolved)
			return true
		}
	case reflect.Struct:
		if resolvedv := reflect.ValueOf(resolved); out.Type() == resolvedv.Type() {
			out.Set(resolvedv)
			return true
		}
	case reflect.Ptr:
		if out.Type().Elem() == reflect.TypeOf(resolved) {
			// TODO DOes this make sense? When is out a Ptr except when decoding a nil value?
			elem := reflect.New(out.Type().Elem())
			elem.Elem().Set(reflect.ValueOf(resolved))
			out.Set(elem)
			return true
		}
	}
	d.terror(n, tag, out)
	return false
}

func settableValueOf(i interface{}) reflect.Value {
	v := reflect.ValueOf(i)
	sv := reflect.New(v.Type()).Elem()
	sv.Set(v)
	return sv
}

func (d *decoder) sequence(n *Node, out reflect.Value) (good bool) {
	l := len(n.Children)

	var iface reflect.Value
	switch out.Kind() {
	case reflect.Slice:
		out.Set(reflect.MakeSlice(out.Type(), l, l))
	case reflect.Array:
		if l != out.Len() {
			failf("invalid array: want %d elements but got %d", out.Len(), l)
		}
	case reflect.Interface:
		// No type hints. Will have to use a generic sequence.
		iface = out
		out = settableValueOf(make([]interface{}, l))
	default:
		d.terror(n, yaml_SEQ_TAG, out)
		return false
	}
	et := out.Type().Elem()

	j := 0
	for i := 0; i < l; i++ {
		e := reflect.New(et).Elem()
		if ok := d.unmarshal(n.Children[i], e); ok {
			out.Index(j).Set(e)
			j++
		}
	}
	if out.Kind() != reflect.Array {
		out.Set(out.Slice(0, j))
	}
	if iface.IsValid() {
		iface.Set(out)
	}
	return true
}

func (d *decoder) mapping(n *Node, out reflect.Value) (good bool) {
	switch out.Kind() {
	case reflect.Struct:
		return d.mappingStruct(n, out)
	case reflect.Slice:
		return d.mappingSlice(n, out)
	case reflect.Map:
		// okay
	case reflect.Interface:
		if d.mapType.Kind() == reflect.Map {
			iface := out
			out = reflect.MakeMap(d.mapType)
			iface.Set(out)
		} else {
			slicev := reflect.New(d.mapType).Elem()
			if !d.mappingSlice(n, slicev) {
				return false
			}
			out.Set(slicev)
			return true
		}
	default:
		d.terror(n, yaml_MAP_TAG, out)
		return false
	}
	outt := out.Type()
	kt := outt.Key()
	et := outt.Elem()

	mapType := d.mapType
	if outt.Key() == ifaceType && outt.Elem() == ifaceType {
		d.mapType = outt
	}

	if out.IsNil() {
		out.Set(reflect.MakeMap(outt))
	}
	l := len(n.Children)
	for i := 0; i < l; i += 2 {
		if isMerge(n.Children[i]) {
			d.merge(n.Children[i+1], out)
			continue
		}
		k := reflect.New(kt).Elem()
		if d.unmarshal(n.Children[i], k) {
			kkind := k.Kind()
			if kkind == reflect.Interface {
				kkind = k.Elem().Kind()
			}
			if kkind == reflect.Map || kkind == reflect.Slice {
				failf("invalid map key: %#v", k.Interface())
			}
			e := reflect.New(et).Elem()
			if d.unmarshal(n.Children[i+1], e) {
				d.setMapIndex(n.Children[i+1], out, k, e)
			}
		}
	}
	d.mapType = mapType
	return true
}

func (d *decoder) setMapIndex(n *Node, out, k, v reflect.Value) {
	if d.strict && out.MapIndex(k) != zeroValue {
		d.terrors = append(d.terrors, fmt.Sprintf("line %d: key %#v already set in map", n.Line, k.Interface()))
		return
	}
	out.SetMapIndex(k, v)
}

func (d *decoder) mappingSlice(n *Node, out reflect.Value) (good bool) {
	outt := out.Type()
	if outt.Elem() != mapItemType {
		d.terror(n, yaml_MAP_TAG, out)
		return false
	}

	mapType := d.mapType
	d.mapType = outt

	var slice []MapItem
	var l = len(n.Children)
	for i := 0; i < l; i += 2 {
		if isMerge(n.Children[i]) {
			d.merge(n.Children[i+1], out)
			continue
		}
		item := MapItem{}
		k := reflect.ValueOf(&item.Key).Elem()
		if d.unmarshal(n.Children[i], k) {
			v := reflect.ValueOf(&item.Value).Elem()
			if d.unmarshal(n.Children[i+1], v) {
				slice = append(slice, item)
			}
		}
	}
	out.Set(reflect.ValueOf(slice))
	d.mapType = mapType
	return true
}

func (d *decoder) mappingStruct(n *Node, out reflect.Value) (good bool) {
	sinfo, err := getStructInfo(out.Type())
	if err != nil {
		panic(err)
	}
	name := settableValueOf("")
	l := len(n.Children)

	var inlineMap reflect.Value
	var elemType reflect.Type
	if sinfo.InlineMap != -1 {
		inlineMap = out.Field(sinfo.InlineMap)
		inlineMap.Set(reflect.New(inlineMap.Type()).Elem())
		elemType = inlineMap.Type().Elem()
	}

	var doneFields []bool
	if d.strict {
		doneFields = make([]bool, len(sinfo.FieldsList))
	}
	for i := 0; i < l; i += 2 {
		ni := n.Children[i]
		if isMerge(ni) {
			d.merge(n.Children[i+1], out)
			continue
		}
		if !d.unmarshal(ni, name) {
			continue
		}
		if info, ok := sinfo.FieldsMap[name.String()]; ok {
			if d.strict {
				if doneFields[info.Id] {
					d.terrors = append(d.terrors, fmt.Sprintf("line %d: field %s already set in type %s", ni.Line, name.String(), out.Type()))
					continue
				}
				doneFields[info.Id] = true
			}
			var field reflect.Value
			if info.Inline == nil {
				field = out.Field(info.Num)
			} else {
				field = out.FieldByIndex(info.Inline)
			}
			d.unmarshal(n.Children[i+1], field)
		} else if sinfo.InlineMap != -1 {
			if inlineMap.IsNil() {
				inlineMap.Set(reflect.MakeMap(inlineMap.Type()))
			}
			value := reflect.New(elemType).Elem()
			d.unmarshal(n.Children[i+1], value)
			d.setMapIndex(n.Children[i+1], inlineMap, name, value)
		} else if d.strict {
			d.terrors = append(d.terrors, fmt.Sprintf("line %d: field %s not found in type %s", ni.Line, name.String(), out.Type()))
		}
	}
	return true
}

func failWantMap() {
	failf("map merge requires map or sequence of maps as the value")
}

func (d *decoder) merge(n *Node, out reflect.Value) {
	switch n.Kind {
	case MappingNode:
		d.unmarshal(n, out)
	case AliasNode:
		an, ok := d.doc.Anchors[n.Value]
		if ok && an.Kind != MappingNode {
			failWantMap()
		}
		d.unmarshal(n, out)
	case SequenceNode:
		// Step backwards as earlier nodes take precedence.
		for i := len(n.Children) - 1; i >= 0; i-- {
			ni := n.Children[i]
			if ni.Kind == AliasNode {
				an, ok := d.doc.Anchors[ni.Value]
				if ok && an.Kind != MappingNode {
					failWantMap()
				}
			} else if ni.Kind != MappingNode {
				failWantMap()
			}
			d.unmarshal(ni, out)
		}
	default:
		failWantMap()
	}
}

func isMerge(n *Node) bool {
	return n.Kind == ScalarNode && n.Value == "<<" && (n.implicit() || n.Tag == yaml_MERGE_TAG)
}
