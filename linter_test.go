package thriftcheck

import (
	"io/ioutil"
	"log"
	"reflect"
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
		node   ast.Node
		checks map[ast.Node][]string
	}

	tests := []test{
		func() (tt test) {
			a := nolint("")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint"
			tt.node = n
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("check")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint:check"
			tt.node = n
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("check.warn")
			f := &ast.Field{}
			n := &ast.Struct{Annotations: []*ast.Annotation{a}, Fields: []*ast.Field{f}}

			tt.desc = "struct nolint:check.warn"
			tt.node = n
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
			tt.node = n
			tt.checks = map[ast.Node][]string{}
			return
		}(),
		func() (tt test) {
			a := nolint("")
			f := &ast.Field{Annotations: []*ast.Annotation{a}}
			n := &ast.Struct{Fields: []*ast.Field{f}}

			tt.desc = "field nolint"
			tt.node = n
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
			tt.node = n
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
			tt.node = n
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
			tt.node = n
			tt.checks = map[ast.Node][]string{
				n: {"check.warn", "check.error"},
			}
			return
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := map[ast.Node][]string{}
			for _, m := range linter.lint(tt.node, "filename.thrift") {
				actual[m.Node] = append(actual[m.Node], m.Check)
			}
			if !reflect.DeepEqual(actual, tt.checks) {
				t.Errorf("expected %#v, got %#v", tt.checks, actual)
			}
		})
	}
}

func TestOverrideableChecksLookup(t *testing.T) {
	root := &Checks{Check{Name: "root"}}
	pnode := &ast.Program{}
	snode := &ast.Struct{}
	fnode := &ast.Field{}
	schecks := &Checks{Check{Name: "s"}}

	checks := overridableChecks{
		root:      root,
		overrides: map[ast.Node]*Checks{snode: schecks},
	}

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
