package yaml_test

import (
	"os"

	. "gopkg.in/check.v1"
	"gopkg.in/niemeyer/ynext.v3"
)

var nodeTests = []struct {
	yaml string
	tag  string
	node yaml.Node
}{
	{
		"null\n",
		"!!null",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "null",
				Line:   1,
				Column: 1,
				Tag:    "",
			}},
		},
	}, {
		"foo\n",
		"!!str",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "foo",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"\"foo\"\n",
		"!!str",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.DoubleQuotedStyle,
				Value:  "foo",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"'foo'\n",
		"!!str",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.SingleQuotedStyle,
				Value:  "foo",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"''\n",
		"!!str",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.SingleQuotedStyle,
				Value:  "",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"|\n  foo\n  bar\n",
		"!!str",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Style:  yaml.LiteralStyle,
				Value:  "foo\nbar\n",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"true\n",
		"!!bool",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "true",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"-10\n",
		"!!int",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "-10",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"4294967296\n",
		"!!int",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "4294967296",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"0.1000\n",
		"!!float",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "0.1000",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"-.inf\n",
		"!!float",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "-.inf",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		".nan\n",
		"!!float",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  ".nan",
				Line:   1,
				Column: 1,
			}},
		},
	}, {
		"{}\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.FlowStyle,
				Value:  "",
				Line:   1,
				Column: 1,
				Tag:    "",
			}},
		},
	}, {
		"a: b c\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Value:  "",
				Line:   1,
				Column: 1,
				Tag:    "",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b c",
					Line:   1,
					Column: 4,
				}},
			}},
		},
	}, {
		"a:\n  b: c\n  d: e\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Line:   1,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 1,
				}, {
					Kind:   yaml.MappingNode,
					Line:   2,
					Column: 3,
					Children: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Line:   2,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Line:   2,
						Column: 6,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "d",
						Line:   3,
						Column: 3,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "e",
						Line:   3,
						Column: 6,
					}},
				}},
			}},
		},
	}, {
		"- a\n- b\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Value:  "",
				Line:   1,
				Column: 1,
				Tag:    "",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b",
					Line:   2,
					Column: 3,
				}},
			}},
		},
	}, {
		"- a\n- - b\n  - c\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Line:   1,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.SequenceNode,
					Line:   2,
					Column: 3,
					Children: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Line:   2,
						Column: 5,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Line:   3,
						Column: 5,
					}},
				}},
			}},
		},
	}, {
		"[a, b]\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Value:  "",
				Line:   1,
				Column: 1,
				Tag:    "",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 2,
				}, {
					Kind:   yaml.ScalarNode,
					Value:  "b",
					Line:   1,
					Column: 5,
				}},
			}},
		},
	}, {
		"- a\n- [b, c]\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    1,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Line:   1,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Value:  "a",
					Line:   1,
					Column: 3,
				}, {
					Kind:   yaml.SequenceNode,
					Style:  yaml.FlowStyle,
					Line:   2,
					Column: 3,
					Children: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Value:  "b",
						Line:   2,
						Column: 4,
					}, {
						Kind:   yaml.ScalarNode,
						Value:  "c",
						Line:   2,
						Column: 7,
					}},
				}},
			}},
		},
	}, {
		"# One\n# Two\ntrue # Three\n# Four\n# Five\n",
		"!!bool",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    3,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "true",
				Line:   3,
				Column: 1,
				Header: "# One\n# Two",
				Inline: "# Three",
				Footer: "# Four\n# Five",
			}},
		},
	}, {
		"# DH1\n\n# DH2\n\n# H1\n# H2\ntrue # I\n# F1\n# F2\n\n# DF1\n\n# DF2\n",
		"!!bool",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    7,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1\n\n# DH2",
			Footer:  "# DF1\n\n# DF2",
			Children: []*yaml.Node{{
				Kind:   yaml.ScalarNode,
				Value:  "true",
				Line:   7,
				Column: 1,
				Header: "# H1\n# H2",
				Inline: "# I",
				Footer: "# F1\n# F2",
			}},
		},
	}, {
		"# DH1\n\n# DH2\n\n# HA1\n# HA2\nka: va # IA\n# FA1\n# FA2\n\n# HB1\n# HB2\nkb: vb # IB\n# FB1\n# FB2\n\n# DF1\n\n# DF2\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    7,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1\n\n# DH2",
			Footer:  "# DF1\n\n# DF2",
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Line:   7,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   7,
					Column: 1,
					Value:  "ka",
					Header: "# HA1\n# HA2",
					Footer: "# FA1\n# FA2",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   7,
					Column: 5,
					Value:  "va",
					Inline: "# IA",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   13,
					Column: 1,
					Value:  "kb",
					Header: "# HB1\n# HB2",
					Footer: "# FB1\n# FB2",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   13,
					Column: 5,
					Value:  "vb",
					Inline: "# IB",
				}},
			}},
		},
	}, {
		"# DH1\n\n# DH2\n\n# HA1\n# HA2\n- la # IA\n# FA1\n# FA2\n\n# HB1\n# HB2\n- lb # IB\n# FB1\n# FB2\n\n# DF1\n\n# DF2\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    7,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1\n\n# DH2",
			Footer:  "# DF1\n\n# DF2",
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Line:   7,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   7,
					Column: 3,
					Value:  "la",
					Header: "# HA1\n# HA2",
					Inline: "# IA",
					Footer: "# FA1\n# FA2",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   13,
					Column: 3,
					Value:  "lb",
					Header: "# HB1\n# HB2",
					Inline: "# IB",
					Footer: "# FB1\n# FB2",
				}},
			}},
		},
	}, {
		"# DH1\n\n- la # IA\n\n# HB1\n- lb\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    3,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Line:   3,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   3,
					Column: 3,
					Value:  "la",
					Inline: "# IA",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 3,
					Value:  "lb",
					Header: "# HB1",
				}},
			}},
		},
	}, {
		"# DH1\n\n# HA1\nka:\n  # HB1\n  kb:\n  # HC1\n  # HC2\n  - lc # IC\n  # FC1\n  # FC2\n\n  # HD1\n  - ld # ID\n  # FD1\n\n# DF1\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    4,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Footer:  "# DF1",
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Line:   4,
				Column: 1,
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   4,
					Column: 1,
					Value:  "ka",
					Header: "# HA1",
				}, {
					Kind:   yaml.MappingNode,
					Line:   6,
					Column: 3,
					Children: []*yaml.Node{{
						Kind:   yaml.ScalarNode,
						Line:   6,
						Column: 3,
						Value:  "kb",
						Header: "# HB1",
					}, {
						Kind:   yaml.SequenceNode,
						Line:   9,
						Column: 3,
						Children: []*yaml.Node{{
							Kind:   yaml.ScalarNode,
							Line:   9,
							Column: 5,
							Value:  "lc",
							Header: "# HC1\n# HC2",
							Inline: "# IC",
							Footer: "# FC1\n# FC2",
						}, {
							Kind:   yaml.ScalarNode,
							Line:   14,
							Column: 5,
							Value:  "ld",
							Header: "# HD1",

							Inline: "# ID",
							Footer: "# FD1",
						}},
					}},
				}},
			}},
		},
	}, {
		"# H1\n[la, lb] # I\n# F1\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    2,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Line:   2,
				Column: 1,
				Header: "# H1",
				Inline: "# I",
				Footer: "# F1",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   2,
					Column: 2,
					Value:  "la",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   2,
					Column: 6,
					Value:  "lb",
				}},
			}},
		},
	}, {
		"# DH1\n\n# SH1\n[\n  # HA1\n  la, # IA\n  # FA1\n\n  # HB1\n  lb, # IB\n  # FB1\n]\n# SF1\n\n# DF1\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    4,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Footer:  "# DF1",
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Line:   4,
				Column: 1,
				Header: "# SH1",
				Footer: "# SF1",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 3,
					Value:  "la",
					Header: "# HA1",
					Inline: "# IA",
					Footer: "# FA1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 3,
					Value:  "lb",
					Header: "# HB1",
					Inline: "# IB",
					Footer: "# FB1",
				}},
			}},
		},
	}, {
		"# DH1\n\n# SH1\n[\n  # HA1\n  la,\n  # FA1\n\n  # HB1\n  lb,\n  # FB1\n]\n# SF1\n\n# DF1\n",
		"!!seq",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    4,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Footer:  "# DF1",
			Children: []*yaml.Node{{
				Kind:   yaml.SequenceNode,
				Style:  yaml.FlowStyle,
				Line:   4,
				Column: 1,
				Header: "# SH1",
				Footer: "# SF1",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 3,
					Value:  "la",
					Header: "# HA1",
					Footer: "# FA1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 3,
					Value:  "lb",
					Header: "# HB1",
					Footer: "# FB1",
				}},
			}},
		},
	}, {
		"# DH1\n\n# MH1\n{\n  # HA1\n  ka: va, # IA\n  # FA1\n\n  # HB1\n  kb: vb, # IB\n  # FB1\n}\n# MF1\n\n# DF1\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    4,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Footer:  "# DF1",
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.FlowStyle,
				Line:   4,
				Column: 1,
				Header: "# MH1",
				Footer: "# MF1",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 3,
					Value:  "ka",
					Header: "# HA1",
					Footer: "# FA1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 7,
					Value:  "va",
					Inline: "# IA",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 3,
					Value:  "kb",
					Header: "# HB1",
					Footer: "# FB1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 7,
					Value:  "vb",
					Inline: "# IB",
				}},
			}},
		},
	}, {
		"# DH1\n\n# MH1\n{\n  # HA1\n  ka: va,\n  # FA1\n\n  # HB1\n  kb: vb,\n  # FB1\n}\n# MF1\n\n# DF1\n",
		"!!map",
		yaml.Node{
			Kind:    yaml.DocumentNode,
			Line:    4,
			Column:  1,
			Anchors: map[string]*yaml.Node{},
			Header:  "# DH1",
			Footer:  "# DF1",
			Children: []*yaml.Node{{
				Kind:   yaml.MappingNode,
				Style:  yaml.FlowStyle,
				Line:   4,
				Column: 1,
				Header: "# MH1",
				Footer: "# MF1",
				Children: []*yaml.Node{{
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 3,
					Value:  "ka",
					Header: "# HA1",
					Footer: "# FA1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   6,
					Column: 7,
					Value:  "va",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 3,
					Value:  "kb",
					Header: "# HB1",
					Footer: "# FB1",
				}, {
					Kind:   yaml.ScalarNode,
					Line:   10,
					Column: 7,
					Value:  "vb",
				}},
			}},
		},
	},
}

func (s *S) TestNodeRoundtrip(c *C) {
	defer os.Setenv("TZ", os.Getenv("TZ"))
	os.Setenv("TZ", "UTC")
	for i, item := range nodeTests {
		c.Logf("test %d: %q", i, item.yaml)
		var node yaml.Node
		err := yaml.Unmarshal([]byte(item.yaml), &node)
		c.Assert(err, IsNil)
		c.Assert(node, DeepEquals, item.node)
		data, err := yaml.Marshal(&node)
		c.Assert(err, IsNil)
		c.Assert(string(data), Equals, item.yaml)
		if len(node.Children) > 0 {
			c.Assert(node.Children[0].ShortTag(), Equals, item.tag)
		}
	}
}
