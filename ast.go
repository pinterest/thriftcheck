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

package thriftcheck

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"go.uber.org/thriftrw/ast"
)

var nodeInterface = reflect.TypeOf((*ast.Node)(nil)).Elem()

// VisitorFunc adapts a function to the ast.Visitor interface. This differs
// from ast.VisitorFunc in that is supports an ast.Visitor-compativle return
// value.
type VisitorFunc func(ast.Walker, ast.Node) VisitorFunc

// Visit the given node and its descendants.
func (f VisitorFunc) Visit(w ast.Walker, n ast.Node) ast.Visitor {
	if f != nil {
		return f(w, n)
	}
	return nil
}

// Doc returns an ast.Node's Doc string.
func Doc(node ast.Node) string {
	if v := reflect.ValueOf(node); v.Kind() == reflect.Ptr {
		if f := v.Elem().FieldByName("Doc"); f.IsValid() {
			return f.Interface().(string)
		}
	}
	return ""
}

// Resolve resolves a named reference to its target node.
//
// The target can either be in the current program's scope or it can refer to
// an included file using dot notation. Included files must exist in one of the
// given search directories.
func Resolve(name string, program *ast.Program, dirs []string) (ast.Node, error) {
	defs := program.Definitions

	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		fname := parts[0] + ".thrift"

		var ipath string
		for _, header := range program.Headers {
			if include, ok := header.(*ast.Include); ok {
				if _, file := filepath.Split(include.Path); file == fname {
					ipath = include.Path
					break
				}
			}
		}
		if ipath == "" {
			return nil, fmt.Errorf("missing \"include\" for type reference %q", name)
		}

		program, _, err := ParseFile(ipath, dirs)
		if err != nil {
			return nil, err
		}

		defs = program.Definitions
		name = parts[1]
	}

	for _, def := range defs {
		if def.Info().Name == name {
			return def, nil
		}
	}

	return nil, fmt.Errorf("%q could not be resolved", name)
}

// ResolveConstant resolves an ast.ConstantReference to its target node.
//
// The following name formats are supported:
// 	- "Constant" (ast.Constant)
// 	- "Enum.Value" (ast.EnumItem)
// 	- "include.Constant" (ast.Constant)
// 	- "include.Enum.Value" (ast.EnumItem)
func ResolveConstant(ref ast.ConstantReference, program *ast.Program, dirs []string) (ast.Node, error) {
	parts := strings.SplitN(ref.Name, ".", 3)

	n, err := Resolve(parts[0], program, dirs)
	if err != nil && len(parts) > 1 {
		n, err = Resolve(parts[0]+"."+parts[1], program, dirs)
	}
	if err != nil {
		return n, fmt.Errorf("%q could not be resolved", ref.Name)
	}

	if e, ok := n.(*ast.Enum); ok {
		for _, ei := range e.Items {
			if ei.Name == parts[len(parts)-1] {
				return ei, nil
			}
		}
		return nil, fmt.Errorf("enum value %q could not be resolved", ref.Name)
	}

	return n, nil
}

// ResolveType calls Resolve and goes one step further by attempting to
// resolve the target node's own type. This is useful when the reference
// points to an ast.Typedef or ast.Constant, for example, and the caller
// is primarily intererested in the target's ast.Type.
func ResolveType(ref ast.TypeReference, program *ast.Program, dirs []string) (ast.Node, error) {
	n, err := Resolve(ref.Name, program, dirs)
	if err != nil {
		return nil, err
	}

	switch t := n.(type) {
	case *ast.Constant:
		return t.Type, nil

	case *ast.Typedef:
		return t.Type, nil

	default:
		return n, nil
	}
}
