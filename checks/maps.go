package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckMapKeyType returns a thriftcheck.Check that ensures that only primitive
// types are used for `map<>` keys.
func CheckMapKeyType() thriftcheck.Check {
	return thriftcheck.NewCheck("map.key.type", func(c *thriftcheck.C, mt ast.MapType) {
		if _, ok := mt.KeyType.(ast.BaseType); !ok {
			c.Errorf(mt, "map key must be a primitive type")
		}
	})
}
