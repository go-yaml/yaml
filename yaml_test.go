package yaml_test

import (
	"fmt"
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
	node := yaml.Node{}
	err := yaml.Unmarshal([]byte(input), &node)
	walkTree(0, &node)
	c.Assert(err, IsNil)
	out, err := yaml.Marshal(&node)
	c.Assert(err, IsNil)
	c.Assert(string(out), DeepEquals, expectedOutput)
	c.Assert(out, DeepEquals, []byte(expectedOutput))
}


func testIdempotent(c *C, data string) {
	testCycle(c, data, data)
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
	testIdempotent(c, `a:
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
	testIdempotent(c, `# foo
`)
}
