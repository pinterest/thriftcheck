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
	"fmt"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// TypeMatcher represents a way to match against AST types.
type TypeMatcher interface {
	Matches(c *thriftcheck.C, t ast.Type) bool
	Name() string
}

// directTypeMatcher matches direct AST types (map, list, set).
type directTypeMatcher struct {
	name    string
	matchFn func(ast.Type) bool
}

func (m *directTypeMatcher) Matches(c *thriftcheck.C, t ast.Type) bool {
	// Resolve TypeReference if needed.
	if typeRef, ok := t.(ast.TypeReference); ok {
		if resolved := c.ResolveType(typeRef); resolved != nil {
			if resolvedType, ok := resolved.(ast.Type); ok {
				t = resolvedType
			} else {
				return false
			}
		} else {
			return false // Unresolved type, can't match.
		}
	}
	return m.matchFn(t)
}

func (m *directTypeMatcher) Name() string {
	return m.name
}

// primitiveTypeMatcher matches specific primitive types by BaseTypeID.
type primitiveTypeMatcher struct {
	name   string
	typeID ast.BaseTypeID
}

func (m *primitiveTypeMatcher) Matches(c *thriftcheck.C, t ast.Type) bool {
	if typeRef, ok := t.(ast.TypeReference); ok {
		if resolved := c.ResolveType(typeRef); resolved != nil {
			if resolvedType, ok := resolved.(ast.Type); ok {
				t = resolvedType
			} else {
				return false
			}
		} else {
			return false
		}
	}
	if baseType, ok := t.(ast.BaseType); ok {
		return baseType.ID == m.typeID
	}
	return false
}

func (m *primitiveTypeMatcher) Name() string {
	return m.name
}

// structureTypeMatcher matches struct-like types (union, struct, exception).
type structureTypeMatcher struct {
	name       string
	structType ast.StructureType
}

func (m *structureTypeMatcher) Matches(c *thriftcheck.C, t ast.Type) bool {
	// Struct/union/exception types are always referenced via TypeReference.
	typeRef, ok := t.(ast.TypeReference)
	if !ok {
		return false
	}
	resolved := c.ResolveType(typeRef)
	if resolved == nil {
		return false
	}
	if structDef, ok := resolved.(*ast.Struct); ok {
		return structDef.Type == m.structType
	}

	return false
}

func (m *structureTypeMatcher) Name() string {
	return m.name
}

// ParseTypes converts TOML string configuration to TypeMatcher slice.
func ParseTypes(typeNames []string) ([]TypeMatcher, error) {
	matchers := make([]TypeMatcher, 0, len(typeNames))

	typeMap := map[string]TypeMatcher{
		// Collection types
		"map": &directTypeMatcher{
			name: "map",
			matchFn: func(t ast.Type) bool {
				_, ok := t.(ast.MapType)
				return ok
			},
		},
		"list": &directTypeMatcher{
			name: "list",
			matchFn: func(t ast.Type) bool {
				_, ok := t.(ast.ListType)
				return ok
			},
		},
		"set": &directTypeMatcher{
			name: "set",
			matchFn: func(t ast.Type) bool {
				_, ok := t.(ast.SetType)
				return ok
			},
		},

		// Primitive types with specific BaseTypeID matching.
		"bool":   &primitiveTypeMatcher{name: "bool", typeID: ast.BoolTypeID},
		"i8":     &primitiveTypeMatcher{name: "i8", typeID: ast.I8TypeID},
		"i16":    &primitiveTypeMatcher{name: "i16", typeID: ast.I16TypeID},
		"i32":    &primitiveTypeMatcher{name: "i32", typeID: ast.I32TypeID},
		"i64":    &primitiveTypeMatcher{name: "i64", typeID: ast.I64TypeID},
		"double": &primitiveTypeMatcher{name: "double", typeID: ast.DoubleTypeID},
		"string": &primitiveTypeMatcher{name: "string", typeID: ast.StringTypeID},
		"binary": &primitiveTypeMatcher{name: "binary", typeID: ast.BinaryTypeID},

		// Structure types with specific StructureType matching.
		"union":     &structureTypeMatcher{name: "union", structType: ast.UnionType},
		"struct":    &structureTypeMatcher{name: "struct", structType: ast.StructType},
		"exception": &structureTypeMatcher{name: "exception", structType: ast.ExceptionType},
	}

	for _, typeName := range typeNames {
		matcher, ok := typeMap[typeName]
		if !ok {
			validTypes := make([]string, 0, len(typeMap))
			for k := range typeMap {
				validTypes = append(validTypes, k)
			}
			return nil, fmt.Errorf("invalid type %q, valid types are: %v", typeName, validTypes)
		}
		matchers = append(matchers, matcher)
	}

	return matchers, nil
}
