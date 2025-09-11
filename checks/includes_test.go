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

package checks_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckIncludeRestricted(t *testing.T) {
	tests := []Test{
		{
			name: "a.thrift",
			node: &ast.Include{Path: "good.thrift"},
			want: []string{},
		},
		{
			name: "a.thrift",
			node: &ast.Include{Path: "bad.thrift"},
			want: []string{
				`a.thrift:0:1: error: "bad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "a.thrift",
			node: &ast.Include{Path: "abad.thrift"},
			want: []string{
				`a.thrift:0:1: error: "abad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "b.thrift",
			node: &ast.Include{Path: "bad.thrift"},
			want: []string{
				`b.thrift:0:1: error: "bad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "b.thrift",
			node: &ast.Include{Path: "abad.thrift"},
			want: []string{
				`b.thrift:0:1: error: "abad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "nested/a.thrift",
			node: &ast.Include{Path: "good.thrift"},
			want: []string{},
		},
		{
			name: "nested/a.thrift",
			node: &ast.Include{Path: "bad.thrift"},
			want: []string{
				`nested/a.thrift:0:1: error: "bad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "nested/a.thrift",
			node: &ast.Include{Path: "inner.thrift"},
			want: []string{
				`nested/a.thrift:0:1: error: "inner.thrift" is a restricted import (include.restricted)`,
			},
		},
	}

	check := checks.CheckIncludeRestricted(map[string]*regexp.Regexp{
		"*":               regexp.MustCompile(`bad.thrift`),
		"a.thrift":        regexp.MustCompile(`abad.thrift`),
		"nested/*.thrift": regexp.MustCompile(`inner.thrift`),
	})
	RunTests(t, &check, tests)
}

// Note that the tests within each group are not independent.
// The data that gets added to the file dependency graph by a test
// will be there for the later tests.
func TestCheckIncludeCycle(t *testing.T) {
	// Simple cycle between two files.
	tests := []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check := checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle between files with absolute paths.
	tests = []Test{
		{
			name: "/home/a.thrift",
			dirs: []string{"/home"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "/home/b.thrift",
			dirs: []string{"/home"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`/home/b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	wd, _ := os.Getwd()

	// Cycle between a file with a relative path and another one with an absolute path.
	tests = []Test{
		{
			name: fmt.Sprintf("%s/a.thrift", wd),
			dirs: []string{wd},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with a file that includes another file from a nested directory.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/b.thrift"}}},
			want: []string{},
		},
		{
			name: "something/b.thrift",
			dirs: []string{"something"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../a.thrift"}}},
			want: []string{`something/b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle between two files inside of the same nested directory.
	tests = []Test{
		{
			name: "some/dir/a.thrift",
			dirs: []string{"some/dir"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "some/dir/b.thrift",
			dirs: []string{"some/dir"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`some/dir/b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with more complex directory inclusions.
	tests = []Test{
		{
			name: "d1/d11/d111/a.thrift",
			dirs: []string{"d1/d11/d111"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../../d12/d121/b.thrift"}}},
			want: []string{},
		},
		{
			name: "d1/d12/d121/b.thrift",
			dirs: []string{"d1/d12/d121"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../../../d2/c.thrift"}}},
			want: []string{},
		},
		{
			name: "d2/c.thrift",
			dirs: []string{"d2"},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../d1/d11/d111/a.thrift"}}},
			want: []string{`d2/c.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift -> c.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Self-loop.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`a.thrift:0:1: error: cycle detected: -> a.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with more files leading to and going out of it.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "d.thrift"}}},
			want: []string{},
		},
		{
			name: "d.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"},
				&ast.Include{Path: "g.thrift"}}},
			want: []string{`d.thrift:0:1: error: cycle detected: -> b.thrift -> c.thrift -> d.thrift (include.cycle)`},
		},
		{
			name: "g.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "h.thrift"}}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle where the involved files have more include statements.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other1.thrift"},
				&ast.Include{Path: "b.thrift"},
				&ast.Include{Path: "other2.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other1.thrift"},
				&ast.Include{Path: "c.thrift"},
				&ast.Include{Path: "other3.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other3.thrift"},
				&ast.Include{Path: "a.thrift"},
				&ast.Include{Path: "other4.thrift"}}},
			want: []string{`c.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift -> c.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with an include path going out of and back into the working directory.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../checks/b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// File that is involved in two cycles. Only one cycle is reported.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"},
				&ast.Include{Path: "b.thrift"}}},
			want: []string{`c.thrift:0:1: error: cycle detected: -> a.thrift -> c.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Two files that include the same file. No cycle.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "d.thrift"}}},
			want: []string{},
		},
		{
			name: "d.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "e.thrift"}}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Different files with the same filename. No cycle.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/a.thrift"},
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			dirs: []string{"."},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/a.thrift"}}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	included_dir := "shared"

	// Cycle involving a file from a dir included with the -I/--include option.
	tests = []Test{
		{
			name: "something/a.thrift",
			dirs: []string{"something", included_dir},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "shared/b.thrift",
			dirs: []string{"shared", included_dir},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../something/a.thrift"}}},
			want: []string{`shared/b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	included_dirs := []string{"abc/xyz", "other/shared"}

	// Cycle involving a file from a dir included with the -I/--include option.
	// Multiple included dirs.
	tests = []Test{
		{
			name: "a.thrift",
			dirs: append([]string{"."}, included_dirs...),
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "other/shared/b.thrift",
			dirs: append([]string{"other/shared"}, included_dirs...),
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../../a.thrift"}}},
			want: []string{`other/shared/b.thrift:0:1: error: cycle detected: -> a.thrift -> b.thrift (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// An included dir but no cycle.
	// The file in the included dir includes the other file but not the other way around.
	tests = []Test{
		{
			name: "something/a.thrift",
			dirs: []string{"something", included_dir},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "shared/b.thrift",
			dirs: []string{"shared", included_dir},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../something/a.thrift"}}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// An included dir but no cycle.
	// The file in the included dir does not include the other file.
	tests = []Test{
		{
			name: "something/a.thrift",
			dirs: []string{"something", included_dir},
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "shared/b.thrift",
			dirs: []string{"shared", included_dir},
			node: &ast.Program{Headers: []ast.Header{}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)
}
