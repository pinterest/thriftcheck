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

// TypeMatcher represents a way to match against AST nodes
type TypeMatcher interface {
	Matches(c *C, n ast.Node) bool
	Name() string
}

// ThriftType implements fig StringUnmarshaler for automatic toml parsing
type ThriftType struct {
	name    string
	matcher TypeMatcher
}

// UnmarshalString implements fig.StringUnmarshaler for automatic toml parsing
func (t *ThriftType) UnmarshalString(v string) error {
	name := strings.ToLower(v)
	factory, ok := typeFactories[name]
	if !ok {
		validTypes := make([]string, 0, len(typeFactories))
		for k := range typeFactories {
			validTypes = append(validTypes, k)
		}
		sort.Strings(validTypes)
		return fmt.Errorf("unknown type: %s, valid types are: %v", v, validTypes)
	}

	t.name = name
	t.matcher = factory(name)
	return nil
}

// Matches delegates to the internal matcher
func (t *ThriftType) Matches(c *C, n ast.Node) bool {
	return t.matcher.Matches(c, n)
}

// Name returns the type name
func (t *ThriftType) Name() string {
	return t.name
}

// thriftTypeMatcher handles all types with unified TypeReference resolution
type thriftTypeMatcher struct {
	name    string
	matchFn func(ast.Node) bool
}

func (m *thriftTypeMatcher) Matches(c *C, n ast.Node) bool {
	// Resolve TypeReference if needed
	if typeRef, ok := n.(ast.TypeReference); ok {
		if resolved := c.ResolveType(typeRef); resolved != nil {
			if resolvedType, ok := resolved.(ast.Type); ok {
				n = resolvedType
			} else {
				return false
			}
		} else {
			return false // Unresolved type, can't match
		}
	}
	return m.matchFn(n)
}

func (m *thriftTypeMatcher) Name() string {
	return m.name
}

// structureTypeMatcher matches struct-like types (union, struct, exception)
type structureTypeMatcher struct {
	name       string
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

	// Check if it's a struct-like definition with the right type
	if structDef, ok := resolved.(*ast.Struct); ok {
		return structDef.Type == m.structType
	}

	return false
}

func (m *structureTypeMatcher) Name() string {
	return m.name
}

// Factory functions for creating type matchers
var typeFactories = map[string]func(string) TypeMatcher{
	// Collection types
	"map": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { _, ok := n.(ast.MapType); return ok }}
	},
	"list": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { _, ok := n.(ast.ListType); return ok }}
	},
	"set": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { _, ok := n.(ast.SetType); return ok }}
	},

	// Primitive types
	"bool": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.BoolTypeID) }}
	},
	"i8": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.I8TypeID) }}
	},
	"i16": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.I16TypeID) }}
	},
	"i32": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.I32TypeID) }}
	},
	"i64": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.I64TypeID) }}
	},
	"double": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.DoubleTypeID) }}
	},
	"string": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.StringTypeID) }}
	},
	"binary": func(name string) TypeMatcher {
		return &thriftTypeMatcher{name, func(n ast.Node) bool { return matchBaseType(n, ast.BinaryTypeID) }}
	},

	// Structure types
	"union":     func(name string) TypeMatcher { return &structureTypeMatcher{name, ast.UnionType} },
	"struct":    func(name string) TypeMatcher { return &structureTypeMatcher{name, ast.StructType} },
	"exception": func(name string) TypeMatcher { return &structureTypeMatcher{name, ast.ExceptionType} },
}

// Helper function for primitive type matching
func matchBaseType(n ast.Node, expectedID ast.BaseTypeID) bool {
	if baseType, ok := n.(ast.BaseType); ok {
		return baseType.ID == expectedID
	}
	return false
}
