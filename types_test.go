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

package thriftcheck_test

import (
	"slices"
	"testing"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

var allTypesMap = map[string]ast.Node{
	"bool":      ast.BaseType{ID: ast.BoolTypeID},
	"i8":        ast.BaseType{ID: ast.I8TypeID},
	"i16":       ast.BaseType{ID: ast.I16TypeID},
	"i32":       ast.BaseType{ID: ast.I32TypeID},
	"i64":       ast.BaseType{ID: ast.I64TypeID},
	"double":    ast.BaseType{ID: ast.DoubleTypeID},
	"string":    ast.BaseType{ID: ast.StringTypeID},
	"binary":    ast.BaseType{ID: ast.BinaryTypeID},
	"list":      ast.ListType{},
	"map":       ast.MapType{},
	"set":       ast.SetType{},
	"enum":      &ast.Enum{},
	"union":     &ast.Struct{Type: ast.UnionType},
	"struct":    &ast.Struct{Type: ast.StructType},
	"exception": &ast.Struct{Type: ast.ExceptionType},
}

func TestThriftTypeUnmarshalString(t *testing.T) {
	for name := range allTypesMap {
		var thriftType thriftcheck.ThriftType
		if err := thriftType.UnmarshalString(name); err != nil {
			t.Error(err)
		}
	}

	for _, name := range []string{"", "invalid", "BOOL"} {
		var thriftType thriftcheck.ThriftType
		if err := thriftType.UnmarshalString(name); err == nil {
			t.Errorf("%s: expected err, got: %v", name, thriftType)
		}
	}
}

func TestMatchTypes(t *testing.T) {
	c := &thriftcheck.C{}

	for name, node := range allTypesMap {
		var thriftType thriftcheck.ThriftType
		if err := thriftType.UnmarshalString(name); err != nil {
			t.Error(err)
		}

		if !thriftType.Matches(c, node) {
			t.Errorf("%s: expected to match %v", name, node)
		}

		for otherName, otherNode := range allTypesMap {
			if otherName == name {
				continue
			}
			if thriftType.Matches(c, otherNode) {
				t.Errorf("%s: expected to not match %v", name, otherNode)
			}
		}
	}
}

func TestMatchBaseTypes(t *testing.T) {
	c := &thriftcheck.C{}

	var baseType thriftcheck.ThriftType
	if err := baseType.UnmarshalString("base"); err != nil {
		t.Error(err)
	}

	baseTypeNames := []string{"bool", "i8", "i16", "i32", "i64", "double", "string", "binary"}

	for _, name := range baseTypeNames {
		node := allTypesMap[name]
		if !baseType.Matches(c, node) {
			t.Errorf("base: expected to match %v", node)
		}

		for otherName, otherNode := range allTypesMap {
			if slices.Contains(baseTypeNames, otherName) {
				continue
			}
			if baseType.Matches(c, otherNode) {
				t.Errorf("base: expected to not match %v", otherNode)
			}
		}
	}
}

func TestMatchTypeReferences(t *testing.T) {
	tests := []struct {
		name string
		def  ast.Definition
	}{
		{name: "enum", def: &ast.Enum{Name: "Enum"}},
		{name: "struct", def: &ast.Struct{Name: "Struct", Type: ast.StructType}},
		{name: "string", def: &ast.Typedef{Name: "String", Type: ast.BaseType{ID: ast.StringTypeID}}},
	}

	for _, tt := range tests {
		var thriftType thriftcheck.ThriftType
		if err := thriftType.UnmarshalString(tt.name); err != nil {
			t.Error(err)
		}

		c := &thriftcheck.C{Program: &ast.Program{Definitions: []ast.Definition{tt.def}}}
		n := ast.TypeReference{Name: tt.def.Info().Name}

		if !thriftType.Matches(c, n) {
			t.Errorf("%s: expected to match %v", tt.name, n)
		}
	}
}
