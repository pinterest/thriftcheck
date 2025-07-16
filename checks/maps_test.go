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

func TestCheckMapKeyType(t *testing.T) {
	tests := []Test{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
		{
			node: ast.MapType{
				KeyType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.StringTypeID},
					ValueType: ast.BaseType{ID: ast.StringTypeID}},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{
				`t.thrift:0:1: error: map key must be a primitive type (map.key.type)`,
			},
		},
		{
			prog: &ast.Program{},
			node: ast.MapType{
				KeyType:   ast.TypeReference{Name: "Enum"},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{
				`t.thrift:0:1: error: map key must be a primitive type (map.key.type)`,
			},
		},
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Enum{Name: "Enum"},
			}},
			node: ast.MapType{
				KeyType:   ast.TypeReference{Name: "Enum"},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
	}

	check := checks.CheckMapKeyType()
	RunTests(t, &check, tests)
}

func TestCheckMapNested(t *testing.T) {
	tests := []Test{
		{
			// Valid flat map - should pass
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
		{
			// Direct nested map - should fail
			node: ast.MapType{
				KeyType: ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.I64TypeID},
					ValueType: ast.BaseType{ID: ast.StringTypeID}}},
			want: []string{
				`t.thrift:0:1: error: nested maps are not allowed; use flat map structures instead (map.value.nested)`,
			},
		},
		{
			// TypeReference that doesn't resolve - should pass
			prog: &ast.Program{},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "UnknownType"}},
			want: []string{},
		},
		{
			// TypeReference that resolves to a map - should fail
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{
					Name: "NestedMapType",
					Type: ast.MapType{
						KeyType:   ast.BaseType{ID: ast.I64TypeID},
						ValueType: ast.BaseType{ID: ast.StringTypeID}},
				},
			}},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "NestedMapType"}},
			want: []string{
				`t.thrift:0:1: error: nested maps are not allowed; use flat map structures instead (map.value.nested)`,
			},
		},
		{
			// TypeReference that resolves to a non-map - should pass
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "MyStruct"},
			}},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "MyStruct"}},
			want: []string{},
		},
	}

	check := checks.CheckMapNested()
	RunTests(t, check, tests)
}
