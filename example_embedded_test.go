package yaml

import (
	"bytes"
	"fmt"
	"log"
)

// An example showing how to unmarshal embedded
// structs from YAML.

type StructA struct {
	A string `yaml:"a"`
}

type StructB struct {
	// Embedded structs are not treated as embedded in YAML by default. To do that,
	// add the ",inline" annotation below
	StructA `yaml:",inline"`
	B       string `yaml:"b"`
}

var data = `
a: a string from struct A
b: a string from struct B
`

func ExampleUnmarshal_embedded() {
	var b StructB

	err := Unmarshal([]byte(data), &b)
	if err != nil {
		log.Fatal("cannot unmarshal data: ", err)
	}
	fmt.Println(b.A)
	fmt.Println(b.B)
	// Output:
	// a string from struct A
	// a string from struct B
}

func ExampleDecoder_embedded() {
	var b StructB

	buf := bytes.NewBufferString(data)
	err := NewDecoder(buf).Decode(&b)
	if err != nil {
		log.Fatal("cannot unmarshal data: ", err)
	}
	fmt.Println(b.A)
	fmt.Println(b.B)
	// Output:
	// a string from struct A
	// a string from struct B
}
