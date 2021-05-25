package checks

import (
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckEnumSize returns a thriftcheck.Check that warns or errors if an
// enumeration's element size grows beyond a limit.
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
