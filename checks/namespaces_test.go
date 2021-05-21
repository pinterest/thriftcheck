package checks_test

import (
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckNamespacePattern(t *testing.T) {
	tests := []struct {
		node *ast.Namespace
		want []string
	}{
		{
			node: &ast.Namespace{Scope: "java", Name: "com.pinterest.idl.test"},
			want: []string{},
		},
		{
			node: &ast.Namespace{Scope: "java", Name: "com.example.idl.test"},
			want: []string{
				`t.thrift:0:1:error: "java" namespace must match "^com\\.pinterest\\.idl\\." (namespace.patterns)`,
			},
		},
	}

	check := checks.CheckNamespacePattern(map[string]string{
		"java": `^com\.pinterest\.idl\.`,
	})

	for _, tt := range tests {
		c := newC(&check)
		check.Call(c, tt.node)
		assertMessageStrings(t, tt.node, tt.want, c.Messages)
	}
}
