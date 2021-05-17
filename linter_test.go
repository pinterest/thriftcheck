package thriftcheck

import (
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestOverrideableChecksLookup(t *testing.T) {
	root := &Checks{Check{key: "root"}}
	pnode := &ast.Program{}
	snode := &ast.Struct{}
	fnode := &ast.Field{}
	schecks := &Checks{Check{key: "s"}}

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
