package yaml

import (
	"io"
	"reflect"
)

type Decoder struct {
	d *decoder
	p *parser
}

func (dec *Decoder) Decode(out interface{}) (err error) {
	defer handleErr(&err)
	if dec.p.event.typ == yaml_STREAM_END_EVENT {
		return io.EOF
	}
	node := dec.p.parse()
	if node != nil {
		v := reflect.ValueOf(out)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
		dec.d.unmarshal(node, v)
	}
	if len(dec.d.terrors) > 0 {
		return &TypeError{dec.d.terrors}
	}
	return nil
}

func (dec *Decoder) Close() {
	dec.p.destroy()
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		d: newDecoder(),
		p: newFileParser(r),
	}
}

type Encoder struct {
	e *encoder
}

func (enc *Encoder) Encode(in interface{}) (err error) {
	defer handleErr(&err)
	enc.e.begin()
	enc.e.marshal("", reflect.ValueOf(in))
	enc.e.end()
	return
}

func (enc *Encoder) Close() {
	enc.e.finish()
	enc.e.destroy()
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		e: newFileEncoder(w),
	}
}
