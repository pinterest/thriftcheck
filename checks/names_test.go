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

func TestCheckNamesReserved(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Struct{Name: "struct"},
			want: []string{},
		},
		{
			node: &ast.Struct{Name: "reserved"},
			want: []string{
				`t.thrift:0:1: error: "reserved" is a reserved name (names.reserved)`,
			},
		},
		{
			node: &ast.Field{Name: "field"},
			want: []string{},
		},
		{
			node: &ast.Field{Name: "reserved"},
			want: []string{
				`t.thrift:0:1: error: "reserved" is a reserved name (names.reserved)`,
			},
		},
	}

	check := checks.CheckNamesReserved([]string{"reserved"})
	RunTests(t, check, tests)
}
