//
// Copyright (c) 2011-2019 Canonical Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	. "gopkg.in/check.v1"
	"gopkg.in/yaml.v3"
)

type action int

const (
	doBoth action = iota // encode and decode
	decodeOnly
	encodeOnly
)

var nodeTests = map[string]struct {
	do   action
	yaml string
	node yaml.Node
}{
	"t00": {doBoth, `
null
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "null",
				Tag:    "!!null",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t01": {encodeOnly, `
null
`,
		yaml.Node{},
	}, "t02": {doBoth, `
foo
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "foo",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t03": {doBoth, `
"foo"
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.DoubleQuotedStyle,
				Value:  "foo",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t04": {doBoth, `
'foo'
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.SingleQuotedStyle,
				Value:  "foo",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t05": {doBoth, `
!!str 123
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.TaggedStyle,
				Value:  "123",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	},
	// Although the node isn't TaggedStyle, dropping the tag would change the value.
	"t06": {encodeOnly, `
!!binary gIGC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "gIGC",
				Tag:    "!!binary",
				Line:   1,
				Column: 1,
			}},
		},
	},
	// Item doesn't have a tag, but needs to be binary encoded due to its content.
	"t07": {encodeOnly, `
!!binary gIGC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "\x80\x81\x82",
				Line:   1,
				Column: 1,
			}},
		},
	},
	// Same, but with strings we can just quote them.
	"t08": {encodeOnly, `
"123"
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "123",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t09": {doBoth, `
!tag:something 123
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.TaggedStyle,
				Value:  "123",
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t10": {encodeOnly, `
!tag:something 123
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "123",
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t11": {doBoth, `
!tag:something {}
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.TaggedStyle | yaml.FlowStyle,
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t12": {encodeOnly, `
!tag:something {}
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.FlowStyle,
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t13": {doBoth, `
!tag:something []
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.TaggedStyle | yaml.FlowStyle,
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t14": {encodeOnly, `
!tag:something []
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Tag:    "!tag:something",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t15": {doBoth, `
''
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.SingleQuotedStyle,
				Value:  "",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t16": {doBoth, `
|
  foo
  bar
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.LiteralStyle,
				Value:  "foo\nbar\n",
				Tag:    "!!str",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t17": {doBoth, `
true
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "true",
				Tag:    "!!bool",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t18": {doBoth, `
-10
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "-10",
				Tag:    "!!int",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t19": {doBoth, `
4294967296
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "4294967296",
				Tag:    "!!int",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t20": {doBoth, `
0.1000
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "0.1000",
				Tag:    "!!float",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t21": {doBoth, `
-.inf
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "-.inf",
				Tag:    "!!float",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t22": {doBoth, `
.nan
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  ".nan",
				Tag:    "!!float",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t23": {doBoth, `
{}
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.FlowStyle,
				Value:  "",
				Tag:    "!!map",
				Line:   1,
				Column: 1,
			}},
		},
	}, "t24": {doBoth, `
a: b c
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Value:  "",
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b c",
					Tag:    "!!str",
					Line:   1,
					Column: 4,
				}},
			}},
		},
	}, "t25": {doBoth, `
a:
  b: c
  d: e
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Tag:    "!!str",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Tag:    "!!str",
						Line:   2,
						Column: 6,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "d",
						Tag:    "!!str",
						Line:   3,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "e",
						Tag:    "!!str",
						Line:   3,
						Column: 6,
					}},
				}},
			}},
		},
	}, "t26": {doBoth, `
a:
  - b: c
    d: e
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.MappingNode,
						Tag:    "!!map",
						Line:   2,
						Column: 5,
						Content: []*yaml.Node{{
							Kind:   yaml.ScalarNode,
							Value:  "b",
							Tag:    "!!str",
							Line:   2,
							Column: 5,
						}, {
							Kind:   yaml.ScalarNode,
							Value:  "c",
							Tag:    "!!str",
							Line:   2,
							Column: 8,
						}, {
							Kind:   yaml.ScalarNode,
							Value:  "d",
							Tag:    "!!str",
							Line:   3,
							Column: 5,
						}, {
							Kind:   yaml.ScalarNode,
							Value:  "e",
							Tag:    "!!str",
							Line:   3,
							Column: 8,
						}},
					}},
				}},
			}},
		},
	}, "t27": {doBoth, `
a: # AI
  - b
c:
  - d
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "a",
					LineComment: "# AI",
					Line:        1,
					Column:      1,
				}, {
					Kind: yaml.SequenceNode,
					Tag:  "!!seq",
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "b",
						Line:   2,
						Column: 5,
					}},
					Line:   2,
					Column: 3,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "c",
					Line:   3,
					Column: 1,
				}, {
					Kind: yaml.SequenceNode,
					Tag:  "!!seq",
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "d",
						Line:   4,
						Column: 5,
					}},
					Line:   4,
					Column: 3,
				}},
			}},
		},
	}, "t28": {decodeOnly, `
a:
  # HM
  - # HB1
    # HB2
    b: # IB
      c # IC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Style:  0x0,
					Tag:    "!!str",
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   3,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.MappingNode,
						Tag:         "!!map",
						HeadComment: "# HM",
						Line:        5,
						Column:      5,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "b",
							HeadComment: "# HB1\n# HB2",
							LineComment: "# IB",
							Line:        5,
							Column:      5,
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "c",
							LineComment: "# IC",
							Line:        6,
							Column:      7,
						}},
					}},
				}},
			}},
		},
	},
	// When encoding the value above, it loses b's inline comment.
	"t29": {encodeOnly, `
a:
  # HM
  - # HB1
    # HB2
    b: c # IC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Style:  0x0,
					Tag:    "!!str",
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   3,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.MappingNode,
						Tag:         "!!map",
						HeadComment: "# HM",
						Line:        5,
						Column:      5,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "b",
							HeadComment: "# HB1\n# HB2",
							LineComment: "# IB",
							Line:        5,
							Column:      5,
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "c",
							LineComment: "# IC",
							Line:        6,
							Column:      7,
						}},
					}},
				}},
			}},
		},
	},
	// Multiple cases of comment inlining next to mapping keys.
	"t30": {doBoth, `
a: | # IA
  str
b: >- # IB
  str
c: # IC
  - str
d: # ID
  str:
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:        yaml.ScalarNode,
					Style:       yaml.LiteralStyle,
					Tag:         "!!str",
					Value:       "str\n",
					LineComment: "# IA",
					Line:        1,
					Column:      4,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "b",
					Line:   3,
					Column: 1,
				}, {
					Kind:        yaml.ScalarNode,
					Style:       yaml.FoldedStyle,
					Tag:         "!!str",
					Value:       "str",
					LineComment: "# IB",
					Line:        3,
					Column:      4,
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "c",
					LineComment: "# IC",
					Line:        5,
					Column:      1,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   6,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "str",
						Line:   6,
						Column: 5,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "d",
					LineComment: "# ID",
					Line:        7,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   8,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "str",
						Line:   8,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!null",
						Line:   8,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Indentless sequence.
	"t31": {decodeOnly, `
a:
# HM
- # HB1
  # HB2
  b: # IB
    c # IC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   3,
					Column: 1,
					Content: []*yaml.Node{{
						Kind:        yaml.MappingNode,
						Tag:         "!!map",
						HeadComment: "# HM",
						Line:        5,
						Column:      3,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "b",
							HeadComment: "# HB1\n# HB2",
							LineComment: "# IB",
							Line:        5,
							Column:      3,
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "c",
							LineComment: "# IC",
							Line:        6,
							Column:      5,
						}},
					}},
				}},
			}},
		},
	}, "t32": {doBoth, `
- a
- b
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Value:  "",
				Tag:    "!!seq",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b",
					Tag:    "!!str",
					Line:   2,
					Column: 3,
				}},
			}},
		},
	}, "t33": {doBoth, `
- a
- - b
  - c
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Tag:    "!!str",
						Line:   2,
						Column: 5,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Tag:    "!!str",
						Line:   3,
						Column: 5,
					}},
				}},
			}},
		},
	}, "t34": {doBoth, `
[a, b]
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Value:  "",
				Tag:    "!!seq",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 2,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b",
					Tag:    "!!str",
					Line:   1,
					Column: 5,
				}},
			}},
		},
	}, "t35": {doBoth, `
- a
- [b, c]
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.SequenceNode,
					Tag:    "!!seq",
					Style:  yaml.FlowStyle,
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Tag:    "!!str",
						Line:   2,
						Column: 4,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Tag:    "!!str",
						Line:   2,
						Column: 7,
					}},
				}},
			}},
		},
	}, "t36": {doBoth, `
a: &x 1
b: &y 2
c: *x
d: *y
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Line:   1,
				Column: 1,
				Tag:    "!!map",
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Tag:    "!!str",
					Line:   1,
					Column: 1,
				},
					saveNode("x", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Value:  "1",
						Tag:    "!!int",
						Anchor: "x",
						Line:   1,
						Column: 4,
					}),
					{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Tag:    "!!str",
						Line:   2,
						Column: 1,
					},
					saveNode("y", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Value:  "2",
						Tag:    "!!int",
						Anchor: "y",
						Line:   2,
						Column: 4,
					}),
					{
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Tag:    "!!str",
						Line:   3,
						Column: 1,
					}, {
						Kind:   yaml.AliasNode,
						Value:  "x",
						Alias:  dropNode("x"),
						Line:   3,
						Column: 4,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "d",
						Tag:    "!!str",
						Line:   4,
						Column: 1,
					}, {
						Kind:   yaml.AliasNode,
						Value:  "y",
						Tag:    "",
						Alias:  dropNode("y"),
						Line:   4,
						Column: 4,
					}},
			}},
		},
	}, "t37": {doBoth, `
# One
# Two
true # Three
# Four
# Five
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   3,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        3,
				Column:      1,
				HeadComment: "# One\n# Two",
				LineComment: "# Three",
				FootComment: "# Four\n# Five",
			}},
		},
	}, "t38": {doBoth, `
# š
true # š
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        2,
				Column:      1,
				HeadComment: "# š",
				LineComment: "# š",
			}},
		},
	}, "t39": {decodeOnly, `

# One

# Two

# Three
true # Four
# Five

# Six

# Seven
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        7,
			Column:      1,
			HeadComment: "# One\n\n# Two",
			FootComment: "# Six\n\n# Seven",
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        7,
				Column:      1,
				HeadComment: "# Three",
				LineComment: "# Four",
				FootComment: "# Five",
			}},
		},
	},
	// Write out the pound character if missing from comments.
	"t40": {encodeOnly, `
# One
# Two
true # Three
# Four
# Five
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   3,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        3,
				Column:      1,
				HeadComment: "One\nTwo\n",
				LineComment: "Three\n",
				FootComment: "Four\nFive\n",
			}},
		},
	}, "t41": {encodeOnly, `
#   One
#   Two
true #   Three
#   Four
#   Five
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   3,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        3,
				Column:      1,
				HeadComment: "  One\n  Two",
				LineComment: "  Three",
				FootComment: "  Four\n  Five",
			}},
		},
	}, "t42": {doBoth, `
# DH1

# DH2

# H1
# H2
true # I
# F1
# F2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        7,
			Column:      1,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Content: []*yaml.Node{{
				Kind:        yaml.ScalarNode,
				Value:       "true",
				Tag:         "!!bool",
				Line:        7,
				Column:      1,
				HeadComment: "# H1\n# H2",
				LineComment: "# I",
				FootComment: "# F1\n# F2",
			}},
		},
	}, "t43": {doBoth, `
# DH1

# DH2

# HA1
# HA2
ka: va # IA
# FA1
# FA2

# HB1
# HB2
kb: vb # IB
# FB1
# FB2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        7,
			Column:      1,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   7,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Line:        7,
					Column:      1,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1\n# HA2",
					FootComment: "# FA1\n# FA2",
				}, {
					Kind:        yaml.ScalarNode,
					Line:        7,
					Column:      5,
					Tag:         "!!str",
					Value:       "va",
					LineComment: "# IA",
				}, {
					Kind:        yaml.ScalarNode,
					Line:        13,
					Column:      1,
					Tag:         "!!str",
					Value:       "kb",
					HeadComment: "# HB1\n# HB2",
					FootComment: "# FB1\n# FB2",
				}, {
					Kind:        yaml.ScalarNode,
					Line:        13,
					Column:      5,
					Tag:         "!!str",
					Value:       "vb",
					LineComment: "# IB",
				}},
			}},
		},
	}, "t44": {doBoth, `
# DH1

# DH2

# HA1
# HA2
- la # IA
# FA1
# FA2

# HB1
# HB2
- lb # IB
# FB1
# FB2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        7,
			Column:      1,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   7,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        7,
					Column:      3,
					Value:       "la",
					HeadComment: "# HA1\n# HA2",
					LineComment: "# IA",
					FootComment: "# FA1\n# FA2",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        13,
					Column:      3,
					Value:       "lb",
					HeadComment: "# HB1\n# HB2",
					LineComment: "# IB",
					FootComment: "# FB1\n# FB2",
				}},
			}},
		},
	}, "t45": {doBoth, `
# DH1

- la # IA
# HB1
- lb
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        3,
			Column:      1,
			HeadComment: "# DH1",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   3,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        3,
					Column:      3,
					Value:       "la",
					LineComment: "# IA",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        5,
					Column:      3,
					Value:       "lb",
					HeadComment: "# HB1",
				}},
			}},
		},
	}, "t46": {doBoth, `
- la # IA
- lb # IB
- lc # IC
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        1,
					Column:      3,
					Value:       "la",
					LineComment: "# IA",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        2,
					Column:      3,
					Value:       "lb",
					LineComment: "# IB",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        3,
					Column:      3,
					Value:       "lc",
					LineComment: "# IC",
				}},
			}},
		},
	}, "t47": {doBoth, `
# DH1

# HL1
- - la
  # HB1
  - lb
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   4,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.SequenceNode,
					Tag:         "!!seq",
					Line:        4,
					Column:      3,
					HeadComment: "# HL1",
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Line:   4,
						Column: 5,
						Value:  "la",
					}, {
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        6,
						Column:      5,
						Value:       "lb",
						HeadComment: "# HB1",
					}},
				}},
			}},
		},
	}, "t48": {doBoth, `
# DH1

# HL1
- # HA1
  - la
  # HB1
  - lb
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   4,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.SequenceNode,
					Tag:         "!!seq",
					Line:        5,
					Column:      3,
					HeadComment: "# HL1",
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        5,
						Column:      5,
						Value:       "la",
						HeadComment: "# HA1",
					}, {
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        7,
						Column:      5,
						Value:       "lb",
						HeadComment: "# HB1",
					}},
				}},
			}},
		},
	}, "t49": {decodeOnly, `
# DH1

# HL1
- # HA1

  - la
  # HB1
  - lb
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   4,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.SequenceNode,
					Tag:         "!!seq",
					Line:        6,
					Column:      3,
					HeadComment: "# HL1",
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        6,
						Column:      5,
						Value:       "la",
						HeadComment: "# HA1\n",
					}, {
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        8,
						Column:      5,
						Value:       "lb",
						HeadComment: "# HB1",
					}},
				}},
			}},
		},
	}, "t50": {doBoth, `
# DH1

# HA1
ka:
  # HB1
  kb:
    # HC1
    # HC2
    - lc # IC
    # FC1
    # FC2

    # HD1
    - ld # ID
    # FD1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   4,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        4,
					Column:      1,
					Value:       "ka",
					HeadComment: "# HA1",
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   6,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        6,
						Column:      3,
						Value:       "kb",
						HeadComment: "# HB1",
					}, {
						Kind:   yaml.SequenceNode,
						Line:   9,
						Column: 5,
						Tag:    "!!seq",
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Line:        9,
							Column:      7,
							Value:       "lc",
							HeadComment: "# HC1\n# HC2",
							LineComment: "# IC",
							FootComment: "# FC1\n# FC2",
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Line:        14,
							Column:      7,
							Value:       "ld",
							HeadComment: "# HD1",

							LineComment: "# ID",
							FootComment: "# FD1",
						}},
					}},
				}},
			}},
		},
	}, "t51": {doBoth, `
# DH1

# HA1
ka:
  # HB1
  kb:
    # HC1
    # HC2
    - lc # IC
    # FC1
    # FC2

    # HD1
    - ld # ID
    # FD1
ke: ve

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   4,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        4,
					Column:      1,
					Value:       "ka",
					HeadComment: "# HA1",
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   6,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        6,
						Column:      3,
						Value:       "kb",
						HeadComment: "# HB1",
					}, {
						Kind:   yaml.SequenceNode,
						Line:   9,
						Column: 5,
						Tag:    "!!seq",
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Line:        9,
							Column:      7,
							Value:       "lc",
							HeadComment: "# HC1\n# HC2",
							LineComment: "# IC",
							FootComment: "# FC1\n# FC2",
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Line:        14,
							Column:      7,
							Value:       "ld",
							HeadComment: "# HD1",
							LineComment: "# ID",
							FootComment: "# FD1",
						}},
					}},
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   16,
					Column: 1,
					Value:  "ke",
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   16,
					Column: 5,
					Value:  "ve",
				}},
			}},
		},
	}, "t52": {doBoth, `
# DH1

# DH2

# HA1
# HA2
ka:
  # HB1
  # HB2
  kb:
    # HC1
    # HC2
    kc:
      # HD1
      # HD2
      kd: vd
      # FD1
      # FD2
    # FC1
    # FC2
  # FB1
  # FB2
# FA1
# FA2

# HE1
# HE2
ke: ve
# FE1
# FE2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Line:        7,
			Column:      1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   7,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1\n# HA2",
					FootComment: "# FA1\n# FA2",
					Line:        7,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   10,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1\n# HB2",
						FootComment: "# FB1\n# FB2",
						Line:        10,
						Column:      3,
					}, {
						Kind:   yaml.MappingNode,
						Tag:    "!!map",
						Line:   13,
						Column: 5,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "kc",
							HeadComment: "# HC1\n# HC2",
							FootComment: "# FC1\n# FC2",
							Line:        13,
							Column:      5,
						}, {
							Kind:   yaml.MappingNode,
							Tag:    "!!map",
							Line:   16,
							Column: 7,
							Content: []*yaml.Node{{
								Kind:        yaml.ScalarNode,
								Tag:         "!!str",
								Value:       "kd",
								HeadComment: "# HD1\n# HD2",
								FootComment: "# FD1\n# FD2",
								Line:        16,
								Column:      7,
							}, {
								Kind:   yaml.ScalarNode,
								Tag:    "!!str",
								Value:  "vd",
								Line:   16,
								Column: 11,
							}},
						}},
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ke",
					HeadComment: "# HE1\n# HE2",
					FootComment: "# FE1\n# FE2",
					Line:        28,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "ve",
					Line:   28,
					Column: 5,
				}},
			}},
		},
	},
	// Same as above but indenting ke in so it's also part of ka's value.
	"t53": {doBoth, `
# DH1

# DH2

# HA1
# HA2
ka:
  # HB1
  # HB2
  kb:
    # HC1
    # HC2
    kc:
      # HD1
      # HD2
      kd: vd
      # FD1
      # FD2
    # FC1
    # FC2
  # FB1
  # FB2

  # HE1
  # HE2
  ke: ve
  # FE1
  # FE2
# FA1
# FA2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Line:        7,
			Column:      1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   7,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1\n# HA2",
					FootComment: "# FA1\n# FA2",
					Line:        7,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   10,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1\n# HB2",
						FootComment: "# FB1\n# FB2",
						Line:        10,
						Column:      3,
					}, {
						Kind:   yaml.MappingNode,
						Tag:    "!!map",
						Line:   13,
						Column: 5,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "kc",
							HeadComment: "# HC1\n# HC2",
							FootComment: "# FC1\n# FC2",
							Line:        13,
							Column:      5,
						}, {
							Kind:   yaml.MappingNode,
							Tag:    "!!map",
							Line:   16,
							Column: 7,
							Content: []*yaml.Node{{
								Kind:        yaml.ScalarNode,
								Tag:         "!!str",
								Value:       "kd",
								HeadComment: "# HD1\n# HD2",
								FootComment: "# FD1\n# FD2",
								Line:        16,
								Column:      7,
							}, {
								Kind:   yaml.ScalarNode,
								Tag:    "!!str",
								Value:  "vd",
								Line:   16,
								Column: 11,
							}},
						}},
					}, {
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "ke",
						HeadComment: "# HE1\n# HE2",
						FootComment: "# FE1\n# FE2",
						Line:        26,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "ve",
						Line:   26,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Decode only due to lack of newline at the end.
	"t54": {decodeOnly, `
# HA1
ka:
  # HB1
  kb: vb
  # FB1
# FA1`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						FootComment: "# FB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Same as above, but with newline at the end.
	"t55": {doBoth, `
# HA1
ka:
  # HB1
  kb: vb
  # FB1
# FA1
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						FootComment: "# FB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Same as above, but without FB1.
	"t56": {doBoth, `
# HA1
ka:
  # HB1
  kb: vb
# FA1
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Same as above, but with two newlines at the end.
	"t57": {decodeOnly, `
# HA1
ka:
  # HB1
  kb: vb
  # FB1
# FA1

`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						FootComment: "# FB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}},
			}},
		},
	},
	// Similar to above, but make HB1 look more like a footer of ka.
	"t58": {decodeOnly, `
# HA1
ka:
# HB1

  kb: vb
# FA1
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   5,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1\n",
						Line:        5,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   5,
						Column: 7,
					}},
				}},
			}},
		},
	}, "t59": {doBoth, `
ka:
  kb: vb
# FA1

kc: vc
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					Line:        1,
					Column:      1,
					FootComment: "# FA1",
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "kb",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   2,
						Column: 7,
					}},
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "kc",
					Line:   5,
					Column: 1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   5,
					Column: 5,
				}},
			}},
		},
	}, "t60": {doBoth, `
ka:
  kb: vb
# HC1
kc: vc
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "ka",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "kb",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   2,
						Column: 7,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "kc",
					HeadComment: "# HC1",
					Line:        4,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   4,
					Column: 5,
				}},
			}},
		},
	},
	// Decode only due to empty line before HC1.
	"t61": {decodeOnly, `
ka:
  kb: vb

# HC1
kc: vc
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "ka",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "kb",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   2,
						Column: 7,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "kc",
					HeadComment: "# HC1",
					Line:        5,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   5,
					Column: 5,
				}},
			}},
		},
	},
	// Decode only due to empty lines around HC1.
	"t62": {decodeOnly, `
ka:
  kb: vb

# HC1

kc: vc
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "ka",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "kb",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   2,
						Column: 7,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "kc",
					HeadComment: "# HC1\n",
					Line:        6,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   6,
					Column: 5,
				}},
			}},
		},
	}, "t63": {doBoth, `
ka: # IA
  kb: # IB
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					Line:        1,
					Column:      1,
					LineComment: "# IA",
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						Line:        2,
						Column:      3,
						LineComment: "# IB",
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!null",
						Line:   2,
						Column: 6,
					}},
				}},
			}},
		},
	}, "t64": {doBoth, `
# HA1
ka:
  # HB1
  kb: vb
  # FB1
# HC1
# HC2
kc: vc
# FC1
# FC2
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						FootComment: "# FB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "kc",
					HeadComment: "# HC1\n# HC2",
					FootComment: "# FC1\n# FC2",
					Line:        8,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   8,
					Column: 5,
				}},
			}},
		},
	},
	// Same as above, but decode only due to empty line between
	// ka's value and kc's headers.
	"t65": {decodeOnly, `
# HA1
ka:
  # HB1
  kb: vb
  # FB1

# HC1
# HC2
kc: vc
# FC1
# FC2
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   2,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "ka",
					HeadComment: "# HA1",
					Line:        2,
					Column:      1,
				}, {
					Kind:   yaml.MappingNode,
					Tag:    "!!map",
					Line:   4,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Value:       "kb",
						HeadComment: "# HB1",
						FootComment: "# FB1",
						Line:        4,
						Column:      3,
					}, {
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "vb",
						Line:   4,
						Column: 7,
					}},
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Value:       "kc",
					HeadComment: "# HC1\n# HC2",
					FootComment: "# FC1\n# FC2",
					Line:        9,
					Column:      1,
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "vc",
					Line:   9,
					Column: 5,
				}},
			}},
		},
	}, "t66": {doBoth, `
# H1
[la, lb] # I
# F1
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   2,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:        yaml.SequenceNode,
				Tag:         "!!seq",
				Style:       yaml.FlowStyle,
				Line:        2,
				Column:      1,
				HeadComment: "# H1",
				LineComment: "# I",
				FootComment: "# F1",
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   2,
					Column: 2,
					Value:  "la",
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   2,
					Column: 6,
					Value:  "lb",
				}},
			}},
		},
	}, "t67": {doBoth, `
# DH1

# SH1
[
  # HA1
  la, # IA
  # FA1

  # HB1
  lb, # IB
  # FB1
]
# SF1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:        yaml.SequenceNode,
				Tag:         "!!seq",
				Style:       yaml.FlowStyle,
				Line:        4,
				Column:      1,
				HeadComment: "# SH1",
				FootComment: "# SF1",
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      3,
					Value:       "la",
					HeadComment: "# HA1",
					LineComment: "# IA",
					FootComment: "# FA1",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      3,
					Value:       "lb",
					HeadComment: "# HB1",
					LineComment: "# IB",
					FootComment: "# FB1",
				}},
			}},
		},
	},
	// Same as above, but with extra newlines before FB1 and FB2
	"t68": {decodeOnly, `
# DH1

# SH1
[
  # HA1
  la, # IA
  # FA1

  # HB1
  lb, # IB


  # FB1

# FB2
]
# SF1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:        yaml.SequenceNode,
				Tag:         "!!seq",
				Style:       yaml.FlowStyle,
				Line:        4,
				Column:      1,
				HeadComment: "# SH1",
				FootComment: "# SF1",
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      3,
					Value:       "la",
					HeadComment: "# HA1",
					LineComment: "# IA",
					FootComment: "# FA1",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      3,
					Value:       "lb",
					HeadComment: "# HB1",
					LineComment: "# IB",
					FootComment: "# FB1\n\n# FB2",
				}},
			}},
		},
	}, "t69": {doBoth, `
# DH1

# SH1
[
  # HA1
  la,
  # FA1

  # HB1
  lb,
  # FB1
]
# SF1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:        yaml.SequenceNode,
				Tag:         "!!seq",
				Style:       yaml.FlowStyle,
				Line:        4,
				Column:      1,
				HeadComment: "# SH1",
				FootComment: "# SF1",
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      3,
					Value:       "la",
					HeadComment: "# HA1",
					FootComment: "# FA1",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      3,
					Value:       "lb",
					HeadComment: "# HB1",
					FootComment: "# FB1",
				}},
			}},
		},
	}, "t70": {doBoth, `
ka:
  kb: [
    # HA1
    la,
    # FA1

    # HB1
    lb,
    # FB1
  ]
`,
		yaml.Node{
			Kind:   yaml.DocumentNode,
			Line:   1,
			Column: 1,
			Content: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Tag:    "!!map",
				Line:   1,
				Column: 1,
				Content: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Value:  "ka",
					Line:   1,
					Column: 1,
				}, {
					Kind:   0x4,
					Tag:    "!!map",
					Line:   2,
					Column: 3,
					Content: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Tag:    "!!str",
						Value:  "kb",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.SequenceNode,
						Style:  0x20,
						Tag:    "!!seq",
						Line:   2,
						Column: 7,
						Content: []*yaml.Node{{
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "la",
							HeadComment: "# HA1",
							FootComment: "# FA1",
							Line:        4,
							Column:      5,
						}, {
							Kind:        yaml.ScalarNode,
							Tag:         "!!str",
							Value:       "lb",
							HeadComment: "# HB1",
							FootComment: "# FB1",
							Line:        8,
							Column:      5,
						}},
					}},
				}},
			}},
		},
	}, "t71": {doBoth, `
# DH1

# MH1
{
  # HA1
  ka: va, # IA
  # FA1

  # HB1
  kb: vb, # IB
  # FB1
}
# MF1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:        yaml.MappingNode,
				Tag:         "!!map",
				Style:       yaml.FlowStyle,
				Line:        4,
				Column:      1,
				HeadComment: "# MH1",
				FootComment: "# MF1",
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      3,
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      7,
					Value:       "va",
					LineComment: "# IA",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      3,
					Value:       "kb",
					HeadComment: "# HB1",
					FootComment: "# FB1",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      7,
					Value:       "vb",
					LineComment: "# IB",
				}},
			}},
		},
	}, "t72": {doBoth, `
# DH1

# MH1
{
  # HA1
  ka: va,
  # FA1

  # HB1
  kb: vb,
  # FB1
}
# MF1

# DF1
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        4,
			Column:      1,
			HeadComment: "# DH1",
			FootComment: "# DF1",
			Content: []*yaml.Node{{
				Kind:        yaml.MappingNode,
				Tag:         "!!map",
				Style:       yaml.FlowStyle,
				Line:        4,
				Column:      1,
				HeadComment: "# MH1",
				FootComment: "# MF1",
				Content: []*yaml.Node{{
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        6,
					Column:      3,
					Value:       "ka",
					HeadComment: "# HA1",
					FootComment: "# FA1",
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   6,
					Column: 7,
					Value:  "va",
				}, {
					Kind:        yaml.ScalarNode,
					Tag:         "!!str",
					Line:        10,
					Column:      3,
					Value:       "kb",
					HeadComment: "# HB1",
					FootComment: "# FB1",
				}, {
					Kind:   yaml.ScalarNode,
					Tag:    "!!str",
					Line:   10,
					Column: 7,
					Value:  "vb",
				}},
			}},
		},
	}, "t73": {doBoth, `
# DH1

# DH2

# HA1
# HA2
- &x la # IA
# FA1
# FA2

# HB1
# HB2
- *x # IB
# FB1
# FB2

# DF1

# DF2
`,
		yaml.Node{
			Kind:        yaml.DocumentNode,
			Line:        7,
			Column:      1,
			HeadComment: "# DH1\n\n# DH2",
			FootComment: "# DF1\n\n# DF2",
			Content: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Tag:    "!!seq",
				Line:   7,
				Column: 1,
				Content: []*yaml.Node{
					saveNode("x", &yaml.Node{
						Kind:        yaml.ScalarNode,
						Tag:         "!!str",
						Line:        7,
						Column:      3,
						Value:       "la",
						HeadComment: "# HA1\n# HA2",
						LineComment: "# IA",
						FootComment: "# FA1\n# FA2",
						Anchor:      "x",
					}), {
						Kind:        yaml.AliasNode,
						Line:        13,
						Column:      3,
						Value:       "x",
						Alias:       dropNode("x"),
						HeadComment: "# HB1\n# HB2",
						LineComment: "# IB",
						FootComment: "# FB1\n# FB2",
					},
				},
			}},
		},
	},
}

const chattyDebugging = false

func (s *S) TestNodeRoundtrip(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range nodeTests {
		theYaml := item.yaml[1:] // Strip leading linefeed used in formatting.
		c.Logf("%s: %v %q", i, item.do, theYaml)
		if chattyDebugging && strings.Contains(theYaml, "#") {
			var buf bytes.Buffer
			fprintComments(&buf, &item.node, "    ")
			c.Logf("  expected comments:\n%s", buf.Bytes())
		}

		if item.do == decodeOnly || item.do == doBoth {
			var node yaml.Node
			c.Assert(yaml.Unmarshal([]byte(theYaml), &node), IsNil)
			if chattyDebugging && strings.Contains(theYaml, "#") {
				var buf bytes.Buffer
				fprintComments(&buf, &node, "    ")
				c.Logf("  obtained comments:\n%s", buf.Bytes())
			}
			if !c.Check(
				&node, DeepEquals, &item.node,
				Commentf("obtained node tree doesn't match expected node tree")) {
				var buf bytes.Buffer
				dumpNode(&buf, "obtained", &node)
				dumpNode(&buf, "expected", &item.node)
				c.Logf("Decoding failure:\n%s", buf.Bytes())
			}
		}
		if item.do == encodeOnly || item.do == doBoth {
			node := deepCopyNode(&item.node, nil)
			buf := bytes.Buffer{}
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(2)
			c.Assert(enc.Encode(node), IsNil)
			c.Assert(enc.Close(), IsNil)
			c.Assert(buf.String(), Equals, theYaml)
			c.Assert(
				node, DeepEquals, &item.node,
				Commentf("unexpected change in node tree"))
		}
	}
}

func deepCopyNode(node *yaml.Node, cache map[*yaml.Node]*yaml.Node) *yaml.Node {
	if n, ok := cache[node]; ok {
		return n
	}
	if cache == nil {
		cache = make(map[*yaml.Node]*yaml.Node)
	}
	copy := *node
	cache[node] = &copy
	copy.Content = nil
	for _, elem := range node.Content {
		copy.Content = append(copy.Content, deepCopyNode(elem, cache))
	}
	if node.Alias != nil {
		copy.Alias = deepCopyNode(node.Alias, cache)
	}
	return &copy
}

var savedNodes = make(map[string]*yaml.Node)

func saveNode(name string, node *yaml.Node) *yaml.Node {
	savedNodes[name] = node
	return node
}

func peekNode(name string) *yaml.Node {
	return savedNodes[name]
}

func dropNode(name string) *yaml.Node {
	node := savedNodes[name]
	delete(savedNodes, name)
	return node
}

var setStringTests = []struct {
	str  string
	yaml string
	node yaml.Node
}{
	{
		"something simple",
		"something simple\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "something simple",
			Tag:   "!!str",
		},
	}, {
		`"quoted value"`,
		"'\"quoted value\"'\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: `"quoted value"`,
			Tag:   "!!str",
		},
	}, {
		"multi\nline",
		"|-\n  multi\n  line\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "multi\nline",
			Tag:   "!!str",
			Style: yaml.LiteralStyle,
		},
	}, {
		"123",
		"\"123\"\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "123",
			Tag:   "!!str",
		},
	}, {
		"multi\nline\n",
		"|\n  multi\n  line\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "multi\nline\n",
			Tag:   "!!str",
			Style: yaml.LiteralStyle,
		},
	}, {
		"\x80\x81\x82",
		"!!binary gIGC\n",
		yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "gIGC",
			Tag:   "!!binary",
		},
	},
}

func (s *S) TestSetString(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range setStringTests {
		c.Logf("test %d: %q", i, item.str)

		var node yaml.Node

		node.SetString(item.str)

		c.Assert(node, DeepEquals, item.node)

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err := enc.Encode(&item.node)
		c.Assert(err, IsNil)
		err = enc.Close()
		c.Assert(err, IsNil)
		c.Assert(buf.String(), Equals, item.yaml)

		var doc yaml.Node
		err = yaml.Unmarshal([]byte(item.yaml), &doc)
		c.Assert(err, IsNil)

		var str string
		err = node.Decode(&str)
		c.Assert(err, IsNil)
		c.Assert(str, Equals, item.str)
	}
}

var nodeEncodeDecodeTests = []struct {
	value interface{}
	yaml  string
	node  yaml.Node
}{{
	"something simple",
	"something simple\n",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "something simple",
		Tag:   "!!str",
	},
}, {
	`"quoted value"`,
	"'\"quoted value\"'\n",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.SingleQuotedStyle,
		Value: `"quoted value"`,
		Tag:   "!!str",
	},
}, {
	123,
	"123",
	yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: `123`,
		Tag:   "!!int",
	},
}, {
	[]interface{}{1, 2},
	"[1, 2]",
	yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
		Content: []*yaml.Node{{
			Kind:  yaml.ScalarNode,
			Value: "1",
			Tag:   "!!int",
		}, {
			Kind:  yaml.ScalarNode,
			Value: "2",
			Tag:   "!!int",
		}},
	},
}, {
	map[string]interface{}{"a": "b"},
	"a: b",
	yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{{
			Kind:  yaml.ScalarNode,
			Value: "a",
			Tag:   "!!str",
		}, {
			Kind:  yaml.ScalarNode,
			Value: "b",
			Tag:   "!!str",
		}},
	},
}}

func (s *S) TestNodeEncodeDecode(c *C) {
	for i, item := range nodeEncodeDecodeTests {
		c.Logf("Encode/Decode test value #%d: %#v", i, item.value)

		var v interface{}
		err := item.node.Decode(&v)
		c.Assert(err, IsNil)
		c.Assert(v, DeepEquals, item.value)

		var n yaml.Node
		err = n.Encode(item.value)
		c.Assert(err, IsNil)
		c.Assert(n, DeepEquals, item.node)
	}
}

func (s *S) TestNodeZeroEncodeDecode(c *C) {
	// Zero node value behaves as nil when encoding...
	var n yaml.Node
	data, err := yaml.Marshal(&n)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "null\n")

	// ... and decoding.
	var v *struct{} = &struct{}{}
	c.Assert(n.Decode(&v), IsNil)
	c.Assert(v, IsNil)

	// ... and even when looking for its tag.
	c.Assert(n.ShortTag(), Equals, "!!null")

	// Kind zero is still unknown, though.
	n.Line = 1
	_, err = yaml.Marshal(&n)
	c.Assert(err, ErrorMatches, "yaml: cannot encode node with unknown kind 0")
	c.Assert(n.Decode(&v), ErrorMatches, "yaml: cannot decode node with unknown kind 0")
}

func (s *S) TestNodeOmitEmpty(c *C) {
	var v struct {
		A int
		B yaml.Node ",omitempty"
	}
	v.A = 1
	data, err := yaml.Marshal(&v)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "a: 1\n")

	v.B.Line = 1
	_, err = yaml.Marshal(&v)
	c.Assert(err, ErrorMatches, "yaml: cannot encode node with unknown kind 0")
}

func dumpNode(out io.Writer, title string, n *yaml.Node) {
	fmt.Fprintf(out, "---- %s -------------------\n", title)
	dumpDetails(out, n, "")
}

func dumpDetails(out io.Writer, n *yaml.Node, indent string) {
	fmt.Fprintf(out, "ln=%3d col=%3d ", n.Line, n.Column)
	fmt.Fprint(out, indent)
	if n.Tag != "" {
		fmt.Fprintf(out, "%v ", n.Tag)
	}
	if n.Value != "" {
		fmt.Fprintf(out, "%q ", n.Value)
	}
	if n.Kind > 0 {
		fmt.Fprintf(out, "kind=%v ", n.Kind)
	}
	if n.Style > 0 {
		fmt.Fprintf(out, "style=%v ", n.Style)
	}
	if n.HeadComment != "" {
		fmt.Fprintf(out, "hCom: %q ", n.HeadComment)
	}
	if n.LineComment != "" {
		fmt.Fprintf(out, "lCom: %q ", n.LineComment)
	}
	if n.FootComment != "" {
		fmt.Fprintf(out, "fCom: %q ", n.FootComment)
	}
	fmt.Fprintln(out)
	for _, child := range n.Content {
		dumpDetails(out, child, indent+"  ")
	}
}

func fprintComments(out io.Writer, node *yaml.Node, indent string) {
	switch node.Kind {
	case yaml.ScalarNode:
		fmt.Fprintf(out, "%s<%s> ", indent, node.Value)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
	case yaml.DocumentNode:
		fmt.Fprintf(out, "%s<DOC> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i++ {
			fprintComments(out, node.Content[i], indent+"  ")
		}
	case yaml.MappingNode:
		fmt.Fprintf(out, "%s<MAP> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i += 2 {
			fprintComments(out, node.Content[i], indent+"  ")
			fprintComments(out, node.Content[i+1], indent+"  ")
		}
	case yaml.SequenceNode:
		fmt.Fprintf(out, "%s<SEQ> ", indent)
		fprintCommentSet(out, node)
		fmt.Fprintf(out, "\n")
		for i := 0; i < len(node.Content); i++ {
			fprintComments(out, node.Content[i], indent+"  ")
		}
	}
}

func fprintCommentSet(out io.Writer, node *yaml.Node) {
	if len(node.HeadComment)+len(node.LineComment)+len(node.FootComment) > 0 {
		fmt.Fprintf(out, "%q / %q / %q", node.HeadComment, node.LineComment, node.FootComment)
	}
}
