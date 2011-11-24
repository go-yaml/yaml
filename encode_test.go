package goyaml_test

import (
	. "launchpad.net/gocheck"
	"launchpad.net/goyaml"
	"math"
	//"reflect"
)

var marshalIntTest = 123

var marshalTests = []struct {
	data  string
	value interface{}
}{
	{"{}\n", &struct{}{}},
	{"v: hi\n", map[string]string{"v": "hi"}},
	{"v: hi\n", map[string]interface{}{"v": "hi"}},
	{"v: \"true\"\n", map[string]string{"v": "true"}},
	{"v: \"false\"\n", map[string]string{"v": "false"}},
	{"v: true\n", map[string]interface{}{"v": true}},
	{"v: false\n", map[string]interface{}{"v": false}},
	{"v: 10\n", map[string]interface{}{"v": 10}},
	{"v: -10\n", map[string]interface{}{"v": -10}},
	{"v: 42\n", map[string]uint{"v": 42}},
	{"v: 4294967296\n", map[string]interface{}{"v": int64(4294967296)}},
	{"v: 4294967296\n", map[string]int64{"v": int64(4294967296)}},
	{"v: 4294967296\n", map[string]uint64{"v": 4294967296}},
	{"v: \"10\"\n", map[string]interface{}{"v": "10"}},
	{"v: 0.1\n", map[string]interface{}{"v": 0.1}},
	{"v: 0.1\n", map[string]interface{}{"v": float64(0.1)}},
	{"v: -0.1\n", map[string]interface{}{"v": -0.1}},
	{"v: .inf\n", map[string]interface{}{"v": math.Inf(+1)}},
	{"v: -.inf\n", map[string]interface{}{"v": math.Inf(-1)}},
	{"v: .nan\n", map[string]interface{}{"v": math.NaN()}},
	{"v: null\n", map[string]interface{}{"v": nil}},
	{"v: \"\"\n", map[string]interface{}{"v": ""}},
	{"v:\n- A\n- B\n", map[string][]string{"v": []string{"A", "B"}}},
	{"v:\n- A\n- 1\n", map[string][]interface{}{"v": []interface{}{"A", 1}}},
	{"a:\n  b: c\n",
		map[string]interface{}{"a": map[interface{}]interface{}{"b": "c"}}},

	// Simple values.
	{"123\n", &marshalIntTest},

	// Structures
	{"hello: world\n", &struct{ Hello string }{"world"}},
	{"a:\n  b: c\n", &struct {
		A struct {
			B string
		}
	}{struct{ B string }{"c"}}},
	{"a:\n  b: c\n", &struct {
		A *struct {
			B string
		}
	}{&struct{ B string }{"c"}}},
	{"a: null\n", &struct {
		A *struct {
			B string
		}
	}{}},
	{"a: 1\n", &struct{ A int }{1}},
	{"a:\n- 1\n- 2\n", &struct{ A []int }{[]int{1, 2}}},
	{"a: 1\n", &struct {
		B int "a"
	}{1}},
	{"a: true\n", &struct{ A bool }{true}},

	// Conditional flag
	{"a: 1\n", &struct {
		A int "a,omitempty"
		B int "b,omitempty"
	}{1, 0}},
	{"{}\n", &struct {
		A int "a,omitempty"
		B int "b,omitempty"
	}{0, 0}},
	{"{}\n", &struct {
		A *struct{ X int } "a,omitempty"
		B int              "b,omitempty"
	}{nil, 0}},

	// Flow flag
	{"a: [1, 2]\n", &struct {
		A []int "a,flow"
	}{[]int{1, 2}}},
	{"a: {b: c}\n",
		&struct {
			A map[string]string "a,flow"
		}{map[string]string{"b": "c"}}},
	{"a: {b: c}\n",
		&struct {
			A struct {
				B string
			} "a,flow"
		}{struct{ B string }{"c"}}},
}

func (s *S) TestMarshal(c *C) {
	for _, item := range marshalTests {
		data, err := goyaml.Marshal(item.value)
		c.Assert(err, IsNil)
		c.Assert(string(data), Equals, item.data)
	}
}

//var unmarshalErrorTests = []struct{data, error string}{
//    {"v: !!float 'error'", "Can't decode !!str 'error' as a !!float"},
//}
//
//func (s *S) TestUnmarshalErrors(c *C) {
//    for _, item := range unmarshalErrorTests {
//        var value interface{}
//        err := goyaml.Unmarshal([]byte(item.data), &value)
//        c.Assert(err, Matches, item.error)
//    }
//}

var marshalTaggedIfaceTest interface{} = &struct{ A string }{"B"}

var getterTests = []struct {
	data, tag string
	value     interface{}
}{
	{"_:\n  hi: there\n", "", map[interface{}]interface{}{"hi": "there"}},
	{"_:\n- 1\n- A\n", "", []interface{}{1, "A"}},
	{"_: 10\n", "", 10},
	{"_: null\n", "", nil},
	{"_: !foo BAR!\n", "!foo", "BAR!"},
	{"_: !foo 1\n", "!foo", "1"},
	{"_: !foo '\"1\"'\n", "!foo", "\"1\""},
	{"_: !foo 1.1\n", "!foo", 1.1},
	{"_: !foo 1\n", "!foo", 1},
	{"_: !foo 1\n", "!foo", uint(1)},
	{"_: !foo true\n", "!foo", true},
	{"_: !foo\n- A\n- B\n", "!foo", []string{"A", "B"}},
	{"_: !foo\n  A: B\n", "!foo", map[string]string{"A": "B"}},
	{"_: !foo\n  a: B\n", "!foo", &marshalTaggedIfaceTest},
}

type typeWithGetter struct {
	tag   string
	value interface{}
}

func (o typeWithGetter) GetYAML() (tag string, value interface{}) {
	return o.tag, o.value
}

type typeWithGetterField struct {
	Field typeWithGetter "_"
}

func (s *S) TestMashalWithGetter(c *C) {
	for _, item := range getterTests {
		obj := &typeWithGetterField{}
		obj.Field.tag = item.tag
		obj.Field.value = item.value
		data, err := goyaml.Marshal(obj)
		c.Assert(err, IsNil)
		c.Assert(string(data), Equals, string(item.data))
	}
}

func (s *S) TestUnmarshalWholeDocumentWithGetter(c *C) {
	obj := &typeWithGetter{}
	obj.tag = ""
	obj.value = map[string]string{"hello": "world!"}
	data, err := goyaml.Marshal(obj)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "hello: world!\n")
}
