// Copyright 2025 Pinterest
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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/danwakefield/fnmatch"
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckIncludePath returns a thriftcheck.Check that verifies that all of the
// files `include`'d by a Thrift file can be found in the includes paths.
func CheckIncludePath() thriftcheck.Check {
	return thriftcheck.NewCheck("include.path", func(c *thriftcheck.C, i *ast.Include) {
		// If the path is absolute, we don't need to check the include paths.
		if filepath.IsAbs(i.Path) {
			if info, err := os.Stat(i.Path); err != nil || info.IsDir() {
				c.Errorf(i, "unable to read file %q", i.Path)
			}
			return
		}

		found := false
		for _, dir := range c.Dirs {
			path := filepath.Join(dir, i.Path)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
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
func CheckIncludeRestricted(patterns map[string]*regexp.Regexp) thriftcheck.Check {
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

type includeEdge struct {
	originalFrom string
	to           string
	include      *ast.Include
}

// CheckIncludeCycle returns a thriftcheck.Check that reports an error
// if there is a circular include.
func CheckIncludeCycle() thriftcheck.Check {
	edgeList := make(map[string][]includeEdge)

	return thriftcheck.NewCheck("include.cycle", func(c *thriftcheck.C, p *ast.Program) {
		// a `include`s b
		var a, b string

		a, err := filepath.Abs(c.Filename)
		if err != nil {
			c.Warningf(p, "could not get absolute path for %s, skipping this file", c.Filename)
			return
		}

		for _, h := range p.Headers {
			i, ok := h.(*ast.Include)
			if !ok {
				continue
			}

			b = filepath.Join(filepath.Dir(a), i.Path)

			edgeList[a] = append(edgeList[a], includeEdge{originalFrom: c.Filename, to: b, include: i})
		}

		cycle := lookForCycle(a, a, make(map[string]bool), []includeEdge{}, edgeList)

		if len(cycle) > 0 {
			m := []string{}
			for _, e := range cycle {
				m = append(m, fmt.Sprintf(
					"\t%s -> %s\n\t\tIncluded as: %s\n\t\tAt: %s:%d:%d\n",
					filepath.Base(e.originalFrom), filepath.Base(e.include.Path), e.include.Path, e.originalFrom, e.include.Line, e.include.Column))
			}
			c.Errorf(p, "Cycle detected:\n%s", strings.Join(m, "\n"))
		}
	})
}

// looksForCycle tries to find a cycle that leads back to the start node (filename).
// If found, it returns the edges in the cycle. Otherwise returns nil.
func lookForCycle(cur, start string, vis map[string]bool, path []includeEdge, edgeList map[string][]includeEdge) []includeEdge {
	if vis[cur] {
		if cur == start {
			return path
		}
		return nil
	}

	vis[cur] = true

	for _, e := range edgeList[cur] {
		if cycle := lookForCycle(e.to, start, vis, append(path, e), edgeList); cycle != nil {
			return cycle
		}
	}

	return nil
}
