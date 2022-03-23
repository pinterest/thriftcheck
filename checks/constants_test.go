// Copyright 2022 Pinterest
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

func TestCheckConstantRef(t *testing.T) {
	tests := []Test{
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Constant{Name: "Constant"},
			}},
			node: ast.ConstantReference{Name: "Constant"},
			want: []string{},
		},
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Enum{
					Name: "Enum",
					Items: []*ast.EnumItem{
						{Name: "Value"},
					},
				},
			}},
			node: ast.ConstantReference{Name: "Enum.Value"},
			want: []string{},
		},
		{
			prog: &ast.Program{},
			node: ast.ConstantReference{Name: "Unknown"},
			want: []string{
				`t.thrift:0:1: error: unable to find a constant or enum value named "Unknown" (constant.ref)`,
			},
		},
	}

	check := checks.CheckConstantRef()
	RunTests(t, check, tests)
}
