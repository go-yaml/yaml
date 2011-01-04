package goyaml_test


import (
    . "gocheck"
    "testing"
    "goyaml"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}


func (s *S) TestHelloWorld(c *C) {
    data := []byte("hello: world")
    value := map[string]string{}
    err := goyaml.Unmarshal(data, value)
    c.Assert(err, IsNil)
    c.Assert(value["hello"], Equals, "world")
}
