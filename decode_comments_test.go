package yaml_test

import (
	"math"
	"strings"
	"time"

	"github.com/Shopify/yaml"
	. "gopkg.in/check.v1"
)

var unmarshalCommentsTests = []struct {
	data  string
	value interface{}
}{
	{
		"",
		yaml.MapSlice{},
	},
	//{
	//	"{}",
	//	yaml.MapSlice{},
	//},
	{
		"v: hi",
		yaml.MapSlice{{Key: "v", Value: "hi", Comment: ""}},
	},
	{
		"v: true",
		yaml.MapSlice{{Key: "v", Value: true, Comment: ""}},
	},
	{
		"v: 10",
		yaml.MapSlice{{Key: "v", Value: 10, Comment: ""}},
	},

	{
		"v: 0b10",
		yaml.MapSlice{{Key: "v", Value: 2, Comment: ""}},
	},
	{
		"v: 0xA",
		yaml.MapSlice{{Key: "v", Value: 10, Comment: ""}},
	},
	{
		"v: 4294967296",
		yaml.MapSlice{{Key: "v", Value: 4294967296, Comment: ""}},
	},
	{
		"v: 0.1",
		yaml.MapSlice{{Key: "v", Value: 0.1, Comment: ""}},
	},
	{
		"v: .1",
		yaml.MapSlice{{Key: "v", Value: 0.1, Comment: ""}},
	},
	{
		"v: .Inf",
		yaml.MapSlice{{Key: "v", Value: math.Inf(+1), Comment: ""}},
	},
	{
		"v: -.Inf",
		yaml.MapSlice{{Key: "v", Value: math.Inf(-1), Comment: ""}},
	},
	{
		"v: -10",
		yaml.MapSlice{{Key: "v", Value: -10, Comment: ""}},
	},
	{
		"v: -.1",
		yaml.MapSlice{{Key: "v", Value: -0.1, Comment: ""}},
	},

	// Simple values.
	// TODO: Add support for plain text in yaml
	//{
	//	"123",
	//	yaml.MapSlice{{Key: nil, Value: &unmarshalIntTest, Comment: ""}},
	//},

	// Floats from spec
	{
		"canonical: 6.8523e+5",
		yaml.MapSlice{{Key: "canonical", Value: 6.8523e+5, Comment: ""}},
	},
	{
		"expo: 685.230_15e+03",
		yaml.MapSlice{{Key: "expo", Value: 685.23015e+03, Comment: ""}},
	},
	{
		"fixed: 685_230.15",
		yaml.MapSlice{{Key: "fixed", Value: 685230.15, Comment: ""}},
	},
	{
		"neginf: -.inf",
		yaml.MapSlice{{Key: "neginf", Value: math.Inf(-1), Comment: ""}},
	},
	//{"sexa: 190:20:30.15", map[string]interface{}{"sexa": 0}}, // Unsupported
	//{"notanum: .NaN", map[string]interface{}{"notanum": math.NaN()}}, // Equality of NaN fails.

	// Bools from spec
	{
		"canonical: y",
		yaml.MapSlice{{Key: "canonical", Value: true, Comment: ""}},
	},
	{
		"answer: NO",
		yaml.MapSlice{{Key: "answer", Value: false, Comment: ""}},
	},
	{
		"logical: True",
		yaml.MapSlice{{Key: "logical", Value: true, Comment: ""}},
	},
	{
		"option: on",
		yaml.MapSlice{{Key: "option", Value: true, Comment: ""}},
	},
	// Ints from spec
	{
		"canonical: 685230",
		yaml.MapSlice{{Key: "canonical", Value: 685230, Comment: ""}},
	},
	{
		"decimal: +685_230",
		yaml.MapSlice{{Key: "decimal", Value: 685230, Comment: ""}},
	},
	{
		"octal: 02472256",
		yaml.MapSlice{{Key: "octal", Value: 685230, Comment: ""}},
	},
	{
		"hexa: 0x_0A_74_AE",
		yaml.MapSlice{{Key: "hexa", Value: 685230, Comment: ""}},
	},
	{
		"bin: 0b1010_0111_0100_1010_1110",
		yaml.MapSlice{{Key: "bin", Value: 685230, Comment: ""}},
	},
	{
		"bin: -0b101010",
		yaml.MapSlice{{Key: "bin", Value: -42, Comment: ""}},
	},
	{
		"bin: -0b1000000000000000000000000000000000000000000000000000000000000000",
		yaml.MapSlice{{Key: "bin", Value: -9223372036854775808, Comment: ""}},
	},
	//{"sexa: 190:20:30", map[string]interface{}{"sexa": 0}}, // Unsupported

	// Nulls from spec
	{
		"empty:",
		yaml.MapSlice{{Key: "empty", Value: nil, Comment: ""}},
	},
	{
		"canonical: ~",
		yaml.MapSlice{{Key: "canonical", Value: nil, Comment: ""}},
	},
	{
		"english: null",
		yaml.MapSlice{{Key: "english", Value: nil, Comment: ""}},
	},
	{
		"~: null key",
		yaml.MapSlice{{Key: nil, Value: "null key", Comment: ""}},
	},

	// Flow sequence
	{
		"seq: [A,B]",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: "B", Comment: ""}}, Comment: ""}},
	},
	{
		"seq: [A,B,C,]",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: "B", Comment: ""}, yaml.SequenceItem{Value: "C", Comment: ""}}, Comment: ""}},
	},
	{
		"seq: [A,1,C]",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: 1, Comment: ""}, yaml.SequenceItem{Value: "C", Comment: ""}}, Comment: ""}},
	},

	// Block sequence
	{
		"seq:\n - A\n - B",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: "B", Comment: ""}}, Comment: ""}},
	},
	{
		"seq:\n - A\n - B\n - C",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: "B", Comment: ""}, yaml.SequenceItem{Value: "C", Comment: ""}}, Comment: ""}},
	},
	{
		"seq:\n - A\n - 1\n - C",
		yaml.MapSlice{{Key: "seq", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: 1, Comment: ""}, yaml.SequenceItem{Value: "C", Comment: ""}}, Comment: ""}},
	},

	// Literal block scalar
	{
		"scalar: | # Comment\n\n literal\n\n \ttext\n\n",
		yaml.MapSlice{{Key: "scalar", Value: "\nliteral\n\n\ttext\n", Comment: ""}},
	},

	// Folded block scalar
	// TODO: add support for comments in folded style
	// {
	// 	"scalar: > # Comment\n\n folded\n line\n \n next\n line\n  * one\n  * two\n\n last\n line\n\n",
	// 	yaml.MapSlice{{Key: "scalar", Value: "\nfolded line\nnext line\n * one\n * two\n\nlast line\n", Comment: " Comment"}},
	// },

	// Map inside interface with no type hints.
	{
		"a: {b: c}",
		yaml.MapSlice{{Key: "a", Value: yaml.MapSlice{yaml.MapItem{Key: "b", Value: "c", Comment: ""}}, Comment: ""}},
	},

	// Some cross type conversions
	{
		"v: 42",
		yaml.MapSlice{{Key: "v", Value: 42, Comment: ""}},
	},
	{
		"v: -42",
		yaml.MapSlice{{Key: "v", Value: -42, Comment: ""}},
	},
	{
		"v: 4294967296",
		yaml.MapSlice{{Key: "v", Value: 4294967296, Comment: ""}},
	},
	{
		"v: -4294967296",
		yaml.MapSlice{{Key: "v", Value: -4294967296, Comment: ""}},
	},

	// int
	{
		"int_max: 2147483647",
		yaml.MapSlice{{Key: "int_max", Value: math.MaxInt32, Comment: ""}},
	},
	{
		"int_min: -2147483648",
		yaml.MapSlice{{Key: "int_min", Value: math.MinInt32, Comment: ""}},
	},

	// int64
	{
		"int64_max: 9223372036854775807",
		yaml.MapSlice{{Key: "int64_max", Value: math.MaxInt64, Comment: ""}},
	},
	{
		"int64_max_base2: 0b111111111111111111111111111111111111111111111111111111111111111",
		yaml.MapSlice{{Key: "int64_max_base2", Value: math.MaxInt64, Comment: ""}},
	},
	{
		"int64_min: -9223372036854775808",
		yaml.MapSlice{{Key: "int64_min", Value: math.MinInt64, Comment: ""}},
	},
	{
		"int64_neg_base2: -0b111111111111111111111111111111111111111111111111111111111111111",
		yaml.MapSlice{{Key: "int64_neg_base2", Value: -math.MaxInt64, Comment: ""}},
	},

	// uint
	{
		"uint_min: 0",
		yaml.MapSlice{{Key: "uint_min", Value: 0, Comment: ""}},
	},
	{
		"uint_max: 4294967295",
		yaml.MapSlice{{Key: "uint_max", Value: math.MaxUint32, Comment: ""}},
	},

	// uint64
	{
		"uint64_min: 0",
		yaml.MapSlice{{Key: "uint64_min", Value: 0, Comment: ""}},
	},
	{
		"uint64_max: 18446744073709551615",
		yaml.MapSlice{{Key: "uint64_max", Value: uint64(math.MaxUint64), Comment: ""}},
	},
	{
		"uint64_max_base2: 0b1111111111111111111111111111111111111111111111111111111111111111",
		yaml.MapSlice{{Key: "uint64_max_base2", Value: uint64(math.MaxUint64), Comment: ""}},
	},
	{
		"uint64_maxint64: 9223372036854775807",
		yaml.MapSlice{{Key: "uint64_maxint64", Value: math.MaxInt64, Comment: ""}},
	},

	// float32
	{
		"float32_max: 3.40282346638528859811704183484516925440e+38",
		yaml.MapSlice{{Key: "float32_max", Value: math.MaxFloat32, Comment: ""}},
	},
	{
		"float32_nonzero: 1.401298464324817070923729583289916131280e-45",
		yaml.MapSlice{{Key: "float32_nonzero", Value: math.SmallestNonzeroFloat32, Comment: ""}},
	},

	// float64
	{
		"float64_max: 1.797693134862315708145274237317043567981e+308",
		yaml.MapSlice{{Key: "float64_max", Value: math.MaxFloat64, Comment: ""}},
	},
	{
		"float64_nonzero: 4.940656458412465441765687928682213723651e-324",
		yaml.MapSlice{{Key: "float64_nonzero", Value: math.SmallestNonzeroFloat64, Comment: ""}},
	},
	{
		"float64_maxuint64: 18446744073709551615.0",
		yaml.MapSlice{{Key: "float64_maxuint64", Value: float64(math.MaxUint64), Comment: ""}},
	},
	{
		"float64_maxuint64+1: 18446744073709551616",
		yaml.MapSlice{{Key: "float64_maxuint64+1", Value: float64(math.MaxUint64 + 1), Comment: ""}},
	},
	// Quoted values.
	{
		"'1': '\"2\"'",
		yaml.MapSlice{{Key: "1", Value: "\"2\"", Comment: ""}},
	},
	{
		"v:\n- A\n- 'B\n\n  C'\n",
		yaml.MapSlice{{Key: "v", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: "A", Comment: ""}, yaml.SequenceItem{Value: "B\nC", Comment: ""}}, Comment: ""}},
	},

	// Explicit tags.
	{
		"v: !!float '1.1'",
		yaml.MapSlice{{Key: "v", Value: 1.1, Comment: ""}},
	},
	{
		"v: !!float 0",
		yaml.MapSlice{{Key: "v", Value: float64(0), Comment: ""}},
	},
	{
		"v: !!float -1",
		yaml.MapSlice{{Key: "v", Value: float64(-1), Comment: ""}},
	},
	{
		"v: !!null ''",
		yaml.MapSlice{{Key: "v", Value: nil, Comment: ""}},
	},

	// Non-specific tag (Issue #75)
	{
		"v: ! test",
		yaml.MapSlice{{Key: "v", Value: "test", Comment: ""}},
	},

	// Anchors and aliases.
	{
		"a: &x 1\nb: &y 2\nc: *x\nd: *y\n",
		yaml.MapSlice{{Key: "a", Value: 1, Comment: ""}, yaml.MapItem{Key: "b", Value: 2, Comment: ""}, yaml.MapItem{Key: "c", Value: 1, Comment: ""}, yaml.MapItem{Key: "d", Value: 2, Comment: ""}},
	},
	{
		"a: &a {c: 1}\nb: *a",
		yaml.MapSlice{{Key: "a", Value: yaml.MapSlice{yaml.MapItem{Key: "c", Value: 1, Comment: ""}}, Comment: ""}, yaml.MapItem{Key: "b", Value: yaml.MapSlice{yaml.MapItem{Key: "c", Value: 1, Comment: ""}}, Comment: ""}},
	},
	{
		"a: &a [1, 2]\nb: *a",
		yaml.MapSlice{{Key: "a", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: 1, Comment: ""}, yaml.SequenceItem{Value: 2, Comment: ""}}, Comment: ""}, yaml.MapItem{Key: "b", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: 1, Comment: ""}, yaml.SequenceItem{Value: 2, Comment: ""}}, Comment: ""}},
	},

	// Bug #1133337
	{
		"foo: ''",
		yaml.MapSlice{{Key: "foo", Value: "", Comment: ""}},
	},
	{
		"foo: null",
		yaml.MapSlice{{Key: "foo", Value: nil, Comment: ""}},
	},

	// Support for ~
	{
		"foo: ~",
		yaml.MapSlice{{Key: "foo", Value: nil, Comment: ""}},
	},

	// TODO: Add support for plain text in yaml
	// // Bug #1191981
	// {
	// 	"" +
	// 		"%YAML 1.1\n" +
	// 		"--- !!str\n" +
	// 		`"Generic line break (no glyph)\n\` + "\n" +
	// 		` Generic line break (glyphed)\n\` + "\n" +
	// 		` Line separator\u2028\` + "\n" +
	// 		` Paragraph separator\u2029"` + "\n",
	// 	"" +
	// 		"Generic line break (no glyph)\n" +
	// 		"Generic line break (glyphed)\n" +
	// 		"Line separator\u2028Paragraph separator\u2029",
	// },

	// bug 1243827
	{
		"a: -b_c",
		yaml.MapSlice{{Key: "a", Value: "-b_c", Comment: ""}},
	},

	// issue #295 (allow scalars with colons in flow mappings and sequences)
	{
		"a: {b: https://github.com/go-yaml/yaml}",
		yaml.MapSlice{{
			Key:     "a",
			Value:   yaml.MapSlice{{Key: "b", Value: "https://github.com/go-yaml/yaml", Comment: ""}},
			Comment: "",
		}},
	},
	{
		"a: [https://github.com/go-yaml/yaml]",
		yaml.MapSlice{{
			Key: "a",
			Value: []yaml.SequenceItem{yaml.SequenceItem{
				Value:   "https://github.com/go-yaml/yaml",
				Comment: ""},
			},
			Comment: ""},
		},
	},

	// Issue #24.
	{
		"a: <foo>",
		yaml.MapSlice{{Key: "a", Value: "<foo>", Comment: ""}},
	},

	// Base 60 floats are obsolete and unsupported.
	{
		"a: 1:1\n",
		yaml.MapSlice{{Key: "a", Value: "1:1", Comment: ""}},
	},

	// Binary data.
	{
		"a: !!binary gIGC\n",
		yaml.MapSlice{{Key: "a", Value: "\x80\x81\x82", Comment: ""}},
	}, {
		"a: !!binary |\n  " + strings.Repeat("kJCQ", 17) + "kJ\n  CQ\n",
		yaml.MapSlice{{Key: "a", Value: strings.Repeat("\x90", 54), Comment: ""}},
	}, {
		"a: !!binary |\n  " + strings.Repeat("A", 70) + "\n  ==\n",
		yaml.MapSlice{{Key: "a", Value: strings.Repeat("\x00", 52), Comment: ""}},
	},

	// Timestamps
	{
		// Date only.
		"a: 2015-01-01\n",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Comment: ""}},
	},
	{
		// RFC3339
		"a: 2015-02-24T18:19:39.12Z\n",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 2, 24, 18, 19, 39, .12e9, time.UTC), Comment: ""}},
	},
	{
		// RFC3339 with short dates.
		"a: 2015-2-3T3:4:5Z",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 2, 3, 3, 4, 5, 0, time.UTC), Comment: ""}},
	},
	{
		// ISO8601 lower case t
		"a: 2015-02-24t18:19:39Z\n",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 2, 24, 18, 19, 39, 0, time.UTC), Comment: ""}},
	},
	{
		// space separate, no time zone
		"a: 2015-02-24 18:19:39\n",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 2, 24, 18, 19, 39, 0, time.UTC), Comment: ""}},
	},
	// Some cases not currently handled. Uncomment these when
	// the code is fixed.
	// {
	// 	// space separated with time zone
	// 	"a: 2001-12-14 21:59:43.10 -5",
	// 	yaml.MapSlice{{Key: "a", Value: time.Date(2001, 12, 14, 21, 59, 43, .1e9, time.UTC), Comment: ""}},
	// },
	// {
	// 	// arbitrary whitespace between fields
	// 	"a: 2001-12-14 \t\t \t21:59:43.10 \t Z",
	// 	yaml.MapSlice{{Key: "a", Value: time.Date(2001, 12, 14, 21, 59, 43, .1e9, time.UTC), Comment: ""}},
	// },
	{
		// explicit string tag
		"a: !!str 2015-01-01",
		yaml.MapSlice{{Key: "a", Value: "2015-01-01", Comment: ""}},
	},
	{
		// explicit timestamp tag on quoted string
		"a: !!timestamp \"2015-01-01\"",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Comment: ""}},
	},
	{
		// explicit timestamp tag on unquoted string
		"a: !!timestamp 2015-01-01",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Comment: ""}},
	},
	{
		// quoted string that's a valid timestamp
		"a: \"2015-01-01\"",
		yaml.MapSlice{{Key: "a", Value: "2015-01-01", Comment: ""}},
	},
	{
		// implicit timestamp tag.
		"a: 2015-01-01",
		yaml.MapSlice{{Key: "a", Value: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), Comment: ""}},
	},

	// Encode empty lists as zero-length slices.
	{
		"a: []",
		yaml.MapSlice{{Key: "a", Value: []yaml.SequenceItem(nil), Comment: ""}},
	},

	// UTF-16-LE
	{
		"\xff\xfe\xf1\x00o\x00\xf1\x00o\x00:\x00 \x00v\x00e\x00r\x00y\x00 \x00y\x00e\x00s\x00\n\x00",
		yaml.MapSlice{{Key: "ñoño", Value: "very yes", Comment: ""}},
	},
	// UTF-16-LE with surrogate.
	{
		"\xff\xfe\xf1\x00o\x00\xf1\x00o\x00:\x00 \x00v\x00e\x00r\x00y\x00 \x00y\x00e\x00s\x00 \x00=\xd8\xd4\xdf\n\x00",
		yaml.MapSlice{{Key: "ñoño", Value: "very yes 🟔", Comment: ""}},
	},
	// UTF-16-BE
	{
		"\xfe\xff\x00\xf1\x00o\x00\xf1\x00o\x00:\x00 \x00v\x00e\x00r\x00y\x00 \x00y\x00e\x00s\x00\n",
		yaml.MapSlice{{Key: "ñoño", Value: "very yes", Comment: ""}},
	},
	// UTF-16-BE with surrogate.
	{
		"\xfe\xff\x00\xf1\x00o\x00\xf1\x00o\x00:\x00 \x00v\x00e\x00r\x00y\x00 \x00y\x00e\x00s\x00 \xd8=\xdf\xd4\x00\n",
		yaml.MapSlice{{Key: "ñoño", Value: "very yes 🟔", Comment: ""}},
	},

	// YAML Float regex shouldn't match this
	{
		"a: 123456e1\n",
		yaml.MapSlice{{Key: "a", Value: "123456e1", Comment: ""}},
	}, {
		"a: 123456E1\n",
		yaml.MapSlice{{Key: "a", Value: "123456E1", Comment: ""}},
	},
	// yaml-test-suite 3GZX: Spec Example 7.1. Alias Nodes
	{
		"First occurrence: &anchor Foo\nSecond occurrence: *anchor\nOverride anchor: &anchor Bar\nReuse anchor: *anchor\n",
		yaml.MapSlice{
			{Key: "First occurrence", Value: "Foo", Comment: ""},
			{Key: "Second occurrence", Value: "Foo", Comment: ""},
			{Key: "Override anchor", Value: "Bar", Comment: ""},
			{Key: "Reuse anchor", Value: "Bar", Comment: ""},
		},
	},
	// TODO: Add support for plain text in yaml
	// Single document with garbage following it.
	// {
	// 	"---\nhello\n...\n}not yaml",
	// 	yaml.MapSlice{{Key: yaml.PreDoc("---\nhello\n...\n}not yaml"), Value: nil, Comment: ""}},
	// },
	{
		"a: 5\n",
		yaml.MapSlice{{Key: "a", Value: 5, Comment: ""}},
	}, {
		"a: 5.5\n",
		yaml.MapSlice{{Key: "a", Value: 5.5, Comment: ""}},
	},

	// Not comments!
	{
		"a: 'Hello #comment'\n",
		yaml.MapSlice{{Key: "a", Value: "Hello #comment", Comment: ""}},
	},
	{
		"a: '你好 #comment'\n",
		yaml.MapSlice{{Key: "a", Value: "你好 #comment", Comment: ""}},
	},
	// the commnt must be seperated from other tokens by a whitespace char
	{
		"a: 5#hello",
		yaml.MapSlice{{Key: "a", Value: "5#hello", Comment: ""}},
	},

	// Comments

	// Examples from the yaml spec:
	// 5.1. Byte Order Mark or 5.5. Comment Indicator
	{
		"#hello\n",
		yaml.MapSlice{{Key: yaml.PreDoc("#hello"), Value: nil}},
	},
	{ // 6.1. Indentation Spaces
		"    #hello\n    a: b",
		yaml.MapSlice{{Key: yaml.PreDoc("    #hello\n    "), Value: nil}, {Key: "a", Value: "b"}},
	},
	// 6.9. Separated Comment
	{
		"a: #comment\n 1",
		yaml.MapSlice{
			{Key: "a", Value: 1, Comment: "comment"},
		},
	},
	// 6.11 Multi-Line Comments
	{
		"a: #hello\n#bye\n 5",
		yaml.MapSlice{{Key: "a", Value: 5, Comment: "hello;bye"}},
	},
	// 6.12. Separation Spaces
	{
		"{ first: Sammy, last: Sosa }:\n # Statistics:\n   hr:  # Home runs\n      65\n   avg: # Average\n    0.278",
		yaml.MapSlice{{
			Key: yaml.MapSlice{
				{Key: "first", Value: "Sammy", Comment: ""},
				{Key: "last", Value: "Sosa", Comment: ""},
			},
			Value: yaml.MapSlice{
				{Key: nil, Value: nil, Comment: " Statistics:"},
				{Key: "hr", Value: 65, Comment: " Home runs"},
				{Key: "avg", Value: 0.278, Comment: " Average"},
			}, Comment: "",
		}},
	},
	// TODO: add support for comments in literal style
	// {
	// 	"# Strip\n  # Comments:\nstrip: |-\n  # text\n\n  \n\n # Clip\n  # comments:\n\n\nclip: |\n  # text\n\n \n\n # Keep\n  # comments:\n\n\nkeep: |+\n  # text\n\n\n\n # Trail\n  # comments.\n",
	// 	yaml.MapSlice{{Key: yaml.PreDoc("# Strip\n  # Comments:"), Value: nil, Comment: ""}, {Key: "strip", Value: "# text", Comment: " Clip"}, {Key: nil, Value: nil, Comment: " comments:"}, {Key: "clip", Value: "# text\n", Comment: " Keep"}, {Key: nil, Value: nil, Comment: " comments:"}, {Key: "keep", Value: "# text\n\n\n\n", Comment: " Trail"}, {Key: nil, Value: nil, Comment: " comments."}},
	// },
	// More comment tests
	// easy Predoc
	{
		"#hello\na: 5",
		yaml.MapSlice{{Key: yaml.PreDoc("#hello"), Value: nil}, {Key: "a", Value: 5}},
	},
	// more elaborate PreDoc comment
	{
		"# comment 1\n\n---\n\n# comment 2\na: 1\n",
		yaml.MapSlice{
			{Key: yaml.PreDoc("# comment 1\n\n---\n\n# comment 2"), Value: nil, Comment: ""},
			{Key: "a", Value: 1, Comment: ""},
		},
	},
	// PreDoc comment with %
	{
		"# pre doc comment 1\n\n%YAML   1.1\n---\n# pre doc comment 2\na: 1",
		yaml.MapSlice{
			{Key: yaml.PreDoc("# pre doc comment 1\n\n%YAML   1.1\n---\n# pre doc comment 2"), Value: nil, Comment: ""},
			{Key: "a", Value: 1, Comment: ""},
		},
	},
	// primitive map value EOL comment
	{
		"b: 2 # my comment\n",
		yaml.MapSlice{
			{Key: "b", Value: 2, Comment: " my comment"},
		},
	},
	// no space EOL before comment
	{
		"a: 5 #hello",
		yaml.MapSlice{{Key: "a", Value: 5, Comment: "hello"}},
	},
	// EOL + trailing comment
	{
		"a: 5 #hello\n #bye",
		yaml.MapSlice{
			{Key: "a", Value: 5, Comment: "hello"},
			{Key: nil, Value: nil, Comment: "bye"},
		},
	},
	// 2 trailing comments
	{
		"a: 5\n # comment1\n # comment2",
		yaml.MapSlice{
			{Key: "a", Value: 5, Comment: ""},
			{Key: nil, Value: nil, Comment: " comment1"},
			{Key: nil, Value: nil, Comment: " comment2"},
		},
	},
	// map item comment
	{
		"a:\n  # my comment\n  b: 3\n  # my comment 2\n  c: 8\n",
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: nil, Value: nil, Comment: " my comment"},
				{Key: "b", Value: 3, Comment: ""},
				{Key: nil, Value: nil, Comment: " my comment 2"},
				{Key: "c", Value: 8, Comment: ""},
			}, Comment: ""},
		},
	},
	{ // primitive sequence item EOL comment
		"a:\n  b:\n  - 3 # my comment\n",
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: "b", Value: []yaml.SequenceItem{
					{Value: 3, Comment: " my comment"},
				}, Comment: ""},
			}, Comment: ""},
		},
	},
	// sequence item comment
	{
		"a:\n  b:\n  # my comment\n  - 3\n  # my comment 2\n  - 8\n",
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: "b", Value: []yaml.SequenceItem{
					{Value: nil, Comment: " my comment"},
					{Value: 3, Comment: ""},
					{Value: nil, Comment: " my comment 2"},
					{Value: 8, Comment: ""},
				}, Comment: ""},
			}, Comment: ""},
		},
	},
	// key comment (non-primitive value)
	{
		"a:\n  b: # my comment\n  - 3\n",
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: "b", Value: []yaml.SequenceItem{
					{Value: 3, Comment: ""},
				}, Comment: " my comment"},
			}, Comment: ""},
		},
	},
	// last line comment
	{
		"a: 1\n# my comment\n",
		yaml.MapSlice{
			{Key: "a", Value: 1, Comment: ""},
			{Key: nil, Value: nil, Comment: " my comment"},
		},
	},
	{
		"b: # map key comment\n  # first line comment\n  c: 2 # end of line comment\n  # flow leading comment\n  d: [3, 4] # flow eol comment\n  # comment 1\n  e: # sequence key comment\n    # comment 2\n    - 3 # seq eol comment 1\n    - 4 # seq eol comment 2\n  # last line comment",
		yaml.MapSlice{
			{Key: "b", Value: yaml.MapSlice{
				{Key: nil, Value: nil, Comment: " first line comment"},
				{Key: "c", Value: 2, Comment: " end of line comment"},
				{Key: nil, Value: nil, Comment: " flow leading comment"},
				{Key: nil, Value: nil, Comment: " flow eol comment"},
				{Key: "d", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: 3, Comment: ""}, yaml.SequenceItem{Value: 4, Comment: ""}}, Comment: ""},
				{Key: nil, Value: nil, Comment: " comment 1"},
				{Key: "e", Value: []yaml.SequenceItem{yaml.SequenceItem{Value: nil, Comment: " comment 2"}, yaml.SequenceItem{Value: 3, Comment: " seq eol comment 1"}, yaml.SequenceItem{Value: 4, Comment: " seq eol comment 2"}}, Comment: " sequence key comment"},
			}, Comment: " map key comment",
			},
			{Key: nil, Value: nil, Comment: " last line comment"},
		},
	},
}

func (s *S) TestCommentsUnmarshal(c *C) {
	for i, item := range unmarshalCommentsTests {
		c.Logf("test %d: %q", i, item.data)
		out, err := yaml.CommentUnmarshal([]byte(item.data))
		if _, ok := err.(*yaml.TypeError); !ok {
			c.Assert(err, IsNil)
		}
		c.Assert(out, DeepEquals, item.value, Commentf("error: %v", err))
	}
}
