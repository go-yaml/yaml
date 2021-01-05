package yaml_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
	. "gopkg.in/check.v1"
)


func walkTree(indent int, node *yaml.Node) {
	fmt.Printf("%s{%d %d %#v:%#v anchor:%#v head:%#v line:%#v foot:%#v %d:%d}\n", strings.Repeat("  ", indent), node.Kind, node.Style, node.Tag, node.Value, node.Anchor, node.HeadComment, node.LineComment, node.FootComment, node.Line, node.Column)
	for _, item := range node.Content {
		walkTree(indent + 1, item)
	}
	if node.Alias != nil {
		walkTree(indent + 1, node.Alias)
	}
}


func testCycle(c *C, input string, expectedOutput string) {
	fmt.Printf("New single-doc input:\n")
	node := yaml.Node{}
	err := yaml.Unmarshal([]byte(input), &node)
	walkTree(0, &node)
	c.Assert(err, IsNil)
	var out []byte
	if node.IsZero() {
		out = []byte(nil)
	} else {
		out, err = yaml.Marshal(&node)
		c.Assert(err, IsNil)
	}
	c.Assert(string(out), DeepEquals, expectedOutput)
	if len(expectedOutput) == 0 {
		c.Assert(out, DeepEquals, []byte(nil))
	} else {
		c.Assert(out, DeepEquals, []byte(expectedOutput))
	}
}


func testIdempotent(c *C, data string) {
	testCycle(c, data, data)
}


func testMDCycle(c *C, input string, expectedOutput string) {
	fmt.Printf("New multi-doc input:\n")
	var nodes []*yaml.Node
	d := yaml.NewDecoder(bytes.NewReader([]byte(input)))
	for true {
		node := &yaml.Node{}
		err := d.Decode(node)
		if err == io.EOF {
			break
		}
		c.Assert(err, IsNil)
		nodes = append(nodes, node)
		walkTree(0, node)
	}
    var b bytes.Buffer
	e := yaml.NewEncoder(io.Writer(&b))
	e.SetIndent(4)
	for _, node := range nodes {
		err := e.Encode(node)
		c.Assert(err, IsNil)
	}
	e.Close()
	out := b.Bytes()
	c.Assert(string(out), DeepEquals, expectedOutput)
	if len(expectedOutput) == 0 {
		c.Assert(out, DeepEquals, []byte(nil))
	} else {
		c.Assert(out, DeepEquals, []byte(expectedOutput))
	}
}


func testMDIdempotent(c *C, data string) {
	testMDCycle(c, data, data)
}


func (s *S) TestCommentMoving1(c *C) {
	testIdempotent(c, `# begin
a:
    # foo
    # bar
    b:
    # baz
    c:
        foo: bar
        # asdf
    # bang
d:
    # a
    # b
    - 1
    # c
    - - 123
      # f
    # d
    - 2
    # e
`)
}


func (s *S) TestCommentMoving2(c *C) {
	// The newline is supposed to go away, but the comment should stay above `c: d`
	testCycle(c, `a:
    b:
        # comment followed by newline

        c: d
`, `a:
    b:
        # comment followed by newline
        c: d
`)
}


func (s *S) TestCommentParsing(c *C) {
	testIdempotent(c, `# beginning
a:
    ## foo
    ##
    b:
`)
}


func (s *S) TestCommentEmptyDoc(c *C) {
	testIdempotent(c, "# foo\n")
}


func (s *S) TestEmptyDocument(c *C) {
	testIdempotent(c, "")
	testCycle(c, "\n", "")
	testCycle(c, "---\n", "\n")
}


func (s *S) TestCommentDocSkip(c *C) {
	testMDIdempotent(c, `key: value

# foo
---
key: value
`)
	testMDIdempotent(c, `# foo
---
key: value
`)
	testMDIdempotent(c, `# foo
---
# bar
key: value
`)
}
