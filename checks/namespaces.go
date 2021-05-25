package checks

import (
	"regexp"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckNamespacePattern returns a thriftcheck.Check that ensures that a
// namespace's name matches a regular expression pattern. The pattern can
// be configured one a per-language basis.
func CheckNamespacePattern(patterns map[string]string) thriftcheck.Check {
	regexps := make(map[string]*regexp.Regexp, len(patterns))
	for scope, pattern := range patterns {
		regexps[scope] = regexp.MustCompile(pattern)
	}

	return thriftcheck.NewCheck("namespace.patterns", func(c *thriftcheck.C, ns *ast.Namespace) {
		if re, ok := regexps[ns.Scope]; ok && !re.Match([]byte(ns.Name)) {
			c.Errorf(ns, "%q namespace must match %q", ns.Scope, re)
		}
	})
}
