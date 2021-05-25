package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckSetValueType returns a thriftcheck.Check that ensures that only primitive
// types are used for `set<>` values.
func CheckSetValueType() thriftcheck.Check {
	return thriftcheck.NewCheck("set.value.type", func(c *thriftcheck.C, st ast.SetType) {
		if _, ok := st.ValueType.(ast.BaseType); !ok {
			c.Errorf(st, "set value must be a primitive type")
		}
	})
}
