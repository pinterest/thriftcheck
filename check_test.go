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

func TestNewCheck(t *testing.T) {
	tests := []struct {
		fn     any
		panics bool
	}{
		{"", true},
		{nil, true},
		{func() {}, true},
		{func(c *C) {}, true},
		{func(c *C, n ast.Node) {}, false},
	}

	shouldPanic := func(f func()) {
		defer func() { _ = recover() }()
		f()
		t.Errorf("should have panicked")
	}

	for _, tt := range tests {
		if tt.panics {
			shouldPanic(func() { NewCheck("", tt.fn) })
		} else {
			NewCheck("", tt.fn)
		}
	}
}

func TestCall(t *testing.T) {
	nodes := []ast.Node{
		&ast.Field{},
		&ast.Struct{},
		&ast.Program{},
	}

	called := Checks{
		NewCheck("", func(c *C, n ast.Node) {}),
		NewCheck("", func(c *C, parent, n ast.Node) {}),
		NewCheck("", func(c *C, n1, n2, n3 ast.Node) {}),
		NewCheck("", func(c *C, s *ast.Struct, n ast.Node) {}),
		NewCheck("", func(c *C, s *ast.Struct, f *ast.Field) {}),
	}
	for _, check := range called {
		if !check.Call(&C{}, nodes...) {
			t.Errorf("expected call: %#v", check.fn)
		}
	}

	notcalled := Checks{
		NewCheck("", func(c *C, n1, n2, n3, n4 ast.Node) {}),
		NewCheck("", func(c *C, s *ast.Program, n ast.Node) {}),
		NewCheck("", func(c *C, s *ast.Program, f *ast.Field) {}),
	}
	for _, check := range notcalled {
		if check.Call(&C{}, nodes...) {
			t.Errorf("unexpected call: %#v", check.fn)
		}
	}
}

func TestFilters(t *testing.T) {
	checks := Checks{
		NewCheck("a", func(c *C, n ast.Node) {}),
		NewCheck("a.b", func(c *C, n ast.Node) {}),
		NewCheck("c", func(c *C, n ast.Node) {}),
	}
	tests := []struct {
		prefixes []string
		with     []string
		without  []string
	}{
		{[]string{"a"}, []string{"a", "a.b"}, []string{"c"}},
		{[]string{"a.b"}, []string{"a.b"}, []string{"a", "c"}},
		{[]string{"c"}, []string{"c"}, []string{"a", "a.b"}},
		{[]string{"d"}, []string{}, []string{"a", "a.b", "c"}},
		{[]string{"a", "c"}, []string{"a", "a.b", "c"}, []string{}},
		{[]string{"a", "a.b"}, []string{"a", "a.b"}, []string{"c"}},
	}

	for _, tt := range tests {
		keys := checks.SortedNames()

		w := checks.With(tt.prefixes).SortedNames()
		if !reflect.DeepEqual(w, tt.with) {
			t.Errorf("%s with %s expected %s, got %s", keys, tt.prefixes, tt.with, w)
		}

		wo := checks.Without(tt.prefixes).SortedNames()
		if !reflect.DeepEqual(wo, tt.without) {
			t.Errorf("%s without %s expected %s, got %s", keys, tt.prefixes, tt.without, wo)
		}
	}
}

func TestC(t *testing.T) {
	c := &C{Filename: "test.thrift", Check: "check"}
	node := &ast.Struct{}
	c.Warningf(node, "Warning")
	c.Errorf(node, "Error")

	expected := Messages{
		Message{Filename: c.Filename, Node: node, Check: c.Check, Severity: Warning, Message: "Warning"},
		Message{Filename: c.Filename, Node: node, Check: c.Check, Severity: Error, Message: "Error"},
	}

	if !reflect.DeepEqual(expected, c.Messages) {
		t.Errorf("expected %s, got %s", expected, c.Messages)
	}
}
