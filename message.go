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
	"fmt"

	"go.uber.org/thriftrw/ast"
)

// Severity represents the severity level of a message.
type Severity int

const (
	// Warning indicates a warning
	Warning Severity = iota
	// Error indicates an error
	Error
)

func (s Severity) String() string {
	if s == Warning {
		return "warning"
	}
	return "error"
}

// Message is a message produced by a Check.
type Message struct {
	Filename string
	Pos      ast.Position
	Node     ast.Node
	Check    string
	Severity Severity
	Message  string
}

func (m Message) String() string {
	col := m.Pos.Column
	if col == 0 {
		col = 1
	}
	return fmt.Sprintf("%s:%d:%d:%s: %s (%s)", m.Filename, m.Pos.Line, col, m.Severity, m.Message, m.Check)
}

// Messages is a list of messages.
type Messages []Message
