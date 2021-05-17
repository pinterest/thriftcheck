package thriftcheck

import (
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestMessageString(t *testing.T) {
	node := &ast.Struct{}
	tests := []struct {
		m *Message
		s string
	}{
		{
			&Message{Filename: "a.thrift", Node: node, Check: "check", Severity: Warning, Message: "Warning"},
			"a.thrift:0:1:warning: Warning (check)",
		},
		{
			&Message{Filename: "a.thrift", Node: node, Check: "check", Severity: Error, Message: "Error"},
			"a.thrift:0:1:error: Error (check)",
		},
	}

	for _, tt := range tests {
		s := tt.m.String()
		if s != tt.s {
			t.Errorf("%#v expected %q, got %q", tt.m, tt.s, s)
		}
	}
}
