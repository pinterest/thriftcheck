// Copyright 2025 Pinterest
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

package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

var TypeToTypeCheckerFunc = map[string]func(n ast.Node) bool{
	"union": isUnionType,
}

func isUnionType(n ast.Node) bool {
	s, ok := n.(*ast.Struct)

	return ok && s.Type == ast.UnionType
}

// CheckTypesDisallowed reports an error if a disallowed type is used.
func CheckTypesDisallowed(disallowedTypes []string) thriftcheck.Check {
	return thriftcheck.NewCheck("types.disallowed", func(c *thriftcheck.C, n ast.Node) {
		for _, t := range disallowedTypes {
			if TypeToTypeCheckerFunc[t](n) {
				c.Errorf(n, "a disallowed type (%s) was used", t)
				break
			}
		}
	})
}
