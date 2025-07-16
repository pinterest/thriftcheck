// Copyright 2025 Pinterest
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

func TestCheckTypesDisallowed(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Struct{Type: ast.UnionType},
			want: []string{
				`t.thrift:0:1: error: a disallowed type (union) was used (types.disallowed)`,
			},
		},
		{
			node: &ast.Struct{Type: ast.StructType},
			want: []string{},
		},
		{
			node: ast.ConstantInteger(0),
			want: []string{},
		},
	}

	check := checks.CheckTypesDisallowed([]string{"union"})
	RunTests(t, check, tests)
}
