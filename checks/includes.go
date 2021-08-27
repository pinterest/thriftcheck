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
	"os"
	"path/filepath"
	"regexp"

	"github.com/danwakefield/fnmatch"
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckIncludePath returns a thriftcheck.Check that verifies that all of the
// files `include`'d by a Thrift file can be found in the includes paths.
func CheckIncludePath() *thriftcheck.Check {
	return thriftcheck.NewCheck("include.path", func(c *thriftcheck.C, i *ast.Include) {
		// If the path is absolute, we don't need to check the include paths.
		if filepath.IsAbs(i.Path) {
			if _, err := os.Stat(i.Path); err != nil {
				c.Errorf(i, "unable to read %q", i.Path)
			}
			return
		}

		found := false
		for _, dir := range c.Dirs {
			if _, err := os.Stat(filepath.Join(dir, i.Path)); err == nil {
				found = true
				break
			}
		}
		if !found {
			c.Errorf(i, "unable to find include file %q", i.Path)
		}
	})
}

// CheckIncludeRestricted returns a thriftcheck.Check that restricts some files
// from being imported by other  files using a map of patterns: the key is a
// file name pattern that matches the including filename and the value is a
// regular expression that matches the included filename. When both match, the
// `include` is flagged as "restricted" and an error is reported.
func CheckIncludeRestricted(patterns map[string]*regexp.Regexp) *thriftcheck.Check {
	return thriftcheck.NewCheck("include.restricted", func(c *thriftcheck.C, i *ast.Include) {
		for fpat, ire := range patterns {
			if fnmatch.Match(fpat, c.Filename, fnmatch.FNM_NOESCAPE) && ire.MatchString(i.Path) {
				c.Logf("%q (%s) matches %q (%s)\n", c.Filename, fpat, i.Path, ire)
				c.Errorf(i, "%q is a restricted import", i.Path)
				return
			}
		}
	})
}
