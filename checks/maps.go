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

// CheckMapKeyType reports an error if a disallowed map key type is used.
// A type is disallowed if it either:
//   - Appears in `disallowedTypes`
//   - Does not appear in a non-empty `allowedTypes`
func CheckMapKeyType(allowedTypes, disallowedTypes []thriftcheck.ThriftType) thriftcheck.Check {
	return thriftcheck.NewCheck("map.key.type", func(c *thriftcheck.C, mt ast.MapType) {
		for _, matcher := range disallowedTypes {
			if matcher.Matches(c, mt.KeyType) {
				c.Errorf(mt, "map key type %q is disallowed", matcher)
				return
			}
		}

		if len(allowedTypes) == 0 {
			return
		}
		for _, matcher := range allowedTypes {
			if matcher.Matches(c, mt.KeyType) {
				return
			}
		}
		c.Errorf(mt, "map key type %q is not in the 'allowed' list", mt.KeyType)
	})
}

// CheckMapValueType returns a thriftcheck.Check that ensures map values don't use disallowed types.
// The disallowedTypes slice allows configurable type disallowances.
// Common use cases: disallow nested maps, unions, complex collections, etc.
func CheckMapValueType(disallowedTypes []thriftcheck.ThriftType) thriftcheck.Check {
	return thriftcheck.NewCheck("map.value.disallowed", func(c *thriftcheck.C, mt ast.MapType) {
		for _, matcher := range disallowedTypes {
			if matcher.Matches(c, mt.ValueType) {
				c.Errorf(mt, "map value type %s is disallowed", matcher)
				return
			}
		}
	})
}
