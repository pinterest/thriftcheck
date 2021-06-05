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
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckNamespacePattern(t *testing.T) {
	tests := []Test{
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
	RunTests(t, &check, tests)
}
