package yaml_test

import (
	"os"

	"github.com/Shopify/yaml"
	. "gopkg.in/check.v1"
)

var roundtripTests = []struct {
	value interface{}
	data  string
}{
	{
		yaml.MapSlice{{Key: "a", Value: 5, Comment: "hello"}},
		"a: 5 #hello",
	},
	{ // Not comments!
		yaml.MapSlice{{Key: "a", Value: "Hello #comment", Comment: ""}},
		"a: 'Hello #comment'\n",
	},
	{
		yaml.MapSlice{{Key: "a", Value: "你好 #comment", Comment: ""}},
		"a: '你好 #comment'\n",
	},
	{ // the must be seperated by other tokens by a whitespace char
		yaml.MapSlice{{Key: "a", Value: "5#hello", Comment: ""}},
		"a: 5#hello",
	},

	// Comments
	// Examples from the yaml spec:
	// TODO broken
	// { // 5.1. Byte Order Mark or 5.5. Comment Indicator
	// 	"#hello\n",
	// 	yaml.MapSlice{{Key: yaml.PreDoc("#hello"), Value: nil}},
	// },
	// TODO broken
	// { // 6.1. Indentation Spaces
	// 	"    #hello\n",
	// 	yaml.MapSlice{{Key: yaml.PreDoc("#hello"), Value: nil}},
	// },
	{ // 6.9. Separated Comment
		yaml.MapSlice{
			{Key: "a", Value: 1, Comment: "comment"},
		},
		"a: #comment\n 1",
	},
	// TODO broken
	// { // 6.11 Multi-Line Comments
	// 	yaml.MapSlice{{Key: "a", Value: 5, Comment: "hello\nbye"}},
	// 	"a: #hello\n#bye\n 5",
	// },
	// TODO broken
	// { // 6.12. Separation Spaces
	// 	yaml.MapSlice{{Key: yaml.MapSlice{{Key: "first", Value: "Sammy", Comment: ""}, {Key: "last", Value: "Sosa", Comment: ""}}, Value: yaml.MapSlice{{Key: nil, Value: nil, Comment: " Statistics:"}, {Key: "hr", Value: 65, Comment: " Home runs"}, {Key: "avg", Value: 0.278, Comment: " Average"}}, Comment: ""}},
	// 	"{ first: Sammy, last: Sosa }:\n # Statistics:\n   hr:  # Home runs\n      65\n   avg: # Average\n    0.278",
	// },
	// More comment tests
	{ // easy Predoc
		yaml.MapSlice{{Key: yaml.PreDoc("#hello"), Value: nil}, {Key: "a", Value: 5}},
		"#hello\na: 5",
	},
	{ // more elaborate PreDoc comment
		yaml.MapSlice{
			{Key: yaml.PreDoc("# comment 1\n\n---\n\n# comment 2"), Value: nil, Comment: ""},
			{Key: "a", Value: 1, Comment: ""},
		},
		"# comment 1\n\n---\n\n# comment 2\na: 1\n",
	},
	{ // PreDoc comment with %
		yaml.MapSlice{
			{Key: yaml.PreDoc("# pre doc comment 1\n\n%YAML   1.1\n---\n# pre doc comment 2"), Value: nil, Comment: ""},
			{Key: "a", Value: 1, Comment: ""},
		},
		"# pre doc comment 1\n\n%YAML   1.1\n---\n# pre doc comment 2\na: 1",
	},
	{ // primitive map value EOL comment
		yaml.MapSlice{{Key: "b", Value: 2, Comment: " my comment"}},
		"b: 2 # my comment\n",
	},
	{ // no space EOL before comment
		yaml.MapSlice{{Key: "a", Value: 5, Comment: "hello"}},
		"a: 5 #hello",
	},
	{ // EOL + trailing comment
		yaml.MapSlice{
			{Key: "a", Value: 5, Comment: "hello"},
			{Key: nil, Value: nil, Comment: "bye"},
		},
		"a: 5 #hello\n #bye",
	},
	{ // 2 trailing comments
		yaml.MapSlice{
			{Key: "a", Value: 5, Comment: ""},
			{Key: nil, Value: nil, Comment: " comment1"},
			{Key: nil, Value: nil, Comment: " comment2"},
		},
		"a: 5\n # comment1\n # comment2",
	},
	{ // map item comment
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: nil, Value: nil, Comment: " my comment"},
				{Key: "b", Value: 3, Comment: ""},
				{Key: nil, Value: nil, Comment: " my comment 2"},
				{Key: "c", Value: 8, Comment: ""},
			}, Comment: ""},
		},
		"a:\n  # my comment\n  b: 3\n  # my comment 2\n  c: 8\n",
	},
	{ // primitive sequence item EOL comment
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: "b", Value: []yaml.SequenceItem{
					{Value: 3, Comment: " my comment"},
				}, Comment: ""},
			}, Comment: ""},
		},
		"a:\n  b:\n  - 3 # my comment\n",
	},
	{ // sequence item comment
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
		"a:\n  b:\n  # my comment\n  - 3\n  # my comment 2\n  - 8\n",
	},
	{ // key comment (non-primitive value)
		yaml.MapSlice{
			{Key: "a", Value: yaml.MapSlice{
				{Key: "b", Value: []yaml.SequenceItem{
					{Value: 3, Comment: ""},
				}, Comment: " my comment"},
			}, Comment: ""},
		},
		"a:\n  b: # my comment\n  - 3\n",
	},
	{ // last line comment
		yaml.MapSlice{
			{Key: "a", Value: 1, Comment: ""},
			{Key: nil, Value: nil, Comment: " my comment"},
		},
		"a: 1\n# my comment\n",
	},
	{
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
		"b: # map key comment\n  # first line comment\n  c: 2 # end of line comment\n  # flow leading comment\n  d: [3, 4] # flow eol comment\n  # comment 1\n  e: # sequence key comment\n    # comment 2\n    - 3 # seq eol comment 1\n    - 4 # seq eol comment 2\n  # last line comment",
	},
}

func (s *S) TestRoundtrip(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range roundtripTests {
		// marshal
		c.Logf("test %d: %q", i, item.data)
		data, err := yaml.Marshal(item.value)
		c.Assert(err, IsNil)
		// unmarshal
		out, err := yaml.CommentUnmarshal([]byte(data))
		if _, ok := err.(*yaml.TypeError); !ok {
			c.Assert(err, IsNil)
		}
		c.Assert(out, DeepEquals, item.value, Commentf("error: %v", err))
	}
}
