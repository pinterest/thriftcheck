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

// CheckUnion reports an error if a union is used.
func CheckUnion() *thriftcheck.Check {
	return thriftcheck.NewCheck("union", func(c *thriftcheck.C, s *ast.Struct) {
		if s.Type == ast.UnionType {
			c.Errorf(s, "unions aren't allowed")
		}
	})
}
