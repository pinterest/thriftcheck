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
	"math"
	"testing"

	"github.com/pinterest/thriftcheck/checks"
	"go.uber.org/thriftrw/ast"
)

func TestCheckInteger64bit(t *testing.T) {
	tests := []Test{
		{
			node: ast.ConstantInteger(0),
			want: []string{},
		},
		{
			node: ast.ConstantInteger(math.MinInt32 - 1),
			want: []string{
				`t.thrift:0:1: warning: 64-bit integer constant -2147483649 may not work in all languages (int.64bit)`,
			},
		},
		{
			node: ast.ConstantInteger(math.MaxInt32 + 1),
			want: []string{
				`t.thrift:0:1: warning: 64-bit integer constant 2147483648 may not work in all languages (int.64bit)`,
			},
		},
	}

	check := checks.CheckInteger64bit()
	RunTests(t, check, tests)
}
