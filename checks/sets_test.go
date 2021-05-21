package checks_test

import (
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckSetValueType(t *testing.T) {
	tests := []struct {
		node ast.SetType
		want []string
	}{
		{
			node: ast.SetType{ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
		{
			node: ast.SetType{ValueType: ast.SetType{ValueType: ast.BaseType{ID: ast.StringTypeID}}},
			want: []string{
				`t.thrift:0:1:error: set value must be a primitive type (set.value.type)`,
			},
		},
	}

	check := checks.CheckSetValueType()

	for _, tt := range tests {
		c := newC(&check)
		check.Call(c, tt.node)
		assertMessageStrings(t, tt.node, tt.want, c.Messages)
	}
}
