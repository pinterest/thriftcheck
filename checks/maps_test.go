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

func TestCheckMapKeyType(t *testing.T) {
	i8Type := ParseType(t, "i8")
	enumType := ParseType(t, "enum")
	stringType := ParseType(t, "string")

	// Test with only allowed types.
	tests := []Test{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.I8TypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
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
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{`t.thrift:0:1: error: map key type "string" is not allowed (map.key.type)`},
		},
	}

	check := checks.CheckMapKeyType([]thriftcheck.ThriftType{i8Type, enumType}, []thriftcheck.ThriftType{})
	RunTests(t, &check, tests)

	// Test with only disallowed types.
	tests = []Test{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.I8TypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{`t.thrift:0:1: error: map key type "i8" is not allowed (map.key.type)`},
		},
		{
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Enum{Name: "Enum"},
			}},
			node: ast.MapType{
				KeyType:   ast.TypeReference{Name: "Enum"},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{`t.thrift:0:1: error: map key type "enum" is not allowed (map.key.type)`},
		},
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
	}

	check = checks.CheckMapKeyType([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{i8Type, enumType})
	RunTests(t, &check, tests)

	// Test with both allowed and disallowed types.
	tests = []Test{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.I8TypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			// Disallowances have precedence over allowances.
			want: []string{`t.thrift:0:1: error: map key type "i8" is not allowed (map.key.type)`},
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
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{`t.thrift:0:1: error: map key type "string" is not allowed (map.key.type)`},
		},
	}

	check = checks.CheckMapKeyType(
		[]thriftcheck.ThriftType{i8Type, enumType},
		[]thriftcheck.ThriftType{i8Type, stringType})
	RunTests(t, &check, tests)
}

func TestCheckMapValueType(t *testing.T) {
	// Test with no disallowed types - should pass all
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

	check := checks.CheckMapValueType([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{})
	RunTests(t, &check, tests)

	// Test with i32 disallowed
	testsI32 := []Test{
		{
			// Should fail for i32 value
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.I32TypeID}},
			want: []string{
				`t.thrift:0:1: error: map value type "i32" is not allowed (map.value.type)`,
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

	i32Type := ParseType(t, "i32")
	checkI32 := checks.CheckMapValueType([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{i32Type})
	RunTests(t, &checkI32, testsI32)

	// Test with map disallowed
	testsMap := []Test{
		{
			// Should fail for nested map
			node: ast.MapType{
				KeyType: ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.I64TypeID},
					ValueType: ast.BaseType{ID: ast.StringTypeID}}},
			want: []string{
				`t.thrift:0:1: error: map value type "map" is not allowed (map.value.type)`,
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
				`t.thrift:0:1: error: map value type "map" is not allowed (map.value.type)`,
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

	mapType := ParseType(t, "map")
	checkMap := checks.CheckMapValueType([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{mapType})
	RunTests(t, &checkMap, testsMap)

	// Test with union disallowed
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
				`t.thrift:0:1: error: map value type "union" is not allowed (map.value.type)`,
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

	unionType := ParseType(t, "union")
	checkUnion := checks.CheckMapValueType([]thriftcheck.ThriftType{}, []thriftcheck.ThriftType{unionType})
	RunTests(t, &checkUnion, testsUnion)
}
