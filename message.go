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
	Node     ast.Node
	Check    string
	Severity Severity
	Message  string
}

// Line returns the line number of this message's AST node.
func (m *Message) Line() int {
	return ast.LineNumber(m.Node)
}

func (m *Message) String() string {
	return fmt.Sprintf("%s:%d:%d:%s: %s (%s)", m.Filename, m.Line(), 1, m.Severity, m.Message, m.Check)
}

// Messages is a list of messages.
type Messages []*Message
