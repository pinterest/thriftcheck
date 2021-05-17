package checks

import (
	"strings"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func CheckNamespacePrefix(prefixes map[string]string) thriftcheck.Check {
	return thriftcheck.NewCheck("namespace.prefix", func(c *thriftcheck.C, ns *ast.Namespace) {
		if prefix, ok := prefixes[ns.Scope]; ok && !strings.HasPrefix(ns.Name, prefix) {
			c.Errorf(ns, "'%s' namespace must start with '%s'", ns.Scope, prefix)
		}
	})
}
