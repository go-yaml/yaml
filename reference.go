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

// Package yaml implements YAML support for the Go language.
//
// Source code and other details for the project are available at GitHub:
//
//	https://github.com/go-yaml/yaml
package yaml

import (
	"regexp"
	"strings"
)

// Range is a programming construct used to represent a contiguous range of values.
type Range struct {
	Begin int
	End   int
}

// Reference is an object that consists of a Name string and an array of
// Range objects. It allows for organizing and referencing multiple
// contiguous value ranges associated with a specific name or identifier.
type Reference struct {
	Name  string
	Range Range
}

// References to represent a group of References with the same Name and Value,
// pointing to the same Node
type References struct {
	Name       string
	Target     *Node
	References []*Reference
}

// ReferenceReverse is an object containing Range and References, which helps
// to identify and replace a range with the corresponding references
type ReferenceReverse struct {
	Range      *Range
	References *References
}

var yamlRegexReference = regexp.MustCompile(`\${(.+?)}`)

func parseReferences(value string) []*Reference {
	items := make([]*Reference, 0)
	if yamlRegexReference.MatchString(value) {
		for _, x := range yamlRegexReference.FindAllStringSubmatchIndex(value, -1) {
			item := &Reference{
				Name: strings.TrimSpace(value[x[2]:x[3]]),
				Range: Range{
					Begin: x[0],
					End:   x[1],
				},
			}
			items = append(items, item)
		}
	}
	return items
}
