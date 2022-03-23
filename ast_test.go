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

package thriftcheck

import (
	"reflect"
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestDoc(t *testing.T) {
	tests := []struct {
		node ast.Node
		want string
	}{
		{&ast.Struct{}, ""},
		{&ast.Struct{Doc: ""}, ""},
		{&ast.Struct{Doc: "String"}, "String"},
	}

	for _, tt := range tests {
		got := Doc(tt.node)
		if got != tt.want {
			t.Errorf("expected %s but got %s", tt.want, got)
		}
	}
}

func TestResolveConstant(t *testing.T) {
	tests := []struct {
		ref  ast.ConstantReference
		prog *ast.Program
		want reflect.Type
		err  bool
	}{
		{
			ast.ConstantReference{Name: "Constant"},
			&ast.Program{Definitions: []ast.Definition{
				&ast.Constant{Name: "Constant"},
			}},
			reflect.TypeOf((*ast.Constant)(nil)),
			false,
		},
		{
			ast.ConstantReference{Name: "Enum.Value"}, &ast.Program{Definitions: []ast.Definition{
				&ast.Enum{
					Name: "Enum",
					Items: []*ast.EnumItem{
						{Name: "Value"},
					},
				},
			}}, reflect.TypeOf((*ast.EnumItem)(nil)),
			false,
		},
		{
			ast.ConstantReference{Name: "Unknown"},
			&ast.Program{},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		n, err := ResolveConstant(tt.ref, tt.prog, nil)
		if tt.err {
			if err == nil {
				t.Errorf("expected an error, got %s", n)
			}
		} else if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if got := reflect.TypeOf(n); tt.want != nil && got != tt.want {
			t.Errorf("expected %v but got %v", tt.want, got)
		}
	}
}
