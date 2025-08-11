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
	"fmt"
	"maps"
	"slices"

	"go.uber.org/thriftrw/ast"
)

type typeMatcher func(ast.Node) bool

// ThriftType implements fig StringUnmarshaler for automatic toml parsing.
type ThriftType struct {
	name    string
	matcher typeMatcher
}

// UnmarshalString implements fig.StringUnmarshaler for automatic toml parsing.
func (t *ThriftType) UnmarshalString(name string) error {
	matcher, ok := typeMatchers[name]
	if !ok {
		validTypes := slices.Sorted(maps.Keys(typeMatchers))
		return fmt.Errorf("unknown type: %s, valid types are: %v", name, validTypes)
	}

	t.name = name
	t.matcher = matcher
	return nil
}

// Matches tests whether the given [ast.Node] matches this Thrift type.
// Type references will be resolved.
func (t ThriftType) Matches(c *C, n ast.Node) bool {
	if ref, ok := n.(ast.TypeReference); ok {
		if n = c.ResolveType(ref); n == nil {
			return false
		}
	}
	return t.matcher(n)
}

func (t ThriftType) String() string {
	return t.name
}

var typeMatchers = map[string]typeMatcher{
	// Base types
	"base":   func(n ast.Node) bool { _, ok := n.(ast.BaseType); return ok },
	"bool":   func(n ast.Node) bool { return matchBaseType(n, ast.BoolTypeID) },
	"i8":     func(n ast.Node) bool { return matchBaseType(n, ast.I8TypeID) },
	"i16":    func(n ast.Node) bool { return matchBaseType(n, ast.I16TypeID) },
	"i32":    func(n ast.Node) bool { return matchBaseType(n, ast.I32TypeID) },
	"i64":    func(n ast.Node) bool { return matchBaseType(n, ast.I64TypeID) },
	"double": func(n ast.Node) bool { return matchBaseType(n, ast.DoubleTypeID) },
	"string": func(n ast.Node) bool { return matchBaseType(n, ast.StringTypeID) },
	"binary": func(n ast.Node) bool { return matchBaseType(n, ast.BinaryTypeID) },

	// Collections
	"list": func(n ast.Node) bool { _, ok := n.(ast.ListType); return ok },
	"map":  func(n ast.Node) bool { _, ok := n.(ast.MapType); return ok },
	"set":  func(n ast.Node) bool { _, ok := n.(ast.SetType); return ok },

	// Definitions
	"enum":      func(n ast.Node) bool { _, ok := n.(*ast.Enum); return ok },
	"union":     func(n ast.Node) bool { return matchStructureType(n, ast.UnionType) },
	"struct":    func(n ast.Node) bool { return matchStructureType(n, ast.StructType) },
	"exception": func(n ast.Node) bool { return matchStructureType(n, ast.ExceptionType) },
}

func matchStructureType(n ast.Node, t ast.StructureType) bool {
	if s, ok := n.(*ast.Struct); ok {
		return s.Type == t
	}

	return false
}

func matchBaseType(n ast.Node, expectedID ast.BaseTypeID) bool {
	if baseType, ok := n.(ast.BaseType); ok {
		return baseType.ID == expectedID
	}
	return false
}
