package checks

import (
	"regexp"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func CheckNamespacePattern(patterns map[string]string) thriftcheck.Check {
	regexps := make(map[string]*regexp.Regexp, len(patterns))
	for scope, pattern := range patterns {
		regexps[scope] = regexp.MustCompile(pattern)
	}

	return thriftcheck.NewCheck("namespace.pattern", func(c *thriftcheck.C, ns *ast.Namespace) {
		if re, ok := regexps[ns.Scope]; ok && !re.Match([]byte(ns.Name)) {
			c.Errorf(ns, "%q namespace must match %q", ns.Scope, re)
		}
	})
}
