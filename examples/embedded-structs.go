package main

import (
	"fmt"
	"log"

        "gopkg.in/yaml.v2"
)

type StructA struct {
	A string `yaml:"a"`
}

type StructB struct {
	// go-yaml will not decode embedded structs by default, to do that
	// you need to add the ",inline" annotation below
	StructA   `yaml:",inline"`
	B string `yaml:"b"`
}

var data = `
a: a string from struct A
b: a string from struct B
`

func main() {
	var b StructB

	err := yaml.Unmarshal([]byte(data), &b)
	if err != nil {
                log.Fatalf("error: %v", err)
	}
        fmt.Println(b.A)
        fmt.Println(b.B)
}
