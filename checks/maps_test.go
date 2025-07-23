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

	"github.com/pinterest/thriftcheck"
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

func TestCheckMapValueType(t *testing.T) {
	// Test with no restrictions - should pass all
	tests := []Test{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.I32TypeID}},
			want: []string{},
		},
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
	}

	check := checks.CheckMapValueType([]thriftcheck.ThriftType{})
	RunTests(t, &check, tests)

	// Test with i32 restriction
	testsI32 := []Test{
		{
			// Should fail for i32 value
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.I32TypeID}},
			want: []string{
				`t.thrift:0:1: error: map value type i32 is restricted (map.value.restricted)`,
			},
		},
		{
			// Should pass for string value
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
	}

	var i32Type thriftcheck.ThriftType
	if err := i32Type.UnmarshalString("i32"); err != nil {
		t.Fatalf("Failed to unmarshal i32 type: %v", err)
	}
	checkI32 := checks.CheckMapValueType([]thriftcheck.ThriftType{i32Type})
	RunTests(t, &checkI32, testsI32)

	// Test with map restriction
	testsMap := []Test{
		{
			// Should fail for nested map
			node: ast.MapType{
				KeyType: ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.I64TypeID},
					ValueType: ast.BaseType{ID: ast.StringTypeID}}},
			want: []string{
				`t.thrift:0:1: error: map value type map is restricted (map.value.restricted)`,
			},
		},
		{
			// Should fail for TypeRef that resolves to map
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{
					Name: "MapType",
					Type: ast.MapType{
						KeyType:   ast.BaseType{ID: ast.I64TypeID},
						ValueType: ast.BaseType{ID: ast.StringTypeID}},
				},
			}},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "MapType"}},
			want: []string{
				`t.thrift:0:1: error: map value type map is restricted (map.value.restricted)`,
			},
		},
		{
			// Should pass for string value
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
	}

	var mapType thriftcheck.ThriftType
	if err := mapType.UnmarshalString("map"); err != nil {
		t.Fatalf("Failed to unmarshal map type: %v", err)
	}
	checkMap := checks.CheckMapValueType([]thriftcheck.ThriftType{mapType})
	RunTests(t, &checkMap, testsMap)

	// Test with union restriction
	testsUnion := []Test{
		{
			// Should fail for union value
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "TestUnion", Type: ast.UnionType},
			}},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "TestUnion"}},
			want: []string{
				`t.thrift:0:1: error: map value type union is restricted (map.value.restricted)`,
			},
		},
		{
			// Should pass for struct value
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "TestStruct"},
			}},
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.TypeReference{Name: "TestStruct"}},
			want: []string{},
		},
	}

	var unionType thriftcheck.ThriftType
	if err := unionType.UnmarshalString("union"); err != nil {
		t.Fatalf("Failed to unmarshal union type: %v", err)
	}
	checkUnion := checks.CheckMapValueType([]thriftcheck.ThriftType{unionType})
	RunTests(t, &checkUnion, testsUnion)
}
