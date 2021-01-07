package yaml

import (
	"bytes"
	"testing"
)

func TestMarshal(t *testing.T) {
	const yamlSample = `# head comment of a
a: # line comment of a
    a1: a1a1a1
    a2: a2a2a2
b: # line comment of b
    # head comment of b1
    b1: 100
    # foot comment of b1
c: ccccc # line comment of c
`

	buf := &Node{}
	if err := Unmarshal([]byte(yamlSample), buf); err != nil {
		t.Error(err.Error())
	} else if out, err := Marshal(buf); err != nil {
		t.Error(err.Error())
	} else if !bytes.Equal([]byte(yamlSample), out) {
		t.Error("inconsistent results")
	}

	// ---

	ex := &Node{
		Kind: MappingNode,
		Content: []*Node{
			&Node{
				Kind:        ScalarNode,
				Value:       "a",
				HeadComment: "head comment of a",
				LineComment: "line comment of a",
			},
			&Node{
				Kind: MappingNode,
				Content: []*Node{
					&Node{
						Kind:  ScalarNode,
						Value: "a1",
					},
					&Node{
						Kind:  ScalarNode,
						Value: "a1a1a1",
					},
					&Node{
						Kind:  ScalarNode,
						Value: "a2",
					},
					&Node{
						Kind:  ScalarNode,
						Value: "a2a2a2",
					},
				},
			},
			&Node{
				Kind:        ScalarNode,
				Value:       "b",
				LineComment: "line comment of b",
			},
			&Node{
				Kind: MappingNode,
				Content: []*Node{
					&Node{
						Kind:        ScalarNode,
						Value:       "b1",
						HeadComment: "head comment of b1",
						FootComment: "foot comment of b1",
					},
					&Node{
						Kind:  ScalarNode,
						Value: "100",
					},
				},
			},
			&Node{
				Kind:  ScalarNode,
				Value: "c",
			},
			&Node{
				Kind:        ScalarNode,
				Value:       "ccccc",
				LineComment: "line comment of c",
			},
		},
	}

	if out, err := Marshal(ex); err != nil {
		t.Error(err.Error())
	} else if !bytes.Equal([]byte(yamlSample), out) {
		t.Error("inconsistent results")
	}
}
