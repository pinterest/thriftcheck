package checks_test

import (
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckEnumSize(t *testing.T) {
	tests := []struct {
		node *ast.Enum
		want []string
	}{
		{
			node: &ast.Enum{Name: "enum"},
			want: []string{},
		},
		{
			node: &ast.Enum{Name: "enum", Items: []*ast.EnumItem{{}, {}}},
			want: []string{
				`t.thrift:0:1:warning: enumeration 'enum' has more than 1 items (enum.size)`,
			},
		},
		{
			node: &ast.Enum{Name: "enum", Items: []*ast.EnumItem{{}, {}, {}}},
			want: []string{
				`t.thrift:0:1:error: enumeration 'enum' has more than 2 items (enum.size)`,
			},
		},
	}

	check := checks.CheckEnumSize(1, 2)

	for _, tt := range tests {
		c := newC(&check)
		check.Call(c, tt.node)
		assertMessageStrings(t, tt.node, tt.want, c.Messages)
	}
}
