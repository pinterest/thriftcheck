package checks_test

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func newC(check *thriftcheck.Check) *thriftcheck.C {
	return &thriftcheck.C{
		Logger:   log.New(ioutil.Discard, "", 0),
		Filename: "t.thrift",
		Check:    check.Name,
	}
}

func assertMessageStrings(t *testing.T, n ast.Node, expect []string, msgs thriftcheck.Messages) {
	strings := make([]string, len(msgs))
	for i, m := range msgs {
		strings[i] = m.String()
	}

	if !reflect.DeepEqual(strings, expect) {
		t.Errorf("%#v:\n- %v\n+ %v", n, expect, strings)
	}
}
