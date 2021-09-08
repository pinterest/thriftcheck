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

package thriftcheck

import (
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestWithLogger(t *testing.T) {
	logger := log.New(ioutil.Discard, "", 0)
	linter := NewLinter(Checks{}, WithLogger(logger))
	if linter.logger != logger {
		t.Errorf("expected logger to be %v, got %v", logger, linter.logger)
	}
}

func TestWithIncludes(t *testing.T) {
	includes := []string{"a", "b"}
	linter := NewLinter(Checks{}, WithIncludes(includes))
	if !reflect.DeepEqual(linter.includes, includes) {
		t.Errorf("expected includes to be %v, got %v", includes, linter.includes)
	}
}

func TestLint(t *testing.T) {
	linter := NewLinter(Checks{
		NewCheck("node", func(c *C, n ast.Node) { c.Errorf(n, "node") }),
		NewCheck("field", func(c *C, f *ast.Field) { c.Errorf(f, "field") }),
		NewCheck("enumitem", func(c *C, f *ast.EnumItem) { c.Errorf(f, "enumitem") }),
	})

	s := strings.NewReader(`
		struct TestStruct {
			1: string field1
			2: bool field2
		}

		enum TestEnum {
			ONE = 1
			TWO = 2
		}
	`)

	msgs, err := linter.Lint(s, "t.thrift")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	counts := make(map[string]int)
	for _, m := range msgs {
		counts[m.Check]++
	}

	if counts["node"] != 9 {
		t.Errorf("expected 9 node; got %v", counts)
	}
	if counts["field"] != 2 {
		t.Errorf("expected 2 fields; got %v", counts)
	}
	if counts["enumitem"] != 2 {
		t.Errorf("expected 2 enumitems; got %v", counts)
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		s    string
		want []string
	}{
		{
			s: `namespace`,
			want: []string{
				`t.thrift:1:1: error: syntax error: unexpected $end, expecting IDENTIFIER or '*' (parse)`,
			},
		},
		{
			s: `struct S {}}`,
			want: []string{
				`t.thrift:1:12: error: syntax error: unexpected '}' (parse)`,
			},
		},
	}

	linter := NewLinter(Checks{})
	for _, tt := range tests {
		msgs, err := linter.Lint(strings.NewReader(tt.s), "t.thrift")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		strings := make([]string, len(msgs))
		for i, m := range msgs {
			strings[i] = m.String()
		}
		if !reflect.DeepEqual(strings, tt.want) {
			t.Errorf("%s:\n- %v\n+ %v", tt.s, tt.want, strings)
		}
	}
}

func TestNoLint(t *testing.T) {
	linter := NewLinter(Checks{
		NewCheck("check.warn", func(c *C, n ast.Node) { c.Warningf(n, "") }),
		NewCheck("check.error", func(c *C, n ast.Node) { c.Errorf(n, "") }),
	})

	nolint := func(v string) *ast.Annotation {
		return &ast.Annotation{Name: "nolint", Value: v}
	}

	type test struct {
		desc   string
		node   *ast.Program
		checks map[ast.Node][]string
	}

	tests := []test{
		func() (tt test) {
			a := nolint("")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("check")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint:check"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("check.warn")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint:check.warn"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{
				a: {"check.error"},
				f: {"check.error"},
				n: {"check.error"},
			}
			return
		}(),
		func() (tt test) {
			a := nolint("check.warn, check.error")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint:check.warn,check.error"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("")
			f := &ast.Field{Annotations: []*ast.Annotation{a}}
			n := &ast.Struct{Fields: []*ast.Field{f}}

			tt.desc = "field nolint"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{
				n: {"check.warn", "check.error"},
			}
			return
		}(),
		func() (tt test) {
			a := nolint("check")
			f := &ast.Field{Annotations: []*ast.Annotation{a}}
			n := &ast.Struct{Fields: []*ast.Field{f}}

			tt.desc = "field nolint:check"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{
				n: {"check.warn", "check.error"},
			}
			return
		}(),
		func() (tt test) {
			a := nolint("check.warn")
			f := &ast.Field{Annotations: []*ast.Annotation{a}}
			n := &ast.Struct{Fields: []*ast.Field{f}}

			tt.desc = "field nolint:check.warn"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{
				a: {"check.error"},
				f: {"check.error"},
				n: {"check.warn", "check.error"},
			}
			return
		}(),
		func() (tt test) {
			a := nolint("check.warn, check.error")
			f := &ast.Field{Annotations: []*ast.Annotation{a}}
			n := &ast.Struct{Fields: []*ast.Field{f}}

			tt.desc = "field nolint:check.warn,check.error"
			tt.node = &ast.Program{Definitions: []ast.Definition{n}}
			tt.checks = map[ast.Node][]string{
				n: {"check.warn", "check.error"},
			}
			return
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := map[ast.Node][]string{}
			for _, m := range linter.lint(tt.node, "filename.thrift", nil) {
				if _, ok := m.Node.(*ast.Program); !ok {
					actual[m.Node] = append(actual[m.Node], m.Check)
				}
			}
			if !reflect.DeepEqual(actual, tt.checks) {
				t.Errorf("expected %#v, got %#v", tt.checks, actual)
			}
		})
	}
}

func TestOverrideableChecksLookup(t *testing.T) {
	root := &Checks{&Check{Name: "root"}}
	pnode := &ast.Program{}
	snode := &ast.Struct{}
	fnode := &ast.Field{}
	schecks := &Checks{&Check{Name: "s"}}

	checks := overridableChecks{root: root}
	checks.add(snode, schecks)

	tests := []struct {
		nodes    []ast.Node
		expected *Checks
	}{
		{[]ast.Node{pnode}, root},
		{[]ast.Node{snode, pnode}, schecks},
		{[]ast.Node{fnode, snode, pnode}, schecks},
		{[]ast.Node{}, root}, // always last because it clears overrides
	}

	for _, tt := range tests {
		actual := checks.lookup(tt.nodes)
		if actual != tt.expected {
			t.Errorf("lookup for %#v got %#v, expected %#v", tt.nodes, actual, tt.expected)
		}
	}
}

func TestOverrideableChecksOverrides(t *testing.T) {
	root := &Checks{}
	pnode := &ast.Program{}
	snode := &ast.Struct{}
	checks := overridableChecks{root: root}

	checks.add(snode, &Checks{})
	if len(checks.overrides) != 1 {
		t.Errorf("expected 1 override but got %#v", checks.overrides)
	}
	if checks.lookup([]ast.Node{pnode}) != root {
		t.Errorf("expected %#v lookup to return the root", pnode)
	}
	if checks.lookup([]ast.Node{snode, pnode}) == root {
		t.Errorf("expected %#v lookup to return the override", snode)
	}

	if checks.lookup([]ast.Node{}) != root {
		t.Errorf("expected empty lookup to return the root")
	}
	if len(checks.overrides) != 0 {
		t.Errorf("expected overrides to be cleared")
	}
}
