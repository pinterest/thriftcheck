package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func CheckMapKeyType() thriftcheck.Check {
	return thriftcheck.NewCheck("map.key.type", func(c *thriftcheck.C, mt ast.MapType) {
		if _, ok := mt.KeyType.(ast.BaseType); !ok {
			c.Errorf(mt, "map key must be a primitive type")
		}
	})
}
