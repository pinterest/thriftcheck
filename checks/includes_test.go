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
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: a.thrift
		At: b.thrift:0:0

	a.thrift -> b.thrift
		Included as: b.thrift
		At: a.thrift:0:0
 (include.cycle)`},
		},
	}

	check := checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle between files with absolute paths.
	tests = []Test{
		{
			name: "/home/a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "/home/b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`/home/b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: a.thrift
		At: /home/b.thrift:0:0

	a.thrift -> b.thrift
		Included as: b.thrift
		At: /home/a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	wd, _ := os.Getwd()

	// Cycle between a file with a relative path and another one with an absolute path.
	tests = []Test{
		{
			name: fmt.Sprintf("%s/a.thrift", wd),
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{fmt.Sprintf(`b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: a.thrift
		At: b.thrift:0:0

	a.thrift -> b.thrift
		Included as: b.thrift
		At: %s/a.thrift:0:0
 (include.cycle)`, wd)},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with a file that includes another file from a nested directory.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/b.thrift"}}},
			want: []string{},
		},
		{
			name: "something/b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../a.thrift"}}},
			want: []string{`something/b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: ../a.thrift
		At: something/b.thrift:0:0

	a.thrift -> b.thrift
		Included as: something/b.thrift
		At: a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle between two files inside of the same nested directory.
	tests = []Test{
		{
			name: "some/dir/a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "some/dir/b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`some/dir/b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: a.thrift
		At: some/dir/b.thrift:0:0

	a.thrift -> b.thrift
		Included as: b.thrift
		At: some/dir/a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with more complex directory inclusions.
	tests = []Test{
		{
			name: "d1/d11/d111/a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../../d12/d121/b.thrift"}}},
			want: []string{},
		},
		{
			name: "d1/d12/d121/b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../../../d2/c.thrift"}}},
			want: []string{},
		},
		{
			name: "d2/c.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../d1/d11/d111/a.thrift"}}},
			want: []string{`d2/c.thrift:0:1: error: Cycle detected:
	c.thrift -> a.thrift
		Included as: ../d1/d11/d111/a.thrift
		At: d2/c.thrift:0:0

	a.thrift -> b.thrift
		Included as: ../../d12/d121/b.thrift
		At: d1/d11/d111/a.thrift:0:0

	b.thrift -> c.thrift
		Included as: ../../../d2/c.thrift
		At: d1/d12/d121/b.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Self-loop.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`a.thrift:0:1: error: Cycle detected:
	a.thrift -> a.thrift
		Included as: a.thrift
		At: a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with more files leading to and going out of it.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "d.thrift"}}},
			want: []string{},
		},
		{
			name: "d.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "b.thrift"},
				&ast.Include{Path: "g.thrift"}}},
			want: []string{`d.thrift:0:1: error: Cycle detected:
	d.thrift -> b.thrift
		Included as: b.thrift
		At: d.thrift:0:0

	b.thrift -> c.thrift
		Included as: c.thrift
		At: b.thrift:0:0

	c.thrift -> d.thrift
		Included as: d.thrift
		At: c.thrift:0:0
 (include.cycle)`},
		},
		{
			name: "g.thrift",
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
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other1.thrift"},
				&ast.Include{Path: "b.thrift"},
				&ast.Include{Path: "other2.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other1.thrift"},
				&ast.Include{Path: "c.thrift"},
				&ast.Include{Path: "other3.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "other3.thrift"},
				&ast.Include{Path: "a.thrift"},
				&ast.Include{Path: "other4.thrift"}}},
			want: []string{`c.thrift:0:1: error: Cycle detected:
	c.thrift -> a.thrift
		Included as: a.thrift
		At: c.thrift:0:0

	a.thrift -> b.thrift
		Included as: b.thrift
		At: a.thrift:0:0

	b.thrift -> c.thrift
		Included as: c.thrift
		At: b.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Cycle with an include path going out of and back into the working directory.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "../checks/b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"}}},
			want: []string{`b.thrift:0:1: error: Cycle detected:
	b.thrift -> a.thrift
		Included as: a.thrift
		At: b.thrift:0:0

	a.thrift -> b.thrift
		Included as: ../checks/b.thrift
		At: a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// File that is involved in two cycles. Only one cycle is reported.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "a.thrift"},
				&ast.Include{Path: "b.thrift"}}},
			want: []string{`c.thrift:0:1: error: Cycle detected:
	c.thrift -> a.thrift
		Included as: a.thrift
		At: c.thrift:0:0

	a.thrift -> c.thrift
		Included as: c.thrift
		At: a.thrift:0:0
 (include.cycle)`},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)

	// Two files that include the same file. No cycle.
	tests = []Test{
		{
			name: "a.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "c.thrift"}}},
			want: []string{},
		},
		{
			name: "c.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "d.thrift"}}},
			want: []string{},
		},
		{
			name: "d.thrift",
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
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/a.thrift"},
				&ast.Include{Path: "b.thrift"}}},
			want: []string{},
		},
		{
			name: "b.thrift",
			node: &ast.Program{Headers: []ast.Header{
				&ast.Include{Path: "something/a.thrift"}}},
			want: []string{},
		},
	}

	check = checks.CheckIncludeCycle()
	RunTests(t, &check, tests)
}
