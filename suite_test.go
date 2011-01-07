package goyaml_test


import (
    . "gocheck"
    "testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})
