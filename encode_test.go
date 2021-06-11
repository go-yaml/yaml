//
// Copyright (c) 2011-2019 Canonical Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml_test

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v3"
)

var marshalIntTest = 123

var marshalTests = []struct {
	value interface{}
	data  string
}{
	{
		nil, `
null
`,
	}, {
		(*marshalerType)(nil), `
null
`,
	}, {
		&struct{}{}, `
{}
`,
	}, {
		map[string]string{"v": "hi"}, `
v: hi
`,
	}, {
		map[string]interface{}{"v": "hi"}, `
v: hi
`,
	}, {
		map[string]string{"v": "true"}, `
v: "true"
`,
	}, {
		map[string]string{"v": "false"}, `
v: "false"
`,
	}, {
		map[string]interface{}{"v": true}, `
v: true
`,
	}, {
		map[string]interface{}{"v": false}, `
v: false
`,
	}, {
		map[string]interface{}{"v": 10}, `
v: 10
`,
	}, {
		map[string]interface{}{"v": -10}, `
v: -10
`,
	}, {
		map[string]uint{"v": 42}, `
v: 42
`,
	}, {
		map[string]interface{}{"v": int64(4294967296)}, `
v: 4294967296
`,
	}, {
		map[string]int64{"v": int64(4294967296)}, `
v: 4294967296
`,
	}, {
		map[string]uint64{"v": 4294967296}, `
v: 4294967296
`,
	}, {
		map[string]interface{}{"v": "10"}, `
v: "10"
`,
	}, {
		map[string]interface{}{"v": 0.1}, `
v: 0.1
`,
	}, {
		map[string]interface{}{"v": float64(0.1)}, `
v: 0.1
`,
	}, {
		map[string]interface{}{"v": float32(0.99)}, `
v: 0.99
`,
	}, {
		map[string]interface{}{"v": -0.1}, `
v: -0.1
`,
	}, {
		map[string]interface{}{"v": math.Inf(+1)}, `
v: .inf
`,
	}, {
		map[string]interface{}{"v": math.Inf(-1)}, `
v: -.inf
`,
	}, {
		map[string]interface{}{"v": math.NaN()}, `
v: .nan
`,
	}, {
		map[string]interface{}{"v": nil}, `
v: null
`,
	}, {
		map[string]interface{}{"v": ""}, `
v: ""
`,
	}, {
		map[string][]string{"v": []string{"A", "B"}}, `
v:
  - A
  - B
`,
	}, {
		map[string][]string{"v": []string{"A", "B\nC"}}, `
v:
  - A
  - |-
    B
    C
`,
	}, {
		map[string][]interface{}{"v": []interface{}{"A", 1, map[string][]int{"B": []int{2, 3}}}}, `
v:
  - A
  - 1
  - B:
      - 2
      - 3
`,
	}, {
		map[string]interface{}{"a": map[interface{}]interface{}{"b": "c"}}, `
a:
    b: c
`,
	}, {
		map[string]interface{}{"a": "-"}, `
a: '-'
`,
	},

	// Simple values.
	{
		&marshalIntTest, `
123
`,
	},

	// Structures
	{
		&struct{ Hello string }{"world"}, `
hello: world
`,
	}, {
		&struct {
			A struct {
				B string
			}
		}{struct{ B string }{"c"}}, `
a:
    b: c
`,
	}, {
		&struct {
			A *struct {
				B string
			}
		}{&struct{ B string }{"c"}}, `
a:
    b: c
`,
	}, {
		&struct {
			A *struct {
				B string
			}
		}{}, `
a: null
`,
	}, {
		&struct{ A int }{1}, `
a: 1
`,
	}, {
		&struct{ A []int }{[]int{1, 2}}, `
a:
  - 1
  - 2
`,
	}, {
		&struct{ A [2]int }{[2]int{1, 2}}, `
a:
  - 1
  - 2
`,
	}, {
		&struct {
			B int "a"
		}{1}, `
a: 1
`,
	}, {
		&struct{ A bool }{true}, `
a: true
`,
	}, {
		&struct{ A string }{"true"}, `
a: "true"
`,
	}, {
		&struct{ A string }{"off"}, `
a: "off"
`,
	},

	// Conditional flag
	{
		&struct {
			A int "a,omitempty"
			B int "b,omitempty"
		}{1, 0}, `
a: 1
`,
	}, {
		&struct {
			A int "a,omitempty"
			B int "b,omitempty"
		}{0, 0}, `
{}
`,
	}, {
		&struct {
			A *struct{ X, y int } "a,omitempty,flow"
		}{&struct{ X, y int }{1, 2}}, `
a: {x: 1}
`,
	}, {
		&struct {
			A *struct{ X, y int } "a,omitempty,flow"
		}{nil}, `
{}
`,
	}, {
		&struct {
			A *struct{ X, y int } "a,omitempty,flow"
		}{&struct{ X, y int }{}}, `
a: {x: 0}
`,
	}, {
		&struct {
			A struct{ X, y int } "a,omitempty,flow"
		}{struct{ X, y int }{1, 2}}, `
a: {x: 1}
`,
	}, {
		&struct {
			A struct{ X, y int } "a,omitempty,flow"
		}{struct{ X, y int }{0, 1}}, `
{}
`,
	}, {
		&struct {
			A float64 "a,omitempty"
			B float64 "b,omitempty"
		}{1, 0}, `
a: 1
`,
	},
	{
		&struct {
			T1 time.Time  "t1,omitempty"
			T2 time.Time  "t2,omitempty"
			T3 *time.Time "t3,omitempty"
			T4 *time.Time "t4,omitempty"
		}{
			T2: time.Date(2018, 1, 9, 10, 40, 47, 0, time.UTC),
			T4: newTime(time.Date(2098, 1, 9, 10, 40, 47, 0, time.UTC)),
		}, `
t2: 2018-01-09T10:40:47Z
t4: 2098-01-09T10:40:47Z
`,
	},
	// Nil interface that implements Marshaler.
	{
		map[string]yaml.Marshaler{
			"a": nil,
		}, `
a: null
`,
	},

	// Flow flag
	{
		&struct {
			A []int "a,flow"
		}{[]int{1, 2}}, `
a: [1, 2]
`,
	}, {
		&struct {
			A map[string]string "a,flow"
		}{map[string]string{"b": "c", "d": "e"}}, `
a: {b: c, d: e}
`,
	}, {
		&struct {
			A struct {
				B, D string
			} "a,flow"
		}{struct{ B, D string }{"c", "e"}}, `
a: {b: c, d: e}
`,
	}, {
		&struct {
			A string "a,flow"
		}{"b\nc"}, `
a: "b\nc"
`,
	},

	// Unexported field
	{
		&struct {
			u int
			A int
		}{0, 1}, `
a: 1
`,
	},

	// Ignored field
	{
		&struct {
			A int
			B int "-"
		}{1, 2}, `
a: 1
`,
	},

	// Struct inlining
	{
		&struct {
			A int
			C inlineB `yaml:",inline"`
		}{1, inlineB{2, inlineC{3}}}, `
a: 1
b: 2
c: 3
`,
	},
	// Struct inlining as a pointer
	{
		&struct {
			A int
			C *inlineB `yaml:",inline"`
		}{1, &inlineB{2, inlineC{3}}}, `
a: 1
b: 2
c: 3
`,
	}, {
		&struct {
			A int
			C *inlineB `yaml:",inline"`
		}{1, nil}, `
a: 1
`,
	}, {
		&struct {
			A int
			D *inlineD `yaml:",inline"`
		}{1, &inlineD{&inlineC{3}, 4}}, `
a: 1
c: 3
d: 4
`,
	},

	// Map inlining
	{
		&struct {
			A int
			C map[string]int `yaml:",inline"`
		}{1, map[string]int{"b": 2, "c": 3}}, `
a: 1
b: 2
c: 3
`,
	},

	// Duration
	{
		map[string]time.Duration{"a": 3 * time.Second}, `
a: 3s
`,
	},

	// Issue #24: bug in map merging logic.
	{
		map[string]string{"a": "<foo>"}, `
a: <foo>
`,
	},

	// Issue #34: marshal unsupported base 60 floats quoted for compatibility
	// with old YAML 1.1 parsers.
	{
		map[string]string{"a": "1:1"}, `
a: "1:1"
`,
	},

	// Binary data.
	{
		map[string]string{"a": "\x00"}, `
a: "\0"
`,
	}, {
		map[string]string{"a": "\x80\x81\x82"}, `
a: !!binary gIGC
`,
	}, {
		map[string]string{"a": strings.Repeat("\x90", 54)}, (`
a: !!binary |
    ` + strings.Repeat("kJCQ", 17) + `kJ
    CQ
`),
	},

	// Encode unicode as utf-8 rather than in escaped form.
	{
		map[string]string{"a": "你好"}, `
a: 你好
`,
	},

	// Support encoding.TextMarshaler.
	{
		map[string]net.IP{"a": net.IPv4(1, 2, 3, 4)}, `
a: 1.2.3.4
`,
	},
	// time.Time gets a timestamp tag.
	{
		map[string]time.Time{"a": time.Date(2015, 2, 24, 18, 19, 39, 0, time.UTC)}, `
a: 2015-02-24T18:19:39Z
`,
	},
	{
		map[string]*time.Time{"a": newTime(time.Date(2015, 2, 24, 18, 19, 39, 0, time.UTC))}, `
a: 2015-02-24T18:19:39Z
`,
	},
	{
		// This is confirmed to be properly decoded in Python (libyaml) without a timestamp tag.
		map[string]time.Time{"a": time.Date(2015, 2, 24, 18, 19, 39, 123456789, time.FixedZone("FOO", -3*60*60))}, `
a: 2015-02-24T18:19:39.123456789-03:00
`,
	},
	// Ensure timestamp-like strings are quoted.
	{
		map[string]string{"a": "2015-02-24T18:19:39Z"}, `
a: "2015-02-24T18:19:39Z"
`,
	},

	// Ensure strings containing ": " are quoted (reported as PR #43, but not reproducible).
	{
		map[string]string{"a": "b: c"}, `
a: 'b: c'
`,
	},

	// Containing hash mark ('#') in string should be quoted
	{
		map[string]string{"a": "Hello #comment"}, `
a: 'Hello #comment'
`,
	},
	{
		map[string]string{"a": "你好 #comment"}, `
a: '你好 #comment'
`,
	},

	// Ensure MarshalYAML also gets called on the result of MarshalYAML itself.
	{
		&marshalerType{marshalerType{true}}, `
true
`,
	}, {
		&marshalerType{&marshalerType{true}}, `
true
`,
	},

	// Check indentation of maps inside sequences inside maps.
	{
		map[string]interface{}{"a": map[string]interface{}{"b": []map[string]int{{"c": 1, "d": 2}}}}, `
a:
    b:
      - c: 1
        d: 2
`,
	},

	// Strings with tabs were disallowed as literals (issue #471).
	{
		map[string]string{"a": "\tB\n\tC\n"}, `
a: |
    	B
    	C
`,
	},

	// Ensure that strings do not wrap
	{
		map[string]string{"a": "abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890 abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890 "}, `
a: 'abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890 abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890 '
`,
	},

	// yaml.Node
	{
		&struct {
			Value yaml.Node
		}{
			yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "foo",
				Style: yaml.SingleQuotedStyle,
			},
		}, `
value: 'foo'
`,
	}, {
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "foo",
			Style: yaml.SingleQuotedStyle,
		}, `
'foo'
`,
	},

	// Enforced tagging with shorthand notation (issue #616).
	{
		&struct {
			Value yaml.Node
		}{
			yaml.Node{
				Kind:  yaml.ScalarNode,
				Style: yaml.TaggedStyle,
				Value: "foo",
				Tag:   "!!str",
			},
		}, `
value: !!str foo
`,
	}, {
		&struct {
			Value yaml.Node
		}{
			yaml.Node{
				Kind:  yaml.MappingNode,
				Style: yaml.TaggedStyle,
				Tag:   "!!map",
			},
		}, `
value: !!map {}
`,
	}, {
		&struct {
			Value yaml.Node
		}{
			yaml.Node{
				Kind:  yaml.SequenceNode,
				Style: yaml.TaggedStyle,
				Tag:   "!!seq",
			},
		}, `
value: !!seq []
`,
	},
}

func (s *S) TestMarshal(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range marshalTests {
		c.Logf("test %d: %q", i, item.data[1:])
		data, err := yaml.Marshal(item.value)
		c.Assert(err, IsNil)
		c.Assert(string(data), Equals, item.data[1:])
	}
}

func (s *S) TestEncoderSingleDocument(c *C) {
	for i, item := range marshalTests {
		c.Logf("test %d. %q", i, item.data[1:])
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		c.Assert(enc.Encode(item.value), IsNil)
		c.Assert(enc.Close(), IsNil)
		c.Assert(buf.String(), Equals, item.data[1:])
	}
}

func (s *S) TestEncoderMultipleDocuments(c *C) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	err := enc.Encode(map[string]string{"a": "b"})
	c.Assert(err, Equals, nil)
	err = enc.Encode(map[string]string{"c": "d"})
	c.Assert(err, Equals, nil)
	err = enc.Close()
	c.Assert(err, Equals, nil)
	c.Assert(buf.String(), Equals, "a: b\n---\nc: d\n")
}

func (s *S) TestEncoderWriteError(c *C) {
	enc := yaml.NewEncoder(errorWriter{})
	err := enc.Encode(map[string]string{"a": "b"})
	c.Assert(err, ErrorMatches, `yaml: write error: some write error`) // Data not flushed yet
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("some write error")
}

var marshalErrorTests = []struct {
	value interface{}
	error string
	panic string
}{{
	value: &struct {
		B       int
		inlineB ",inline"
	}{1, inlineB{2, inlineC{3}}},
	panic: `duplicated key 'b' in struct struct \{ B int; .*`,
}, {
	value: &struct {
		A int
		B map[string]int ",inline"
	}{1, map[string]int{"a": 2}},
	panic: `cannot have key "a" in inlined map: conflicts with struct field`,
}}

func (s *S) TestMarshalErrors(c *C) {
	for _, item := range marshalErrorTests {
		if item.panic != "" {
			c.Assert(func() { yaml.Marshal(item.value) }, PanicMatches, item.panic)
		} else {
			_, err := yaml.Marshal(item.value)
			c.Assert(err, ErrorMatches, item.error)
		}
	}
}

func (s *S) TestMarshalTypeCache(c *C) {
	var data []byte
	var err error
	func() {
		type T struct{ A int }
		data, err = yaml.Marshal(&T{})
		c.Assert(err, IsNil)
	}()
	func() {
		type T struct{ B int }
		data, err = yaml.Marshal(&T{})
		c.Assert(err, IsNil)
	}()
	c.Assert(string(data), Equals, "b: 0\n")
}

var marshalerTests = []struct {
	data  string
	value interface{}
}{
	{"_:\n    hi: there\n", map[interface{}]interface{}{"hi": "there"}},
	{"_:\n  - 1\n  - A\n", []interface{}{1, "A"}},
	{"_: 10\n", 10},
	{"_: null\n", nil},
	{"_: BAR!\n", "BAR!"},
}

type marshalerType struct {
	value interface{}
}

func (o marshalerType) MarshalText() ([]byte, error) {
	panic("MarshalText called on type with MarshalYAML")
}

func (o marshalerType) MarshalYAML() (interface{}, error) {
	return o.value, nil
}

type marshalerValue struct {
	Field marshalerType "_"
}

func (s *S) TestMarshaler(c *C) {
	for _, item := range marshalerTests {
		obj := &marshalerValue{}
		obj.Field.value = item.value
		data, err := yaml.Marshal(obj)
		c.Assert(err, IsNil)
		c.Assert(string(data), Equals, string(item.data))
	}
}

func (s *S) TestMarshalerWholeDocument(c *C) {
	obj := &marshalerType{}
	obj.value = map[string]string{"hello": "world!"}
	data, err := yaml.Marshal(obj)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "hello: world!\n")
}

type failingMarshaler struct{}

func (ft *failingMarshaler) MarshalYAML() (interface{}, error) {
	return nil, failingErr
}

func (s *S) TestMarshalerError(c *C) {
	_, err := yaml.Marshal(&failingMarshaler{})
	c.Assert(err, Equals, failingErr)
}

func (s *S) TestSetIndent(c *C) {
	makeObject := func() map[string]interface{} {
		return map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]string{
					"c": "d",
				},
				"e": []int{1, 2, 3},
				"f": []string{"g", "h"},
			},
		}
	}
	const (
		indentTwo = `
a:
  b:
    c: d
  e:
    - 1
    - 2
    - 3
  f:
    - g
    - h
`
		indentFour = `
a:
    b:
        c: d
    e:
      - 1
      - 2
      - 3
    f:
      - g
      - h
`
		indentEight = `
a:
        b:
                c: d
        e:
              - 1
              - 2
              - 3
        f:
              - g
              - h
`
	)

	testCases := []struct {
		indent   int
		expected string
	}{
		{
			indent:   -99, // Signal to not call SetIndent.  Default seen to be four.
			expected: indentFour,
		},
		{
			indent:   0, // zero silently treated as default (four).
			expected: indentFour,
		},
		{
			indent:   1, // one silently treated as two.
			expected: indentTwo,
		},
		{
			indent:   2, // two accepted as two.
			expected: indentTwo,
		},
		{
			indent:   8, // eight accepted as eight.
			expected: indentEight,
		},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		if tc.indent != -99 {
			enc.SetIndent(tc.indent)
		}
		c.Assert(enc.Encode(makeObject()), IsNil)
		c.Assert(enc.Close(), IsNil)
		c.Assert(
			buf.String(), Equals, tc.expected[1:],
			Commentf("Unexpected encoding; tc.indent = %d", tc.indent))
	}
}

func (s *S) TestSortedOutput(c *C) {
	order := []interface{}{
		false,
		true,
		1,
		uint(1),
		1.0,
		1.1,
		1.2,
		2,
		uint(2),
		2.0,
		2.1,
		"",
		".1",
		".2",
		".a",
		"1",
		"2",
		"a!10",
		"a/0001",
		"a/002",
		"a/3",
		"a/10",
		"a/11",
		"a/0012",
		"a/100",
		"a~10",
		"ab/1",
		"b/1",
		"b/01",
		"b/2",
		"b/02",
		"b/3",
		"b/03",
		"b1",
		"b01",
		"b3",
		"c2.10",
		"c10.2",
		"d1",
		"d7",
		"d7abc",
		"d12",
		"d12a",
		"e2b",
		"e4b",
		"e21a",
	}
	m := make(map[interface{}]int)
	for _, k := range order {
		m[k] = 1
	}
	data, err := yaml.Marshal(m)
	c.Assert(err, IsNil)
	out := "\n" + string(data)
	last := 0
	for i, k := range order {
		repr := fmt.Sprint(k)
		if s, ok := k.(string); ok {
			if _, err = strconv.ParseFloat(repr, 32); s == "" || err == nil {
				repr = `"` + repr + `"`
			}
		}
		index := strings.Index(out, "\n"+repr+":")
		if index == -1 {
			c.Fatalf("%#v is not in the output: %#v", k, out)
		}
		if index < last {
			c.Fatalf("%#v was generated before %#v: %q", k, order[i-1], out)
		}
		last = index
	}
}

func newTime(t time.Time) *time.Time {
	return &t
}
