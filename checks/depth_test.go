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

package checks_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckDepth(t *testing.T) {
	// A max depth of 0 is treated as no max depth. Depth starts at 1.
	maxDepth := 0
	cyclesAllowed := false
	tests := []Test{
		// Struct with extra depth.
		{
			name: "a.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{&ast.Struct{Name: "Something"}}},
			node: &ast.Struct{
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something"}}}},
			want: []string{},
		},
		// Annotation overriding the max depth of 0.
		{
			name: "b.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{&ast.Struct{Name: "Something"}}},
			node: &ast.Struct{
				Type:        ast.StructType,
				Fields:      []*ast.Field{{Type: ast.TypeReference{Name: "Something"}}},
				Annotations: []*ast.Annotation{{Name: "maxDepth", Value: "1"}}},
			want: []string{`b.thrift:0:1: error:  exceeded maximum depth of 1
	b.thrift:0:0 (Something) +1 (2) (depth)`},
		},
		// Cycle.
		{
			name: "c.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "A"},
				&ast.Struct{Name: "B", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "A"}}}}}},
			node: &ast.Struct{
				Name:   "A",
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "B"}}}},
			want: []string{`c.thrift:0:1: error: A led to a cycle
	c.thrift:0:0 (B) +1 (2)
	c.thrift:0:0 (A) +1 (3) (depth)`},
		},
		// Self-loop.
		{
			name: "d.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "A", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "A"}}}}}},
			node: &ast.Struct{
				Name:   "A",
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "A"}}}},
			want: []string{`d.thrift:0:1: error: A led to a cycle
	d.thrift:0:0 (A) +1 (2) (depth)`},
		},
	}

	check := checks.CheckDepth(maxDepth, cyclesAllowed)
	RunTests(t, &check, tests)

	maxDepth = 0
	cyclesAllowed = true
	tests = []Test{
		// Cycle.
		{
			name: "a.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "B", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "A"}}}},
				&ast.Struct{Name: "A"}}},
			node: &ast.Struct{
				Name:   "A",
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "B"}}}},
			want: []string{},
		},
	}

	check = checks.CheckDepth(maxDepth, cyclesAllowed)
	RunTests(t, &check, tests)

	maxDepth = 0
	cyclesAllowed = false
	tests = []Test{
		// Cycle.
		{
			name: "a.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "A"},
				&ast.Struct{Name: "B", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "A"}}}}}},
			node: &ast.Struct{
				Name:   "A",
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "B"}}}},
			want: []string{`a.thrift:0:1: error: A led to a cycle
	a.thrift:0:0 (B) +1 (2)
	a.thrift:0:0 (A) +1 (3) (depth)`},
		},
	}

	check = checks.CheckDepth(maxDepth, cyclesAllowed)
	RunTests(t, &check, tests)

	maxDepth = 1
	cyclesAllowed = true
	tests = []Test{
		// Staying within the depth limit.
		{
			name: "a.thrift",
			node: &ast.Struct{
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.BaseType{ID: ast.BoolTypeID}}}},
			want: []string{},
		},
		// Exceeding the max depth.
		{
			name: "b.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{&ast.Struct{Name: "Something"}}},
			node: &ast.Struct{
				Type:   ast.StructType,
				Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something"}}}},
			want: []string{`b.thrift:0:1: error:  exceeded maximum depth of 1
	b.thrift:0:0 (Something) +1 (2) (depth)`},
		},
	}

	check = checks.CheckDepth(maxDepth, cyclesAllowed)
	RunTests(t, &check, tests)

	maxDepth = 2
	cyclesAllowed = true
	tests = []Test{
		// Having extra depth from a set, but staying within the depth limit.
		{
			name: "a1.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.SetType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}},
			want: []string{},
		},
		// Exceeding the max depth from a nested set.
		{
			name: "a2.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.SetType{ValueType: ast.SetType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}}},
			want: []string{`a2.thrift:0:1: error:  exceeded maximum depth of 2
	a2.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Having extra depth from a list, but staying within the depth limit.
		{
			name: "a3.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.ListType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}},
			want: []string{},
		},
		// Exceeding the max depth from a nested list.
		{
			name: "a4.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.ListType{ValueType: ast.ListType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}}},
			want: []string{`a4.thrift:0:1: error:  exceeded maximum depth of 2
	a4.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Exceeding the max depth mixing a list and a Map.
		{
			name: "a5.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{
				Type: ast.ListType{ValueType: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.BoolTypeID},
					ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}}},
			want: []string{`a5.thrift:0:1: error:  exceeded maximum depth of 2
	a5.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Having extra depth from a map, but staying within the depth limit.
		{
			name: "a6.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.MapType{
					KeyType:   ast.BaseType{ID: ast.BoolTypeID},
					ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}},
			want: []string{},
		},
		// Exceeding the max depth from a nested map.
		{
			name: "a7.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.MapType{
					KeyType: ast.BaseType{ID: ast.BoolTypeID},
					ValueType: ast.MapType{
						KeyType:   ast.BaseType{ID: ast.BoolTypeID},
						ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}}},
			want: []string{`a7.thrift:0:1: error:  exceeded maximum depth of 2
	a7.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Multiple fields.
		{
			name: "a8.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1"},
				&ast.Struct{Name: "Something2"},
				&ast.Struct{Name: "Something3"}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Something1"}},
				{Type: ast.TypeReference{Name: "Something2"}},
				{Type: ast.TypeReference{Name: "Something3"}}}},
			want: []string{},
		},
		// Mltiple fields, with one exceeding the max depth.
		{
			name: "a9.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1"},
				&ast.Struct{Name: "Something2", Fields: []*ast.Field{
					{Type: ast.SetType{ValueType: ast.SetType{ValueType: ast.BaseType{ID: ast.BinaryTypeID}}}}}},
				&ast.Struct{Name: "Something3"}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Something1"}},
				{Type: ast.TypeReference{Name: "Something2"}},
				{Type: ast.TypeReference{Name: "Something3"}}}},
			want: []string{`a9.thrift:0:1: error:  exceeded maximum depth of 2
	a9.thrift:0:0 (Something2) +1 (2)
	a9.thrift:0:0 (binary) +2 (4) (depth)`},
		},
		// Multiple fields exceeding the max depth.
		{
			name: "x1.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something2"}}}},
				&ast.Struct{Name: "Something2", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something3"}}}},
				&ast.Struct{Name: "Something4", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something5"}}}},
				&ast.Struct{Name: "Something5", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something6"}}}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Something1"}},
				{Type: ast.TypeReference{Name: "Something4"}}}},
			want: []string{`x1.thrift:0:1: error:  exceeded maximum depth of 2
	x1.thrift:0:0 (Something1) +1 (2)
	x1.thrift:0:0 (Something2) +1 (3) (depth)`},
		},
		// Struct 'bypassing' the global max depth thanks to an annotation.
		{
			name: "x2.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something2"}}}},
				&ast.Struct{Name: "Something2", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something3"}}}},
				&ast.Struct{Name: "Something4", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something5"}}}},
				&ast.Struct{Name: "Something5", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something6"}}}}}},
			node: &ast.Struct{
				Type:        ast.StructType,
				Annotations: []*ast.Annotation{{Name: "maxDepth", Value: strconv.Itoa(maxDepth + 1)}},
				Fields: []*ast.Field{
					{Type: ast.TypeReference{Name: "Something1"}},
					{Type: ast.TypeReference{Name: "Something4"}}}},
			want: []string{},
		},
		// Typedefs by themselves do not increase depth.
		{
			name: "b1.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something2"}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "Something3"}},
				&ast.Typedef{Name: "Something3", Type: ast.BaseType{ID: ast.BoolTypeID}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{},
		},
		// Exceeding the max depth from a nested list after a series of typedefs.
		{
			name: "b2.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something2"}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "Something3"}},
				&ast.Typedef{Name: "Something3", Type: ast.ListType{ValueType: ast.ListType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{`b2.thrift:0:1: error:  exceeded maximum depth of 2
	b2.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Exceeding the max depth from struct references made through typedefs.
		{
			name: "b3.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "A"}},
				&ast.Struct{Name: "A", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something2"}}}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "B"}},
				&ast.Struct{Name: "B"}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{`b3.thrift:0:1: error:  exceeded maximum depth of 2
	b3.thrift:0:0 (A) +1 (2)
	b3.thrift:0:0 (B) +1 (3) (depth)`},
		},
		// Typedef self-loop.
		{
			name: "b4.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something1"}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{`b4.thrift:0:1: warning: found a cycle resolving typedef "Something1" (depth)`},
		},
		// 2-node typedef cycle.
		{
			name: "b5.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something2"}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "Something1"}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{`b5.thrift:0:1: warning: found a cycle resolving typedef "Something1" (depth)`},
		},
		// 3-node typedef cycle.
		{
			name: "b6.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something2"}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "Something3"}},
				&ast.Typedef{Name: "Something3", Type: ast.TypeReference{Name: "Something1"}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{`b6.thrift:0:1: warning: found a cycle resolving typedef "Something1" (depth)`},
		},
		// Reporting both a typedef cycle warning and a max depth error.
		{
			name: "b7.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Typedef{Name: "Something1", Type: ast.TypeReference{Name: "Something2"}},
				&ast.Typedef{Name: "Something2", Type: ast.TypeReference{Name: "Something1"}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.SetType{ValueType: ast.SetType{ValueType: ast.BaseType{ID: ast.BoolTypeID}}}},
				{Type: ast.TypeReference{Name: "Something1"}}}},
			want: []string{
				`b7.thrift:0:1: warning: found a cycle resolving typedef "Something1" (depth)`,
				`b7.thrift:0:1: error:  exceeded maximum depth of 2
	b7.thrift:0:0 (bool) +2 (3) (depth)`},
		},
		// Exceeding the max depth from a constant.
		{
			name: "b8.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Constant{Name: "Constant", Type: ast.SetType{ValueType: ast.SetType{ValueType: ast.BaseType{ID: ast.I16TypeID}}}}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Constant"}}}},
			want: []string{`b8.thrift:0:1: error:  exceeded maximum depth of 2
	b8.thrift:0:0 (i16) +2 (3) (depth)`},
		},
		// Exceeding the max depth with unions.
		{
			name: "b9.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1", Type: ast.UnionType, Fields: []*ast.Field{
					{Type: ast.TypeReference{Name: "Something3"}},
					{Type: ast.TypeReference{Name: "Something4"}}}},
				&ast.Struct{Name: "Something3", Type: ast.UnionType, Fields: []*ast.Field{
					{Type: ast.TypeReference{Name: "Something5"}},
					{Type: ast.TypeReference{Name: "Something6"}}}}}},
			node: &ast.Struct{Type: ast.UnionType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Something1"}},
				{Type: ast.TypeReference{Name: "Something2"}}}},
			want: []string{`b9.thrift:0:1: error:  exceeded maximum depth of 2
	b9.thrift:0:0 (Something1) +1 (2)
	b9.thrift:0:0 (Something3) +1 (3) (depth)`},
		},
		// Exceeding the max depth with exceptions.
		{
			name: "c1.thrift",
			prog: &ast.Program{Definitions: []ast.Definition{
				&ast.Struct{Name: "Something1", Type: ast.ExceptionType, Fields: []*ast.Field{
					{Type: ast.TypeReference{Name: "Something3"}},
					{Type: ast.TypeReference{Name: "Something4"}}}},
				&ast.Struct{Name: "Something3", Type: ast.ExceptionType, Fields: []*ast.Field{
					{Type: ast.TypeReference{Name: "Something5"}},
					{Type: ast.TypeReference{Name: "Something6"}}}}}},
			node: &ast.Struct{Type: ast.ExceptionType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "Something1"}},
				{Type: ast.TypeReference{Name: "Something2"}}}},
			want: []string{`c1.thrift:0:1: error:  exceeded maximum depth of 2
	c1.thrift:0:0 (Something1) +1 (2)
	c1.thrift:0:0 (Something3) +1 (3) (depth)`},
		},
	}

	check = checks.CheckDepth(maxDepth, cyclesAllowed)
	RunTests(t, &check, tests)

	test_dir := t.TempDir()

	files := []struct {
		name    string
		content string
	}{
		{
			name: fmt.Sprintf("%s/b.thrift", test_dir),
			content: fmt.Sprintf(`
			include "%s/c.thrift"

			struct A {
				1: optional map<string,list<B>> something
				2: optional i16 number
			}

			struct B {
				1: optional bool something
				2: optional c.A anotherField
			}`, test_dir),
		},
		{
			name: fmt.Sprintf("%s/c.thrift", test_dir),
			content: `
			struct A {
				1: required map<i16, bool> something
				2: optional i16 number
			}

			struct B {
				1: optional bool something
			}`,
		},
	}

	for _, f := range files {
		os.WriteFile(f.name, []byte(f.content), 0644)
	}

	tests = []Test{
		// Exceeding the max depth with a mix of included and same-file references.
		{
			name: "a.thrift",
			prog: &ast.Program{
				Headers: []ast.Header{
					&ast.Include{Path: fmt.Sprintf("%s/b.thrift", test_dir)}},
				Definitions: []ast.Definition{
					&ast.Struct{Name: "A"},
					&ast.Struct{Name: "B", Fields: []*ast.Field{{Type: ast.TypeReference{Name: "b.A"}}}}}},
			node: &ast.Struct{Name: "A", Type: ast.StructType, Fields: []*ast.Field{{Type: ast.TypeReference{Name: "B"}}}},
			want: []string{fmt.Sprintf(`a.thrift:0:1: error: A exceeded maximum depth of 7
	a.thrift:0:0 (B) +1 (2)
	a.thrift:0:0 (b.A) +1 (3)
	%s/b.thrift:5:33 (B) +3 (6)
	%s/b.thrift:11:17 (c.A) +1 (7)
	%s/c.thrift:3:26 (bool) +1 (8) (depth)`, test_dir, test_dir, test_dir)},
		},
	}

	check = checks.CheckDepth(7, true)
	RunTests(t, &check, tests)

	test_dir = t.TempDir()

	os.WriteFile(
		fmt.Sprintf("%s/a.thrift", test_dir),
		[]byte(
			`struct A {
				1: optional bool dummyField
			}`),
		0644)

	tests = []Test{
		{
			name: fmt.Sprintf("%s/a.thrift", test_dir),
			node: &ast.Struct{Name: "A", Type: ast.StructType, Fields: []*ast.Field{
				{Name: "something", Type: ast.ListType{ValueType: ast.ListType{ValueType: ast.BaseType{ID: ast.StringTypeID}}}}}},
			want: []string{},
		},
		// Reference to a field that was already processed (in the previous test).
		{
			name: "b.thrift",
			prog: &ast.Program{Headers: []ast.Header{&ast.Include{Path: fmt.Sprintf("%s/a.thrift", test_dir)}}},
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Type: ast.TypeReference{Name: "a.A"}}}},
			want: []string{fmt.Sprintf(`b.thrift:0:1: error:  exceeded maximum depth of 3
	b.thrift:0:0 (a.A) +1 (2)
	%s/a.thrift:0:0 (string) +2 (4) (depth)`, test_dir)},
		},
		// Field with the same name as another field which was already processed.
		{
			name: "c.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Name: "something", Type: ast.BaseType{ID: ast.BoolTypeID}}}},
			want: []string{},
		},
	}

	check = checks.CheckDepth(3, true)
	RunTests(t, &check, tests)

	tests = []Test{
		{
			name: fmt.Sprintf("%s/a.thrift", test_dir),
			node: &ast.Struct{Name: "A", Type: ast.StructType, Fields: []*ast.Field{{
				Name: "something",
				Type: ast.ListType{ValueType: ast.ListType{ValueType: ast.ListType{ValueType: ast.BaseType{ID: ast.StringTypeID}}}}}}},
			want: []string{fmt.Sprintf(`%s/a.thrift:0:1: error: A exceeded maximum depth of 2
	%s/a.thrift:0:0 (string) +3 (4) (depth)`, test_dir, test_dir)},
		},
		// Field with the same name as another field which was already processed and exceeded the max depth.
		{
			name: "c.thrift",
			node: &ast.Struct{Type: ast.StructType, Fields: []*ast.Field{
				{Name: "something", Type: ast.BaseType{ID: ast.BoolTypeID}}}},
			want: []string{},
		},
	}

	check = checks.CheckDepth(2, true)
	RunTests(t, &check, tests)
}
