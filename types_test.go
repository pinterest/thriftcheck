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

package thriftcheck

import (
	"strings"
	"testing"

	"go.uber.org/thriftrw/ast"
)

func parseTypes(typeNames []string) ([]ThriftType, error) {
	types := make([]ThriftType, 0, len(typeNames))
	for _, name := range typeNames {
		var thriftType ThriftType
		if err := thriftType.UnmarshalString(name); err != nil {
			return nil, err
		}
		types = append(types, thriftType)
	}
	return types, nil
}

func TestParseTypes(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectError   bool
		expectedCount int
	}{
		{
			name:          "valid collection types",
			input:         []string{"map", "list", "set"},
			expectedCount: 3,
		},
		{
			name:          "valid primitive types",
			input:         []string{"bool", "i32", "string"},
			expectedCount: 3,
		},
		{
			name:          "valid structure types",
			input:         []string{"union", "struct", "exception"},
			expectedCount: 3,
		},
		{
			name: "all valid types",
			input: []string{
				"map", "list", "set",
				"bool", "i8", "i16", "i32", "i64", "double", "string", "binary",
				"union", "struct", "exception",
				"enum",
			},
			expectedCount: 15,
		},
		{
			name:        "invalid type",
			input:       []string{"invalid"},
			expectError: true,
		},
		{
			name:        "mixed valid and invalid",
			input:       []string{"map", "invalid", "string"},
			expectError: true,
		},
		{
			name:          "empty input",
			input:         []string{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types, err := parseTypes(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %v", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(types) != tt.expectedCount {
				t.Errorf("expected %d types, got %d", tt.expectedCount, len(types))
			}
		})
	}
}

func TestTypeMatchers_Functionality(t *testing.T) {
	c := &C{}

	tests := []struct {
		name     string
		typeName string
		astType  ast.Node
		matches  bool
	}{
		// Collection types
		{"map matches MapType", "map", ast.MapType{}, true},
		{"list matches ListType", "list", ast.ListType{}, true},
		{"set matches SetType", "set", ast.SetType{}, true},
		{"map doesn't match ListType", "map", ast.ListType{}, false},

		// Primitive types
		{"i32 matches I32", "i32", ast.BaseType{ID: ast.I32TypeID}, true},
		{"string matches String", "string", ast.BaseType{ID: ast.StringTypeID}, true},
		{"bool matches Bool", "bool", ast.BaseType{ID: ast.BoolTypeID}, true},
		{"i32 doesn't match String", "i32", ast.BaseType{ID: ast.StringTypeID}, false},

		// Structure types - Direct *ast.Struct
		{"union matches direct Union struct", "union", &ast.Struct{Type: ast.UnionType}, true},
		{"struct matches direct Struct", "struct", &ast.Struct{Type: ast.StructType}, true},
		{"exception matches direct Exception", "exception", &ast.Struct{Type: ast.ExceptionType}, true},
		{"union doesn't match Struct", "union", &ast.Struct{Type: ast.StructType}, false},
		{"struct doesn't match Union", "struct", &ast.Struct{Type: ast.UnionType}, false},
		{"exception doesn't match Union", "exception", &ast.Struct{Type: ast.UnionType}, false},

		// Other
		{"enum matches Enum", "enum", &ast.Enum{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types, err := parseTypes([]string{tt.typeName})
			if err != nil {
				t.Fatal(err)
			}

			if len(types) != 1 {
				t.Fatalf("expected 1 matcher, got %d", len(types))
			}

			if types[0].Matches(c, tt.astType) != tt.matches {
				t.Errorf("%s: expected %v", tt.name, tt.matches)
			}
		})
	}
}

func TestParseTypes_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		expectedInMsg []string
	}{
		{
			name:  "invalid type error message",
			input: []string{"invalid"},
			expectedInMsg: []string{
				"unknown type: invalid",
				"valid types are:",
				"map",
				"union",
			},
		},
		{
			name:  "unknown type error message",
			input: []string{"unknown"},
			expectedInMsg: []string{
				"unknown type: unknown",
				"valid types are:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTypes(tt.input)
			if err == nil {
				t.Fatal("expected error")
			}

			errMsg := err.Error()
			for _, expected := range tt.expectedInMsg {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("error message should contain %q, got: %s", expected, errMsg)
				}
			}
		})
	}
}
