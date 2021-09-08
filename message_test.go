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
	"testing"

	"go.uber.org/thriftrw/ast"
)

func TestMessageString(t *testing.T) {
	tests := []struct {
		m *Message
		s string
	}{
		{
			&Message{Filename: "a.thrift", Pos: ast.Position{Line: 5}, Check: "check", Severity: Warning, Message: "Warning"},
			"a.thrift:5:1: warning: Warning (check)",
		},
		{
			&Message{Filename: "a.thrift", Pos: ast.Position{Line: 5}, Check: "check", Severity: Error, Message: "Error"},
			"a.thrift:5:1: error: Error (check)",
		},
	}

	for _, tt := range tests {
		s := tt.m.String()
		if s != tt.s {
			t.Errorf("%#v expected %q, got %q", tt.m, tt.s, s)
		}
	}
}
