package thriftcheck

import (
	"reflect"
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestAnnotations(t *testing.T) {
	annotations := []*ast.Annotation{
		{Name: "test1", Value: "value1"},
		{Name: "test2", Value: "value2"},
	}
	tests := []struct {
		node     ast.Node
		expected []*ast.Annotation
	}{
		{&ast.Struct{}, nil},
		{&ast.Struct{Annotations: annotations}, annotations},
		{&ast.Struct{Annotations: annotations[1:]}, annotations[1:]},
		{&ast.Annotation{}, nil},
	}

	for _, tt := range tests {
		actual := Annotations(tt.node)
		if !reflect.DeepEqual(actual, tt.expected) {
			t.Errorf("expected %s but got %s", tt.expected, actual)
		}
	}
}
