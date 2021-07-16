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

func TestAnnotations(t *testing.T) {
	annotations := []*ast.Annotation{
		{Name: "test1", Value: "value1"},
		{Name: "test2", Value: "value2"},
	}
	tests := []struct {
		node ast.Node
		want []*ast.Annotation
	}{
		{&ast.Struct{}, nil},
		{&ast.Struct{Annotations: annotations}, annotations},
		{&ast.Struct{Annotations: annotations[1:]}, annotations[1:]},
		{&ast.Annotation{}, nil},
	}

	for _, tt := range tests {
		got := Annotations(tt.node)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("expected %s but got %s", tt.want, got)
		}
	}
}

func TestDoc(t *testing.T) {
	tests := []struct {
		node ast.Node
		want string
	}{
		{&ast.Struct{}, ""},
		{&ast.Struct{Doc: ""}, ""},
		{&ast.Struct{Doc: "String"}, "String"},
	}

	for _, tt := range tests {
		got := Doc(tt.node)
		if got != tt.want {
			t.Errorf("expected %s but got %s", tt.want, got)
		}
	}
}
