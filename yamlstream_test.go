package yaml_test

import (
	"bytes"
	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
	"io"
	"reflect"
)

var decodeStreamTest = []struct {
	data  string
	value interface{}
}{
	{
		`---
time: 20:03:20
player: Sammy Sosa
action: strike (miss)
...
---
time: 20:03:47
player: Sammy Sosa
action: grand slam
...
`,
		[]map[string]string{
			{"time": "20:03:20", "player": "Sammy Sosa", "action": "strike (miss)"},
			{"time": "20:03:47", "player": "Sammy Sosa", "action": "grand slam"},
		},
	},
	{
		`---
time: 20:03:20
player: Sammy Sosa
action: strike (miss)
---
time: 20:03:47
player: Sammy Sosa
action: grand slam
`,
		[]map[string]string{
			{"time": "20:03:20", "player": "Sammy Sosa", "action": "strike (miss)"},
			{"time": "20:03:47", "player": "Sammy Sosa", "action": "grand slam"},
		},
	},
}

func testDecode(c *C, v reflect.Value, decoder *yaml.Decoder) {
	t := v.Type()
	var value interface{}
	switch t.Kind() {
	case reflect.Map:
		value = reflect.MakeMap(t).Interface()
	case reflect.String:
		value = reflect.New(t).Interface()
	case reflect.Ptr:
		value = reflect.New(t.Elem()).Interface()
	default:
		c.Fatalf("missing case for %s", t)
	}
	err := decoder.Decode(value)
	if _, ok := err.(*yaml.TypeError); !ok {
		if err != io.EOF {
			c.Assert(err, IsNil)
		}
	}
	if t.Kind() == reflect.String {
		c.Assert(*value.(*string), Equals, v.Interface())
	} else {
		c.Assert(value, DeepEquals, v.Interface())
	}
}

func (s *S) TestDecode(c *C) {
	testCases := append(unmarshalTests, decodeStreamTest...)
	for _, item := range testCases {
		// prepare YAML Decoder
		buf := bytes.NewBufferString(item.data)
		decoder := yaml.NewDecoder(buf)

		// test agains reference
		ref := reflect.ValueOf(item.value)
		switch ref.Kind() {
		default:
			testDecode(c, ref, decoder)
		case reflect.Slice:
			// iterate through reference values
			for idx := 0; idx < ref.Len(); idx++ {
				v := ref.Index(idx)
				testDecode(c, v, decoder)
			}
		}
	}
}

func (s *S) TestDecodeError(c *C) {
	for _, item := range unmarshalErrorTests {
		// prepare YAML Decoder
		buf := bytes.NewBufferString(item.data)
		decoder := yaml.NewDecoder(buf)
		var value interface{}
		err := decoder.Decode(&value)
		c.Assert(err, ErrorMatches, item.error, Commentf("Partial unmarshal: %#v", value))
	}
}
