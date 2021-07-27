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
	"reflect"
	"testing"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

type Test struct {
	name string
	prog *ast.Program
	node ast.Node
	want []string
}

func RunTests(t *testing.T, check *thriftcheck.Check, tests []Test) {
	t.Helper()

	for _, tt := range tests {
		c := &thriftcheck.C{
			Filename: tt.name,
			Program:  tt.prog,
			Check:    check.Name,
		}
		if c.Filename == "" {
			c.Filename = "t.thrift"
		}

		check.Call(c, tt.node)

		if len(tt.want) > 0 || len(c.Messages) > 0 {
			strings := make([]string, len(c.Messages))
			for i, m := range c.Messages {
				strings[i] = m.String()
			}
			if !reflect.DeepEqual(strings, tt.want) {
				t.Errorf("%#v:\n- %v\n+ %v", tt.node, tt.want, strings)
			}
		}
	}
}
