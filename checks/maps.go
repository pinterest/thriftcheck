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

// CheckMapKeyType returns a thriftcheck.Check that checks if a `map<>` key
// type is allowed.
func CheckMapKeyType(allowedTypes, disallowedTypes []thriftcheck.ThriftType) thriftcheck.Check {
	return thriftcheck.NewCheck("map.key.type", func(c *thriftcheck.C, mt ast.MapType) {
		if ok, name := c.CheckType(mt.KeyType, allowedTypes, disallowedTypes); !ok {
			c.Errorf(mt, "map key type %q is not allowed", name)
		}
	})
}

// CheckMapKeyType returns a thriftcheck.Check that checks if a `map<>` value
// type is allowed.
func CheckMapValueType(allowedTypes, disallowedTypes []thriftcheck.ThriftType) thriftcheck.Check {
	return thriftcheck.NewCheck("map.value.type", func(c *thriftcheck.C, mt ast.MapType) {
		if ok, name := c.CheckType(mt.ValueType, allowedTypes, disallowedTypes); !ok {
			c.Errorf(mt, "map value type %q is not allowed", name)
		}
	})
}
