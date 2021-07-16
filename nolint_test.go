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
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestNolint(t *testing.T) {
	tests := []struct {
		desc  string
		node  ast.Node
		ok    bool
		names []string
	}{
		{
			desc:  "empty",
			node:  &ast.Struct{},
			ok:    false,
			names: nil,
		},
		{
			desc:  "words",
			node:  &ast.Struct{Doc: "Just some words"},
			ok:    false,
			names: nil,
		},
		{
			desc: `(nolint = "")`,
			node: &ast.Struct{
				Annotations: []*ast.Annotation{{Name: "nolint", Value: ""}},
			},
			ok:    true,
			names: nil,
		},
		{
			desc: `(nolint = "a")`,
			node: &ast.Struct{
				Annotations: []*ast.Annotation{{Name: "nolint", Value: "a"}},
			},
			ok:    true,
			names: []string{"a"},
		},
		{
			desc: `(nolint = "a,b")`,
			node: &ast.Struct{
				Annotations: []*ast.Annotation{{Name: "nolint", Value: "a,b"}},
			},
			ok:    true,
			names: []string{"a", "b"},
		},
		{
			desc:  "@nolint",
			node:  &ast.Struct{Doc: "@nolint"},
			ok:    true,
			names: nil,
		},
		{
			desc:  "@nolint()",
			node:  &ast.Struct{Doc: "@nolint()"},
			ok:    true,
			names: nil,
		},
		{
			desc:  "@nolint(a)",
			node:  &ast.Struct{Doc: "@nolint(a)"},
			ok:    true,
			names: []string{"a"},
		},
		{
			desc:  "@nolint(a,b)",
			node:  &ast.Struct{Doc: "@nolint(a,b)"},
			ok:    true,
			names: []string{"a", "b"},
		},
		{
			desc: `(nolint = "a,b") and @nolint(b,c)`,
			node: &ast.Struct{
				Annotations: []*ast.Annotation{{Name: "nolint", Value: "a,b"}},
				Doc:         "@nolint(b,c)",
			},
			ok:    true,
			names: []string{"a", "b", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			names, ok := nolint(tt.node)
			if ok != tt.ok {
				t.Errorf("expected %v, got %v", tt.ok, ok)
			} else if !reflect.DeepEqual(names, tt.names) {
				t.Errorf("expected %v, got %v", tt.names, names)
			}
		})
	}
}
