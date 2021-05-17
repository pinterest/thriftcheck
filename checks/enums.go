package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func CheckEnumSize(warningLimit, errorLimit int) thriftcheck.Check {
	return thriftcheck.NewCheck("enum.size", func(c *thriftcheck.C, e *ast.Enum) {
		size := len(e.Items)
		if errorLimit > 0 && size > errorLimit {
			c.Errorf(e, "enumeration '%s' has more than %d items", e.Name, errorLimit)
		} else if warningLimit > 0 && size > warningLimit {
			c.Warningf(e, "enumeration '%s' has more than %d items", e.Name, warningLimit)
		}
	})
}
