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
	"path"
	"path/filepath"
	"regexp"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckIncludePath returns a thriftcheck.Check that verifies that all of the
// files `include`'d by a Thrift file can be found in the includes paths.
func CheckIncludePath() *thriftcheck.Check {
	return thriftcheck.NewCheck("include.path", func(c *thriftcheck.C, i *ast.Include) {
		// Always check the file's directory first to match `thrift`s behavior.
		dirs := append([]string{filepath.Dir(c.Filename)}, c.Includes...)

		found := false
		for _, dir := range dirs {
			if _, err := os.Stat(path.Join(dir, i.Path)); err == nil {
				found = true
				break
			}
		}
		if !found {
			c.Errorf(i, "unable to find include path for %q", i.Path)
		}
	})
}

// CheckIncludeRestricted returns a thriftcheck.Check that restricts some files
// from being imported by other  files using a map of patterns: the key is a
// file name pattern that matches the including filename and the value is a
// regular expression that matches the included filename. When both match, the
// `include` is flagged as "restricted" and an error is reported.
func CheckIncludeRestricted(patterns map[string]string) *thriftcheck.Check {
	regexps := make(map[string]*regexp.Regexp, len(patterns))
	for fpat, ipat := range patterns {
		regexps[fpat] = regexp.MustCompile(ipat)
	}

	return thriftcheck.NewCheck("include.restricted", func(c *thriftcheck.C, i *ast.Include) {
		for fpat, ire := range regexps {
			if ok, _ := filepath.Match(fpat, c.Filename); ok && ire.MatchString(i.Path) {
				c.Logf("%q (%s) matches %q (%s)\n", c.Filename, fpat, i.Path, ire)
				c.Errorf(i, "%q is a restricted import", i.Path)
				return
			}
		}
	})
}
