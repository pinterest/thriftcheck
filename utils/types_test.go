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

package utils

import (
	"strings"
	"testing"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func TestParseTypes(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expectError bool
		expectedCount int
	}{
		{
			name: "valid collection types",
			input: []string{"map", "list", "set"},
			expectedCount: 3,
		},
		{
			name: "valid primitive types",
			input: []string{"bool", "i32", "string"},
			expectedCount: 3,
		},
		{
			name: "valid structure types",
			input: []string{"union", "struct", "exception"},
			expectedCount: 3,
		},
		{
			name: "all valid types",
			input: []string{
				"map", "list", "set",
				"bool", "i8", "i16", "i32", "i64", "double", "string", "binary",
				"union", "struct", "exception",
			},
			expectedCount: 14,
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
			matchers, err := ParseTypes(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %v", tt.input)
				}
				return
			}
			
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if len(matchers) != tt.expectedCount {
				t.Errorf("expected %d matchers, got %d", tt.expectedCount, len(matchers))
			}
		})
	}
}

func TestTypeMatchers_Functionality(t *testing.T) {
	c := &thriftcheck.C{}
	
	tests := []struct {
		name     string
		typeName string
		astType  ast.Type
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
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchers, err := ParseTypes([]string{tt.typeName})
			if err != nil {
				t.Fatal(err)
			}
			
			if len(matchers) != 1 {
				t.Fatalf("expected 1 matcher, got %d", len(matchers))
			}
			
			if matchers[0].Matches(c, tt.astType) != tt.matches {
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
				"invalid type \"invalid\"",
				"valid types are:",
				"map",
				"union",
			},
		},
		{
			name:  "unknown type error message",
			input: []string{"unknown"},
			expectedInMsg: []string{
				"invalid type \"unknown\"",
				"valid types are:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTypes(tt.input)
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
