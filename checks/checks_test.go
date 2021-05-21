package checks_test

import (
	"reflect"
	"testing"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func assertMessageStrings(t *testing.T, n ast.Node, expect []string, msgs thriftcheck.Messages) {
	strings := make([]string, len(msgs))
	for i, m := range msgs {
		strings[i] = m.String()
	}

	if !reflect.DeepEqual(strings, expect) {
		t.Errorf("%#v:\n- %v\n+ %v", n, expect, strings)
	}
}
