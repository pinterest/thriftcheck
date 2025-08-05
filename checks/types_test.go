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

	"github.com/pinterest/thriftcheck"
	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckTypesDisallowed(t *testing.T) {
	unionType := ParseType(t, "union")
	mapType := ParseType(t, "map")

	// Tests with a single disallowed type (union).
	tests := []Test{
		{
			node: &ast.Struct{Type: ast.UnionType},
			want: []string{
				`t.thrift:0:1: error: type "union" is not allowed (types)`,
			},
		},
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "MyUnion", Type: ast.UnionType},
			}},
			node: ast.TypeReference{Name: "MyUnion"},
			want: []string{
				`t.thrift:0:1: error: type "union" is not allowed (types)`,
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

	check := checks.CheckTypes([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{unionType})
	RunTests(t, &check, tests)

	// Tests with multiple disallowed types (union and map).
	tests = []Test{
		{
			node: &ast.Struct{Type: ast.UnionType},
			want: []string{
				`t.thrift:0:1: error: type "union" is not allowed (types)`,
			},
		},
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.I16TypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{
				`t.thrift:0:1: error: type "map" is not allowed (types)`,
			},
		},
		{
			node: &ast.Struct{Type: ast.ExceptionType},
			want: []string{},
		},
	}

	check = checks.CheckTypes([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{unionType, mapType})
	RunTests(t, &check, tests)

	// Tests with no disallowed types.
	tests = []Test{
		{
			node: &ast.Struct{Type: ast.UnionType},
			want: []string{},
		},
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "MyUnion", Type: ast.UnionType},
			}},
			node: ast.TypeReference{Name: "MyUnion"},
			want: []string{},
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

	check = checks.CheckTypes([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{})
	RunTests(t, &check, tests)
}
