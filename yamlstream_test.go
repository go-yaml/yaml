package yaml

import (
	"bytes"
	"io"
	"math"
	"net"
	"reflect"
	"strings"
	"time"

	. "gopkg.in/check.v1"
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

func testDecode(c *C, v reflect.Value, decoder *Decoder) {
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
	if _, ok := err.(*TypeError); !ok {
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
		decoder := NewDecoder(buf)

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
	// NotStrict
	for _, item := range unmarshalErrorTests {
		// prepare YAML Decoder
		buf := bytes.NewBufferString(item.data)
		decoder := NewDecoder(buf)
		var value interface{}
		err := decoder.Decode(&value)
		c.Assert(err, ErrorMatches, item.error, Commentf("Partial unmarshal: %#v", value))
	}
	// Strict
	for _, item := range unmarshalErrorTests {
		// prepare YAML Decoder
		buf := bytes.NewBufferString(item.data)
		decoder := NewDecoder(buf)
		var value interface{}
		err := decoder.DecodeStrict(&value)
		c.Assert(err, ErrorMatches, item.error, Commentf("Partial unmarshal: %#v", value))
	}
}

var encoderTests = []struct {
	value interface{}
	data  string
}{
	{
		nil,
		"--- null\n...\n",
	}, {
		&struct{}{},
		"--- {}\n...\n",
	}, {
		map[string]string{"v": "hi"},
		"---\nv: hi\n...\n",
	}, {
		map[string]interface{}{"v": "hi"},
		"---\nv: hi\n...\n",
	}, {
		map[string]string{"v": "true"},
		"---\nv: \"true\"\n...\n",
	}, {
		map[string]string{"v": "false"},
		"---\nv: \"false\"\n...\n",
	}, {
		map[string]interface{}{"v": true},
		"---\nv: true\n...\n",
	}, {
		map[string]interface{}{"v": false},
		"---\nv: false\n...\n",
	}, {
		map[string]interface{}{"v": 10},
		"---\nv: 10\n...\n",
	}, {
		map[string]interface{}{"v": -10},
		"---\nv: -10\n...\n",
	}, {
		map[string]uint{"v": 42},
		"---\nv: 42\n...\n",
	}, {
		map[string]interface{}{"v": int64(4294967296)},
		"---\nv: 4294967296\n...\n",
	}, {
		map[string]int64{"v": int64(4294967296)},
		"---\nv: 4294967296\n...\n",
	}, {
		map[string]uint64{"v": 4294967296},
		"---\nv: 4294967296\n...\n",
	}, {
		map[string]interface{}{"v": "10"},
		"---\nv: \"10\"\n...\n",
	}, {
		map[string]interface{}{"v": 0.1},
		"---\nv: 0.1\n...\n",
	}, {
		map[string]interface{}{"v": float64(0.1)},
		"---\nv: 0.1\n...\n",
	}, {
		map[string]interface{}{"v": -0.1},
		"---\nv: -0.1\n...\n",
	}, {
		map[string]interface{}{"v": math.Inf(+1)},
		"---\nv: .inf\n...\n",
	}, {
		map[string]interface{}{"v": math.Inf(-1)},
		"---\nv: -.inf\n...\n",
	}, {
		map[string]interface{}{"v": math.NaN()},
		"---\nv: .nan\n...\n",
	}, {
		map[string]interface{}{"v": nil},
		"---\nv: null\n...\n",
	}, {
		map[string]interface{}{"v": ""},
		"---\nv: \"\"\n...\n",
	}, {
		map[string][]string{"v": {"A", "B"}},
		"---\nv:\n- A\n- B\n...\n",
	}, {
		map[string][]string{"v": {"A", "B\nC"}},
		"---\nv:\n- A\n- |-\n  B\n  C\n...\n",
	}, {
		map[string][]interface{}{"v": {"A", 1, map[string][]int{"B": {2, 3}}}},
		"---\nv:\n- A\n- 1\n- B:\n  - 2\n  - 3\n...\n",
	}, {
		map[string]interface{}{"a": map[interface{}]interface{}{"b": "c"}},
		"---\na:\n  b: c\n...\n",
	}, {
		map[string]interface{}{"a": "-"},
		"---\na: '-'\n...\n",
	},

	// Simple values.
	{
		&marshalIntTest,
		"--- 123\n...\n",
	},

	// Structures
	{
		&struct{ Hello string }{"world"},
		"---\nhello: world\n...\n",
	}, {
		&struct {
			A struct {
				B string
			}
		}{struct{ B string }{"c"}},
		"---\na:\n  b: c\n...\n",
	}, {
		&struct {
			A *struct {
				B string
			}
		}{&struct{ B string }{"c"}},
		"---\na:\n  b: c\n...\n",
	}, {
		&struct {
			A *struct {
				B string
			}
		}{},
		"---\na: null\n...\n",
	}, {
		&struct{ A int }{1},
		"---\na: 1\n...\n",
	}, {
		&struct{ A []int }{[]int{1, 2}},
		"---\na:\n- 1\n- 2\n...\n",
	}, {
		&struct {
			B int `yaml:"a"`
		}{1},
		"---\na: 1\n...\n",
	}, {
		&struct{ A bool }{true},
		"---\na: true\n...\n",
	},

	// Conditional flag
	{
		&struct {
			A int `yaml:"a,omitempty"`
			B int `yaml:"b,omitempty"`
		}{1, 0},
		"---\na: 1\n...\n",
	}, {
		&struct {
			A int `yaml:"a,omitempty"`
			B int `yaml:"b,omitempty"`
		}{0, 0},
		"--- {}\n...\n",
	}, {
		&struct {
			A *struct{ X, y int } `yaml:"a,omitempty,flow"`
		}{&struct{ X, y int }{1, 2}},
		"---\na: {x: 1}\n...\n",
	}, {
		&struct {
			A *struct{ X, y int } `yaml:"a,omitempty,flow"`
		}{nil},
		"--- {}\n...\n",
	}, {
		&struct {
			A *struct{ X, y int } `yaml:"a,omitempty,flow"`
		}{&struct{ X, y int }{}},
		"---\na: {x: 0}\n...\n",
	}, {
		&struct {
			A struct{ X, y int } `yaml:"a,omitempty,flow"`
		}{struct{ X, y int }{1, 2}},
		"---\na: {x: 1}\n...\n",
	}, {
		&struct {
			A struct{ X, y int } `yaml:"a,omitempty,flow"`
		}{struct{ X, y int }{0, 1}},
		"--- {}\n...\n",
	}, {
		&struct {
			A float64 `yaml:"a,omitempty"`
			B float64 `yaml:"b,omitempty"`
		}{1, 0},
		"---\na: 1\n...\n",
	},

	// Flow flag
	{
		&struct {
			A []int `yaml:"a,flow"`
		}{[]int{1, 2}},
		"---\na: [1, 2]\n...\n",
	}, {
		&struct {
			A map[string]string `yaml:"a,flow"`
		}{map[string]string{"b": "c", "d": "e"}},
		"---\na: {b: c, d: e}\n...\n",
	}, {
		&struct {
			A struct {
				B, D string
			} `yaml:"a,flow"`
		}{struct{ B, D string }{"c", "e"}},
		"---\na: {b: c, d: e}\n...\n",
	},

	// Unexported field
	{
		&struct {
			u int
			A int
		}{0, 1},
		"---\na: 1\n...\n",
	},

	// Ignored field
	{
		&struct {
			A int
			B int `yaml:"-"`
		}{1, 2},
		"---\na: 1\n...\n",
	},

	// Struct inlining
	{
		&struct {
			A int
			C inlineB `yaml:",inline"`
		}{1, inlineB{2, inlineC{3}}},
		"---\na: 1\nb: 2\nc: 3\n...\n",
	},

	// Map inlining
	{
		&struct {
			A int
			C map[string]int `yaml:",inline"`
		}{1, map[string]int{"b": 2, "c": 3}},
		"---\na: 1\nb: 2\nc: 3\n...\n",
	},

	// Duration
	{
		map[string]time.Duration{"a": 3 * time.Second},
		"---\na: 3s\n...\n",
	},

	// Issue #24: bug in map merging logic.
	{
		map[string]string{"a": "<foo>"},
		"---\na: <foo>\n...\n",
	},

	// Issue #34: marshal unsupported base 60 floats quoted for compatibility
	// with old YAML 1.1 parsers.
	{
		map[string]string{"a": "1:1"},
		"---\na: \"1:1\"\n...\n",
	},

	// Binary data.
	{
		map[string]string{"a": "\x00"},
		"---\na: \"\\0\"\n...\n",
	}, {
		map[string]string{"a": "\x80\x81\x82"},
		"---\na: !!binary gIGC\n...\n",
	}, {
		map[string]string{"a": strings.Repeat("\x90", 54)},
		"---\na: !!binary |\n  " + strings.Repeat("kJCQ", 17) + "kJ\n  CQ\n...\n",
	},

	// Ordered maps.
	{
		&MapSlice{{"b", 2}, {"a", 1}, {"d", 4}, {"c", 3}, {"sub", MapSlice{{"e", 5}}}},
		"---\nb: 2\na: 1\nd: 4\nc: 3\nsub:\n  e: 5\n...\n",
	},

	// Encode unicode as utf-8 rather than in escaped form.
	{
		map[string]string{"a": "你好"},
		"---\na: 你好\n...\n",
	},

	// Support encoding.TextMarshaler.
	{
		map[string]net.IP{"a": net.IPv4(1, 2, 3, 4)},
		"---\na: 1.2.3.4\n...\n",
	},
	{
		map[string]time.Time{"a": time.Unix(1424801979, 0).UTC()},
		"---\na: 2015-02-24T18:19:39Z\n...\n",
	},

	// Ensure strings containing ": " are quoted (reported as PR #43, but not reproducible).
	{
		map[string]string{"a": "b: c"},
		"---\na: 'b: c'\n...\n",
	},

	// Containing hash mark ('#') in string should be quoted
	{
		map[string]string{"a": "Hello #comment"},
		"---\na: 'Hello #comment'\n...\n",
	},
	{
		map[string]string{"a": "你好 #comment"},
		"---\na: '你好 #comment'\n...\n",
	},
}

func (s *S) TestEncode(c *C) {
	for _, item := range encoderTests {
		var buf = bytes.NewBuffer(nil)
		err := NewEncoder(buf).Encode(item.value)
		c.Assert(err, IsNil)
		c.Assert(buf.String(), Equals, item.data)
	}
}
