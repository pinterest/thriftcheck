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
	"errors"
	"fmt"
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

// CheckCircularIncludes returns a thriftcheck.Check that reports an error
// if there are circular includes.
func CheckCircularIncludes() *thriftcheck.Check {
	fn := func(c *thriftcheck.C, cc *circularIncludesC, i *ast.Include) {
		includer, err := getRelPath(c.Filename)

		if err != nil {
			c.Warningf(i, "could not get path relative to working directory for %s, skipping this include", i.Path)
			return
		}

		included, err := getRelPath(i.Path)

		if err != nil {
			c.Warningf(i, "could not get path relative to working directory for %s, skipping this include", i.Path)
			return
		}

		// a includes b
		a := getFilenameId(cc, includer)
		b := getFilenameId(cc, included)

		for _, v := range []int{a, b} {
			if _, exists := cc.adjList[v]; !exists {
				cc.inDegrees[v] = 0
				cc.adjList[v] = []int{}
			}
		}

		cc.inDegrees[b] += 1
		cc.adjList[a] = append(cc.adjList[a], b)

		if _, exists := cc.edgeMeta[a]; !exists {
			cc.edgeMeta[a] = make(map[int]*ast.Include)
			cc.edgeMeta[a][b] = i
		}
	}

	circularImportCtx := &circularIncludesC{
		adjList:      make(map[int][]int),
		edgeMeta:     make(map[int]map[int]*ast.Include),
		inDegrees:    make(map[int]int),
		filenameToId: make(map[string]int),
		idToFilename: make(map[int]string),
	}

	return thriftcheck.NewMultiFileCheck("import.cycle.disallowed", fn, circularImportCtx, func(cc *circularIncludesC) {
		imports, cycle := lookForCycle(cc.adjList, cc.inDegrees)

		if cycle {
			fmt.Println("Cycle detected:")

			for i, im := range imports {
				inc := cc.edgeMeta[im][imports[(i+1)%len(imports)]]
				fmt.Printf(
					"%s -> %s\n"+
						"\tIncluded as: %s\n"+
						"\tAt: %s:%d:%d\n\n",
					filepath.Base(cc.idToFilename[im]), filepath.Base(inc.Path),
					inc.Path,
					cc.idToFilename[im], inc.Line, inc.Column,
				)
			}
		}
	})
}

type circularIncludesC struct {
	adjList      map[int][]int
	edgeMeta     map[int]map[int]*ast.Include
	inDegrees    map[int]int
	filenameToId map[string]int
	idToFilename map[int]string
}

func getFilenameId(c *circularIncludesC, f string) int {
	if _, exists := c.filenameToId[f]; !exists {
		nextId := len(c.filenameToId) + 1

		c.filenameToId[f] = nextId
		c.idToFilename[nextId] = f
	}

	return c.filenameToId[f]
}

func getRelPath(f string) (string, error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", errors.New("could not get current working directory")
	}

	absP, err := filepath.Abs(f)

	if err != nil {
		return "", fmt.Errorf("could not get absolute path for %s", f)
	}

	return filepath.Rel(wd, absP)
}

// Topological processing
// https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm
func lookForCycle(adjList map[int][]int, inDegrees map[int]int) ([]int, bool) {
	count := 0
	sources := []int{}

	for v := range adjList {
		if inDegrees[v] == 0 {
			count += 1
			sources = append(sources, v)
		}
	}

	for len(sources) != 0 {
		newSources := []int{}

		for _, source := range sources {
			for _, v := range adjList[source] {
				inDegrees[v] -= 1
				if inDegrees[v] == 0 {
					count += 1
					newSources = append(sources, v)
				}
			}
		}

		sources = newSources
	}

	// there is at least one cycle,
	// so find the vertices of any of them
	if count != len(adjList) {
		return findCycleVertices(adjList), true
	}

	return nil, false
}

func findCycleVertices(adjList map[int][]int) []int {
	vis := make(map[int]bool)

	for v := range adjList {
		if vs := dfs(v, adjList, []int{}, make(map[int]bool), vis); vs != nil {
			return vs
		}
	}

	panic("unreachable (expected a cycle to exist)")
}

// Returns all of the vertices of a cycle if found, otherwise returns nil.
func dfs(cur int, adjList map[int][]int, vertices []int, vis map[int]bool, globalVis map[int]bool) []int {
	if vis[cur] {
		// return just the cycle (remove the vertices leading to it)
		for i, v := range vertices {
			if v == cur {
				return vertices[i:]
			}
		}
	}

	// path already explored
	if globalVis[cur] {
		return nil
	}

	vis[cur], globalVis[cur] = true, true

	for _, v := range adjList[cur] {
		if vs := dfs(v, adjList, append(vertices, cur), vis, globalVis); vs != nil {
			return vs
		}
	}

	return nil
}
