// Copyright 2021 Pinterest
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
	"regexp"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckNamespacePattern returns a thriftcheck.Check that ensures that a
// namespace's name matches a regular expression pattern. The pattern can
// be configured one a per-language basis.
func CheckNamespacePattern(patterns map[string]*regexp.Regexp) thriftcheck.Check {
	return thriftcheck.NewCheck("namespace.patterns", func(c *thriftcheck.C, ns *ast.Namespace) {
		if re, ok := patterns[ns.Scope]; ok && !re.MatchString(ns.Name) {
			c.Errorf(ns, "%q namespace must match %q", ns.Scope, re)
		}
	})
}
