package yaml

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"
)

const (
	documentNode = 1 << iota
	mappingNode
	sequenceNode
	scalarNode
	aliasNode
)

type node struct {
	kind         int
	line, column int
	tag          string
	value        string
	implicit     bool
	children     []*node
	anchors      map[string]*node
}

// ----------------------------------------------------------------------------
// Parser, produces a node tree out of a libyaml event stream.

type parser struct {
	parser yaml_parser_t
	event  yaml_event_t
	doc    *node
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

	p.skip()
	if p.event.typ != yaml_STREAM_START_EVENT {
		panic("expected stream start event, got " + strconv.Itoa(int(p.event.typ)))
	}
	p.skip()
	return &p
}

func (p *parser) destroy() {
	if p.event.typ != yaml_NO_EVENT {
		yaml_event_delete(&p.event)
	}
	yaml_parser_delete(&p.parser)
}

func (p *parser) skip() {
	if p.event.typ != yaml_NO_EVENT {
		if p.event.typ == yaml_STREAM_END_EVENT {
			failf("attempted to go past the end of stream; corrupted value?")
		}
		yaml_event_delete(&p.event)
	}
	if !yaml_parser_parse(&p.parser, &p.event) {
		p.fail()
	}
}

func (p *parser) fail() {
	var where string
	var line int
	if p.parser.problem_mark.line != 0 {
		line = p.parser.problem_mark.line
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

func (p *parser) anchor(n *node, anchor []byte) {
	if anchor != nil {
		p.doc.anchors[string(anchor)] = n
	}
}

func (p *parser) parse() *node {
	switch p.event.typ {
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
		panic("attempted to parse unknown event: " + strconv.Itoa(int(p.event.typ)))
	}
	panic("unreachable")
}

func (p *parser) node(kind int) *node {
	return &node{
		kind:   kind,
		line:   p.event.start_mark.line,
		column: p.event.start_mark.column,
	}
}

func (p *parser) document() *node {
	n := p.node(documentNode)
	n.anchors = make(map[string]*node)
	p.doc = n
	p.skip()
	n.children = append(n.children, p.parse())
	if p.event.typ != yaml_DOCUMENT_END_EVENT {
		panic("expected end of document event but got " + strconv.Itoa(int(p.event.typ)))
	}
	p.skip()
	return n
}

func (p *parser) alias() *node {
	n := p.node(aliasNode)
	n.value = string(p.event.anchor)
	p.skip()
	return n
}

func (p *parser) scalar() *node {
	n := p.node(scalarNode)
	n.value = string(p.event.value)
	n.tag = string(p.event.tag)
	n.implicit = p.event.implicit
	p.anchor(n, p.event.anchor)
	p.skip()
	return n
}

func (p *parser) sequence() *node {
	n := p.node(sequenceNode)
	p.anchor(n, p.event.anchor)
	p.skip()
	for p.event.typ != yaml_SEQUENCE_END_EVENT {
		n.children = append(n.children, p.parse())
	}
	p.skip()
	return n
}

func (p *parser) mapping() *node {
	n := p.node(mappingNode)
	p.anchor(n, p.event.anchor)
	p.skip()
	for p.event.typ != yaml_MAPPING_END_EVENT {
		n.children = append(n.children, p.parse(), p.parse())
	}
	p.skip()
	return n
}

// ----------------------------------------------------------------------------
// Decoder, unmarshals a node into a provided value.

type decoder struct {
	doc     *node
	aliases map[string]bool
	mapType reflect.Type
	terrors []string
}

var (
	mapItemType    = reflect.TypeOf(MapItem{})
	durationType   = reflect.TypeOf(time.Duration(0))
	defaultMapType = reflect.TypeOf(map[interface{}]interface{}{})
	ifaceType      = defaultMapType.Elem()
)

func newDecoder() *decoder {

	d := &decoder{mapType: defaultMapType}
	d.aliases = make(map[string]bool)
	return d
}

func (d *decoder) terror(n *node, tag string, out reflect.Value) {

	if n.tag != "" {
		tag = n.tag
	}

	value := n.value
	if tag != yaml_SEQ_TAG && tag != yaml_MAP_TAG {
		if len(value) > 10 {
			value = " `" + value[:7] + "...`"
		} else {
			value = " `" + value + "`"
		}
	}

	d.terrors = append(d.terrors, fmt.Sprintf("line %d: cannot unmarshal %s%s into %s", n.line+1, shortTag(tag), value, out.Type()))
}

// Use the Unmarshaler.UnmarshalYAML for types that are Unmarshalers, performing
// custom unmarshaling.
func (d *decoder) callUnmarshaler(n *node, u Unmarshaler) (good bool) {

	// Get the amount of existing decoder errors
	terrlen := len(d.terrors)

	// Call the Unmarshaler, giving it a chance to handle the unmarshalling
	// any way it wants, into v.
	err := u.UnmarshalYAML(func(v interface{}) (err error) {
		defer handleErr(&err)

		// Try and unmarshal the node into the type inputted
		d.unmarshal(n, reflect.ValueOf(v))

		// // If we have new decoder errors
		if len(d.terrors) > terrlen {

			// Get new decoder errors
			issues := d.terrors[terrlen:]

			// Remove them from the decoder errors list
			d.terrors = d.terrors[:terrlen]

			// And return just the specific decoder errors to the calling type
			return &TypeError{issues}
		}

		return nil
	})

	// Append all decoding errors received, maintaing order - allow the
	// calling object to return the errors it received
	if e, ok := err.(*TypeError); ok {
		d.terrors = append(d.terrors, e.Errors...)
		return false
	}

	// Or, if we had a different errors, fail immediately
	if err != nil {
		fail(err)
	}

	// Otherwise, we're good
	return true
}

// d.prepare initializes and dereferences pointers and calls UnmarshalYAML
// if a value is found to implement it.
// It returns the initialized and dereferenced out value, whether
// unmarshalling was already done by UnmarshalYAML, and if so whether
// its types unmarshalled appropriately.
//
// If n holds a null value, prepare returns before doing anything.
func (d *decoder) prepare(n *node, out reflect.Value) (newout reflect.Value, unmarshaled, good bool) {

	// TODO: Add support for Null tags, perhaps writing a to field etc
	if n.tag == yaml_NULL_TAG || n.kind == scalarNode && n.tag == "" && (n.value == "null" || n.value == "") {
		return out, false, false
	}

	// Loop until we have an unmarshaled value, or we are sure we have a
	// scalar, mapping or sequence
	again := true
	for again {
		again = false

		// Do we have a pointer to unmarshal into?
		if out.Kind() == reflect.Ptr {
			// If it is Nil, create an approriate new value and point the
			// pointer to it
			if out.IsNil() {
				out.Set(reflect.New(out.Type().Elem()))
			}

			// And dereference it, returning the actual value
			out = out.Elem()

			// Try again, looking at the new value - which might not be
			// addressable
			again = true
		}

		// Is the value an addressable value?
		if out.CanAddr() {

			// Is the value an Unmarshaler?
			if u, ok := out.Addr().Interface().(Unmarshaler); ok {

				// Call the Unmarshaler, allow custom unmarshalling
				good = d.callUnmarshaler(n, u)

				// Value is now nmarshalled (or, attempt failed)
				return out, true, good
			}
		}

		// If it isn't an addressable value, stop
	}

	// Couldn't find an addressable value that is an Unmarshaler, return
	// the last value we found
	return out, false, false
}

func (d *decoder) unmarshal(n *node, out reflect.Value) (good bool) {

	// Unmarshal a node based on its kind. Documents and liases first
	switch n.kind {
	case documentNode:
		return d.document(n, out)
	case aliasNode:
		return d.alias(n, out)
	}

	// Prepare for unmarshaling - checking if, perhaps, it is unnecessary
	// (already custom unmarshaled via Unmarshaler implementing values)
	out, unmarshaled, good := d.prepare(n, out)

	// There was an attempt to unmarshal, return the result
	if unmarshaled {
		return good
	}

	// Otherwise, unmarshal based on Go types
	switch n.kind {
	case scalarNode:
		good = d.scalar(n, out)
	case mappingNode:
		good = d.mapping(n, out)
	case sequenceNode:
		good = d.sequence(n, out)
	default:
		panic("internal error: unknown node kind: " + strconv.Itoa(n.kind))
	}

	return good
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
		failf("unknown anchor '%s' referenced", n.value)
	}
	if d.aliases[n.value] {
		failf("anchor '%s' value contains itself", n.value)
	}
	d.aliases[n.value] = true
	good = d.unmarshal(an, out)
	delete(d.aliases, n.value)
	return good
}

var zeroValue reflect.Value

// Reset a map, clearing all of its values
func resetMap(out reflect.Value) {

	// Go over all of the Map keys
	for _, k := range out.MapKeys() {

		// Delete the all of the values
		out.SetMapIndex(k, zeroValue)
	}
}

func (d *decoder) scalar(n *node, out reflect.Value) (good bool) {
	var tag string
	var resolved interface{}

	if n.tag == "" && !n.implicit {
		tag = yaml_STR_TAG
		resolved = n.value
	} else {
		tag, resolved = resolve(n.tag, n.value)
		if tag == yaml_BINARY_TAG {
			data, err := base64.StdEncoding.DecodeString(resolved.(string))
			if err != nil {
				failf("!!binary value contains invalid base64 data")
			}
			resolved = string(data)
		}
	}

	// Couldn't get proper tag/value for node
	if resolved == nil {

		// Is the output value for unmarshaling a map?
		if out.Kind() == reflect.Map && !out.CanAddr() {

			// Clear its values
			resetMap(out)
		} else {

			// Otherwise, clear whatever value is there
			out.Set(reflect.Zero(out.Type()))
		}

		return true
	}

	// Is the output value type is a TextUnmarshaler - let it handle
	// unmarshaling itself.
	if s, ok := resolved.(string); ok && out.CanAddr() {
		if u, ok := out.Addr().Interface().(encoding.TextUnmarshaler); ok {
			err := u.UnmarshalText([]byte(s))
			if err != nil {
				fail(err)
			}
			return true
		}
	}

	// Otherwise, look at the kind of output value
	switch out.Kind() {

	// Unmarshal into a string
	case reflect.String:

		if tag == yaml_BINARY_TAG {
			out.SetString(resolved.(string))
			good = true
		} else if resolved != nil {
			out.SetString(n.value)
			good = true
		}

	// Unmarshal into an interface{} - might be a string or a different basic
	// type
	case reflect.Interface:
		if resolved == nil {
			out.Set(reflect.Zero(out.Type()))
		} else {
			out.Set(reflect.ValueOf(resolved))
		}
		good = true

	// Basic Go types - integers
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch resolved := resolved.(type) {
		case int:
			if !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				good = true
			}
		case int64:
			if !out.OverflowInt(resolved) {
				out.SetInt(resolved)
				good = true
			}
		case uint64:
			if resolved <= math.MaxInt64 && !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				good = true
			}
		case float64:
			if resolved <= math.MaxInt64 && !out.OverflowInt(int64(resolved)) {
				out.SetInt(int64(resolved))
				good = true
			}
		case string:
			if out.Type() == durationType {
				d, err := time.ParseDuration(resolved)
				if err == nil {
					out.SetInt(int64(d))
					good = true
				}
			}
		}

	// Basic Go types - unsigned integers
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch resolved := resolved.(type) {
		case int:
			if resolved >= 0 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				good = true
			}
		case int64:
			if resolved >= 0 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				good = true
			}
		case uint64:
			if !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				good = true
			}
		case float64:
			if resolved <= math.MaxUint64 && !out.OverflowUint(uint64(resolved)) {
				out.SetUint(uint64(resolved))
				good = true
			}
		}

	// Basic Go types - booleans
	case reflect.Bool:
		switch resolved := resolved.(type) {
		case bool:
			out.SetBool(resolved)
			good = true
		}

	// Basic Go types - floating point numbers
	case reflect.Float32, reflect.Float64:
		switch resolved := resolved.(type) {
		case int:
			out.SetFloat(float64(resolved))
			good = true
		case int64:
			out.SetFloat(float64(resolved))
			good = true
		case uint64:
			out.SetFloat(float64(resolved))
			good = true
		case float64:
			out.SetFloat(resolved)
			good = true
		}

	// A pointer...
	case reflect.Ptr:
		if out.Type().Elem() == reflect.TypeOf(resolved) {
			// TODO: DOes this make sense? When is out a Ptr except when decoding a nil value?
			elem := reflect.New(out.Type().Elem())
			elem.Elem().Set(reflect.ValueOf(resolved))
			out.Set(elem)
			good = true
		}
	}

	if !good {
		d.terror(n, tag, out)
	}

	return good
}

func settableValueOf(i interface{}) reflect.Value {
	v := reflect.ValueOf(i)
	sv := reflect.New(v.Type()).Elem()
	sv.Set(v)
	return sv
}

func (d *decoder) sequence(n *node, out reflect.Value) (good bool) {
	var iface reflect.Value
	switch out.Kind() {
	case reflect.Slice:
		// okay
	case reflect.Interface:
		// No type hints. Will have to use a generic sequence.
		iface = out
		out = settableValueOf(make([]interface{}, 0))
	default:
		d.terror(n, yaml_SEQ_TAG, out)
		return false
	}
	et := out.Type().Elem()

	l := len(n.children)
	for i := 0; i < l; i++ {
		e := reflect.New(et).Elem()
		if ok := d.unmarshal(n.children[i], e); ok {
			out.Set(reflect.Append(out, e))
		}
	}
	if iface.IsValid() {
		iface.Set(out)
	}
	return true
}

// Unmarshal a YAML mapping into a map, a struct, a slice or even into an
// interface{} (actually creating a map or slice with the last encountered type)
func (d *decoder) mapping(n *node, out reflect.Value) (good bool) {

	switch out.Kind() {
	case reflect.Struct:
		return d.mappingStruct(n, out)
	case reflect.Slice:
		return d.mappingSlice(n, out)
	case reflect.Map:
		// Simple case, unmarshal the YAML mapping into a map
	case reflect.Interface:
		// Create a map/slice of the last type encountered and set it to out.
		// Starts as map[interface{}]interface{}, but might change as we go
		// deeper in the children nodes and find new maps/slices to unmarshal.
		if d.mapType.Kind() == reflect.Map {
			iface := out
			out = reflect.MakeMap(d.mapType)
			iface.Set(out)

			// Got a map - continue with unmarshaling the YAML mapping into
			// this new map
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

	// Go over the children nodes, unmarshaling each pair into a key/value
	l := len(n.children)
	for i := 0; i < l; i += 2 {

		// YAML << merge key
		if isMerge(n.children[i]) {
			d.merge(n.children[i+1], out)
			continue
		}

		// Create a new value of the map's key type
		k := reflect.New(kt).Elem()

		// Unmarshal into the key value
		if d.unmarshal(n.children[i], k) {

			// Successful unmarshalling.

			// Get the kind of the key - a go basic type
			kkind := k.Kind()
			if kkind == reflect.Interface {
				kkind = k.Elem().Kind()
			}

			// If the map key is a map or slice - die, we can't use it as a
			// key.
			if kkind == reflect.Map || kkind == reflect.Slice {
				failf("invalid map key: %#v", k.Interface())
			}

			// Otherwise, create a new value of the map's element type
			e := reflect.New(et).Elem()

			// Unmarshal into the element value
			if d.unmarshal(n.children[i+1], e) {

				// Set it into the map!
				out.SetMapIndex(k, e)
			} else {
				// Error is set inside the unmarshal call
			}
		}
	}
	d.mapType = mapType
	return true
}

// Map a YAML mapping into a slice
func (d *decoder) mappingSlice(n *node, out reflect.Value) (good bool) {
	outType := out.Type()
	if outType.Elem() != mapItemType {
		d.terror(n, yaml_MAP_TAG, out)
		return false
	}

	mapType := d.mapType
	d.mapType = outType

	var slice []MapItem
	var l = len(n.children)
	for i := 0; i < l; i += 2 {
		if isMerge(n.children[i]) {
			d.merge(n.children[i+1], out)
			continue
		}
		item := MapItem{}
		k := reflect.ValueOf(&item.Key).Elem()
		if d.unmarshal(n.children[i], k) {
			v := reflect.ValueOf(&item.Value).Elem()
			if d.unmarshal(n.children[i+1], v) {
				slice = append(slice, item)
			}
		}
	}
	out.Set(reflect.ValueOf(slice))
	d.mapType = mapType
	return true
}

// Map a YAML mapping into a struct
func (d *decoder) mappingStruct(n *node, out reflect.Value) (good bool) {

	sinfo, err := getStructInfo(out.Type())

	if err != nil {
		panic(err)
	}

	// Create a string-type Value, which will contain the YAML key
	name := settableValueOf("")

	// Go over the children nodes, unmarshaling each pair into a key/value
	l := len(n.children)
	for i := 0; i < l; i += 2 {
		ni := n.children[i]

		// YAML << merge key
		if isMerge(ni) {
			d.merge(n.children[i+1], out)
			continue
		}

		// Unmarshal the YAML key
		if !d.unmarshal(ni, name) {

			// Problems unmarshalling the key..
			// Error is set inside the unmarshal call
			continue
		}

		// Search to see if we have an exact match between the YAML key and
		// a field Key in the fields map. This is case sensitive, obviously
		if info, ok := sinfo.FieldsMap[name.String()]; ok {

			var field reflect.Value
			if info.Inline == nil {
				field = out.Field(info.Num)
			} else {
				field = out.FieldByIndex(info.Inline)
			}

			d.unmarshal(n.children[i+1], field)
		} else {
			// Otherwise, we try to see if the YAML key matches any regular
			// expression
			for _, info := range sinfo.RegexpFieldsList {
				if info.Regexp.MatchString(name.String()) {

					// Get the field. It must be a map or a slice
					var field reflect.Value = out.Field(info.Num)

					// Will we write to a map or a slice?
					if field.Kind() == reflect.Map {

						// If the map doesn't exist yet..
						if field.IsNil() {

							// Create a map, set it
							iface := field
							field = reflect.MakeMap(field.Type())
							iface.Set(field)
						}

						// Create a new value of the map element type
						e := reflect.New(field.Type().Elem()).Elem()

						// Unmarshal into the element value
						if d.unmarshal(n.children[i+1], e) {

							// Set it into the map!
							field.SetMapIndex(name, e)

						} else {
							// Error is set inside the unmarshal call
						}

					} else {

						// If the array doesn't exist yet..
						if field.IsNil() {

							// Create a slice, set it
							newSlice := reflect.MakeSlice(field.Type(), 0, 0)
							field.Set(newSlice)
						}

						// Create a new value of the map element type
						e := reflect.New(field.Type().Elem()).Elem()

						// Unmarshal into the element value
						if d.unmarshal(n.children[i+1], e) {

							// Append it to the slice
							newSlice := reflect.Append(field, e)
							field.Set(newSlice)

						} else {
							// Error is set inside the unmarshal call
						}
					}

					break

				}
			}
		}
	}
	return true
}

func failWantMap() {
	failf("map merge requires map or sequence of maps as the value")
}

func (d *decoder) merge(n *node, out reflect.Value) {
	switch n.kind {
	case mappingNode:
		d.unmarshal(n, out)
	case aliasNode:
		an, ok := d.doc.anchors[n.value]
		if ok && an.kind != mappingNode {
			failWantMap()
		}
		d.unmarshal(n, out)
	case sequenceNode:
		// Step backwards as earlier nodes take precedence.
		for i := len(n.children) - 1; i >= 0; i-- {
			ni := n.children[i]
			if ni.kind == aliasNode {
				an, ok := d.doc.anchors[ni.value]
				if ok && an.kind != mappingNode {
					failWantMap()
				}
			} else if ni.kind != mappingNode {
				failWantMap()
			}
			d.unmarshal(ni, out)
		}
	default:
		failWantMap()
	}
}

func isMerge(n *node) bool {
	return n.kind == scalarNode && n.value == "<<" && (n.implicit == true || n.tag == yaml_MERGE_TAG)
}
