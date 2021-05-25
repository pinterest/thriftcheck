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

package checks_test

import (
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckMapKeyType(t *testing.T) {
	tests := []struct {
		node ast.MapType
		want []string
	}{
		{
			node: ast.MapType{
				KeyType:   ast.BaseType{ID: ast.StringTypeID},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{},
		},
		{
			node: ast.MapType{
				KeyType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.StringTypeID},
					ValueType: ast.BaseType{ID: ast.StringTypeID}},
				ValueType: ast.BaseType{ID: ast.StringTypeID}},
			want: []string{
				`t.thrift:0:1:error: map key must be a primitive type (map.key.type)`,
			},
		},
	}

	check := checks.CheckMapKeyType()

	for _, tt := range tests {
		c := newC(&check)
		check.Call(c, tt.node)
		assertMessageStrings(t, tt.node, tt.want, c.Messages)
	}
}
