package yaml

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

type testStruct struct {
	Meta StructMeta `yaml:"yaml_meta"`
	B    smChildStruct
	A    string
}

type smChildStruct struct {
	Meta StructMeta `yaml:"yaml_meta"`
	D    string
	C    string
}

var structMetaTests = []string{
	"a: a\nb:\n    c: c\n    d: d\n",
	"b: #foo\n    c: c #bar\n    d: d #baz\na: a #end\n",
	"a: ant #ant\n# a\nb: #beeline\n    c: cockroach #cockroach\n    #c\n    #d\n    d: dragonfly\n    #dragonfly\n",
}

var structMetaNilTests = []struct {
	expected string
	test     testStruct
}{
	{
		"b:\n    d: d\n    c: c\na: a\n",
		testStruct{
			A: "a",
			B: smChildStruct{
				C: "c",
				D: "d",
			},
		},
	},
	{
		"b:\n    d: d\n    c: c\na: a\n",
		testStruct{
			A: "a",
			B: smChildStruct{
				C: "c",
				D: "d",
			},
			Meta: StructMeta(&structMeta{nil, nil}),
		},
	},
	{
		"b:\n    d: d\n    c: c\na: a\n",
		testStruct{
			A: "a",
			B: smChildStruct{
				C: "c",
				D: "d",
			},
			Meta: StructMeta(&structMeta{
				[]fieldInfo{{
					Key: "a",
					Num: 2,
				}},
				[][]comments{},
			}),
		},
	},
}

func (s *S) TestStructMeta(c *C) {
	for _, expected := range structMetaTests {
		c.Logf("test %s.", expected)

		test := &testStruct{}
		err := Unmarshal([]byte(expected), test)
		c.Assert(err, Equals, nil)

		actual, err := Marshal(test)
		c.Assert(err, Equals, nil)
		c.Assert(string(actual), Equals, expected)
	}
}

func (s *S) TestStructMetaNil(c *C) {
	for _, in := range structMetaNilTests {
		c.Logf("test %s", in.expected)

		actual, err := Marshal(in.test)
		c.Assert(err, Equals, nil)
		c.Assert(string(actual), Equals, in.expected)
	}
}
