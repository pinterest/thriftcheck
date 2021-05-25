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
	"reflect"

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

// Annotations returns an ast.Node's Annotations.
func Annotations(node ast.Node) []*ast.Annotation {
	if v := reflect.ValueOf(node); v.Kind() == reflect.Ptr {
		if f := v.Elem().FieldByName("Annotations"); f.IsValid() {
			return f.Interface().([]*ast.Annotation)
		}
	}
	return nil
}
