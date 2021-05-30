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
	t.Helper()

	strings := make([]string, len(msgs))
	for i, m := range msgs {
		strings[i] = m.String()
	}

	if !reflect.DeepEqual(strings, expect) {
		t.Errorf("%#v:\n- %v\n+ %v", n, expect, strings)
	}
}
