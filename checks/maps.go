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

package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckMapKeyType returns a thriftcheck.Check that ensures that only primitive
// types are used for `map<>` keys.
func CheckMapKeyType() thriftcheck.Check {
	return thriftcheck.NewCheck("map.key.type", func(c *thriftcheck.C, mt ast.MapType) {
		switch t := mt.KeyType.(type) {
		case ast.BaseType:
			break
		case ast.TypeReference:
			switch c.ResolveType(t).(type) {
			case ast.BaseType, *ast.Enum, *ast.Typedef:
				break
			default:
				c.Errorf(mt, "map key must be a primitive type")
			}
		default:
			c.Errorf(mt, "map key must be a primitive type")
		}
	})
}

// CheckMapNested returns a thriftcheck.Check that ensures maps are not nested.
// Nested maps (map<K, map<K2, V>>) are disallowed to enforce flat map structures.
func CheckMapNested() *thriftcheck.Check {
	return thriftcheck.NewCheck("map.value.nested", func(c *thriftcheck.C, mt ast.MapType) {
		// Check if the value type is directly a MapType
		if _, isMap := mt.ValueType.(ast.MapType); isMap {
			c.Errorf(mt, "nested maps are not allowed; use flat map structures instead")
			return
		}

		// Check if the value type is a TypeReference that resolves to a MapType
		if typeRef, isTypeRef := mt.ValueType.(ast.TypeReference); isTypeRef {
			if resolved := c.ResolveType(typeRef); resolved != nil {
				if _, isMap := resolved.(ast.MapType); isMap {
					c.Errorf(mt, "nested maps are not allowed; use flat map structures instead")
				}
			}
		}
	})
}
