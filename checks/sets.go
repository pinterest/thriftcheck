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

// CheckSetValueType returns a thriftcheck.Check that ensures that only primitive
// types are used for `set<>` values.
func CheckSetValueType() thriftcheck.Check {
	return thriftcheck.NewCheck("set.value.type", func(c *thriftcheck.C, st ast.SetType) {
		switch t := st.ValueType.(type) {
		case ast.BaseType:
			break
		case ast.TypeReference:
			switch c.ResolveType(t).(type) {
			case ast.BaseType, *ast.Enum, *ast.Typedef:
				break
			default:
				c.Errorf(st, "set value must be a primitive type")
			}
		default:
			c.Errorf(st, "set value must be a primitive type")
		}
	})
}
