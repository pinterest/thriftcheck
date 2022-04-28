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

func TestCheckFieldIDMissing(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Field{ID: 1},
			want: []string{},
		},
		{
			node: &ast.Field{IDUnset: true},
			want: []string{
				`t.thrift:0:1: error: field ID for "" is missing (field.id.missing)`,
			},
		},
	}

	check := checks.CheckFieldIDMissing()
	RunTests(t, check, tests)
}

func TestCheckFieldIDNegative(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Field{ID: 1, Name: "Field"},
			want: []string{},
		},
		{
			node: &ast.Field{ID: 0, Name: "Field"},
			want: []string{},
		},
		{
			node: &ast.Field{ID: -1, Name: "Field"},
			want: []string{
				`t.thrift:0:1: error: field ID for "Field" (-1) is negative (field.id.negative)`,
			},
		},
	}

	check := checks.CheckFieldIDNegative()
	RunTests(t, check, tests)
}

func TestCheckFieldIDZero(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Field{ID: 1, Name: "Field"},
			want: []string{},
		},
		{
			node: &ast.Field{ID: 0, Name: "Field"},
			want: []string{
				`t.thrift:0:1: error: field ID for "Field" is zero (field.id.zero)`,
			},
		},
		{
			node: &ast.Field{ID: -1, Name: "Field"},
			want: []string{},
		},
	}

	check := checks.CheckFieldIDZero()
	RunTests(t, check, tests)
}

func TestCheckFieldOptional(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Unspecified},
			want: []string{
				`t.thrift:0:1: warning: field "Field" (1) should be "optional" (field.optional)`,
			},
		},
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Required},
			want: []string{
				`t.thrift:0:1: warning: field "Field" (1) should be "optional" (field.optional)`,
			},
		},
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Optional},
			want: []string{},
		},
	}

	check := checks.CheckFieldOptional()
	RunTests(t, check, tests)
}

func TestCheckFieldRequiredness(t *testing.T) {
	tests := []Test{
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Unspecified},
			want: []string{
				`t.thrift:0:1: warning: field "Field" (1) should be explicitly "required" or "optional" (field.requiredness)`,
			},
		},
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Required},
			want: []string{},
		},
		{
			node: &ast.Field{ID: 1, Name: "Field", Requiredness: ast.Optional},
			want: []string{},
		},
	}

	check := checks.CheckFieldRequiredness()
	RunTests(t, check, tests)
}
