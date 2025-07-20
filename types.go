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
	"sort"
	"strings"

	"go.uber.org/thriftrw/ast"
)

type typeMatcher interface {
	Matches(c *C, n ast.Node) bool
}

// ThriftType implements fig StringUnmarshaler for automatic toml parsing.
type ThriftType struct {
	name    string
	matcher typeMatcher
}

// UnmarshalString implements fig.StringUnmarshaler for automatic toml parsing.
func (t *ThriftType) UnmarshalString(v string) error {
	name := strings.ToLower(v)
	matcher, ok := typeMatchers[name]
	if !ok {
		validTypes := make([]string, 0, len(typeMatchers))
		for k := range typeMatchers {
			validTypes = append(validTypes, k)
		}
		sort.Strings(validTypes)
		return fmt.Errorf("unknown type: %s, valid types are: %v", v, validTypes)
	}

	t.name = name
	t.matcher = matcher
	return nil
}

func (t ThriftType) Matches(c *C, n ast.Node) bool {
	return t.matcher.Matches(c, n)
}

func (t ThriftType) String() string {
	return t.name
}

// thriftTypeMatcher handles all types with unified TypeReference resolution.
type thriftTypeMatcher struct {
	matches func(ast.Node) bool
}

func (m *thriftTypeMatcher) Matches(c *C, n ast.Node) bool {
	if typeRef, ok := n.(ast.TypeReference); ok {
		if resolved := c.ResolveType(typeRef); resolved != nil {
			if resolvedType, ok := resolved.(ast.Type); ok {
				n = resolvedType
			} else {
				return false
			}
		} else {
			return false
		}
	}
	return m.matches(n)
}

// structureTypeMatcher matches struct-like types (union, struct, exception).
type structureTypeMatcher struct {
	structType ast.StructureType
}

func (m *structureTypeMatcher) Matches(c *C, n ast.Node) bool {
	// Handle direct *ast.Struct types (when used directly)
	if structDef, ok := n.(*ast.Struct); ok {
		return structDef.Type == m.structType
	}

	// Handle TypeReference (struct/union/exception referenced by name)
	typeRef, ok := n.(ast.TypeReference)
	if !ok {
		return false
	}

	// Resolve the reference to get the actual definition
	resolved := c.ResolveType(typeRef)
	if resolved == nil {
		return false
	}

	if structDef, ok := resolved.(*ast.Struct); ok {
		return structDef.Type == m.structType
	}

	return false
}

var typeMatchers = map[string]typeMatcher{
	// Collection types
	"map":  &thriftTypeMatcher{func(n ast.Node) bool { _, ok := n.(ast.MapType); return ok }},
	"list": &thriftTypeMatcher{func(n ast.Node) bool { _, ok := n.(ast.ListType); return ok }},
	"set":  &thriftTypeMatcher{func(n ast.Node) bool { _, ok := n.(ast.SetType); return ok }},

	// Primitive types
	"bool":   &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.BoolTypeID) }},
	"i8":     &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.I8TypeID) }},
	"i16":    &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.I16TypeID) }},
	"i32":    &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.I32TypeID) }},
	"i64":    &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.I64TypeID) }},
	"double": &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.DoubleTypeID) }},
	"string": &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.StringTypeID) }},
	"binary": &thriftTypeMatcher{func(n ast.Node) bool { return matchBaseType(n, ast.BinaryTypeID) }},

	// Structure types
	"union":     &structureTypeMatcher{ast.UnionType},
	"struct":    &structureTypeMatcher{ast.StructType},
	"exception": &structureTypeMatcher{ast.ExceptionType},
}

// Helper function for primitive type matching.
func matchBaseType(n ast.Node, expectedID ast.BaseTypeID) bool {
	if baseType, ok := n.(ast.BaseType); ok {
		return baseType.ID == expectedID
	}
	return false
}
