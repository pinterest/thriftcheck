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
	"github.com/pinterest/thriftcheck/utils"
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

// CheckMapValueType returns a thriftcheck.Check that ensures map values don't use restricted types.
// The restrictedTypes slice allows configurable type restrictions.
// Common use cases: disallow nested maps, unions, complex collections, etc.
func CheckMapValueType(restrictedTypes []string) *thriftcheck.Check {
	restrictedTypeMatchers, err := utils.ParseTypes(restrictedTypes)
	if err != nil {
		return nil
	}
	return thriftcheck.NewCheck("map.value", func(c *thriftcheck.C, mt ast.MapType) {
		// If no restrictions configured, this is a no-op
		if len(restrictedTypes) == 0 {
			return
		}

		// Check if the value type matches any restricted types
		for _, matcher := range restrictedTypeMatchers {
			if matcher.Matches(c, mt.ValueType) {
				c.Errorf(mt, "map value type %s is restricted", matcher.Name())
				return
			}
		}
	})
}
