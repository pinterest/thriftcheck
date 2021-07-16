// Copyright 2021 Pinterest
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

package thriftcheck

import (
	"regexp"
	"strings"

	"go.uber.org/thriftrw/ast"
)

var nolintRegexp = regexp.MustCompile(`@nolint(?:\((.*)\))?`)

func nolint(n ast.Node) ([]string, bool) {
	var names []string

	if annotations := Annotations(n); annotations != nil {
		for _, annotation := range annotations {
			if annotation.Name == "nolint" {
				if annotation.Value == "" {
					return nil, true
				}
				names = append(names, splitTrim(annotation.Value, ",")...)
			}
		}
	}

	if doc := Doc(n); doc != "" {
		if m := nolintRegexp.FindStringSubmatch(doc); len(m) == 2 {
			if m[1] == "" {
				return nil, true
			}
			names = append(names, splitTrim(m[1], ",")...)
		}
	}

	if names == nil {
		return nil, false
	}

	// This may contain duplicate names at this point, but given how this
	// return value is used, we can tolerate that without doing the extra
	// work here to de-duplicate.
	return names, true
}

func splitTrim(s, sep string) []string {
	values := strings.Split(s, sep)
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return values
}
