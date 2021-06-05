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

func TestCheckIncludeRestricted(t *testing.T) {
	tests := []Test{
		{
			name: "a.thrift",
			node: &ast.Include{Path: "good.thrift"},
			want: []string{},
		},
		{
			name: "a.thrift",
			node: &ast.Include{Path: "bad.thrift"},
			want: []string{
				`a.thrift:0:1:error: "bad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "a.thrift",
			node: &ast.Include{Path: "abad.thrift"},
			want: []string{
				`a.thrift:0:1:error: "abad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "b.thrift",
			node: &ast.Include{Path: "bad.thrift"},
			want: []string{
				`b.thrift:0:1:error: "bad.thrift" is a restricted import (include.restricted)`,
			},
		},
		{
			name: "b.thrift",
			node: &ast.Include{Path: "abad.thrift"},
			want: []string{
				`b.thrift:0:1:error: "abad.thrift" is a restricted import (include.restricted)`,
			},
		},
	}

	check := checks.CheckIncludeRestricted(map[string]string{
		"*":        `bad.thrift`,
		"a.thrift": `abad.thrift`,
	})
	RunTests(t, check, tests)
}
