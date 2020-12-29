package yaml_test

import (
	"gopkg.in/yaml.v3"
	. "gopkg.in/check.v1"
)


func testCycle(c *C, input string, expectedOutput string) {
	node := yaml.Node{}
	err := yaml.Unmarshal([]byte(input), &node)
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


func (s *S) TestCommentParsing(c *C) {
	testIdempotent(c, `# beginning
a:
    ## foo
    ##
    b:
`)
}
