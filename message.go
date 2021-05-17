package thriftcheck

import (
	"fmt"

	"go.uber.org/thriftrw/ast"
)

type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Warning {
		return "warning"
	}
	return "error"
}

type Message struct {
	Filename string
	Node     ast.Node
	Check    string
	Severity Severity
	Message  string
}

func (m *Message) Line() int {
	return ast.LineNumber(m.Node)
}

func (m *Message) String() string {
	return fmt.Sprintf("%s:%d:%d:%s: %s (%s)", m.Filename, m.Line(), 1, m.Severity, m.Message, m.Check)
}

type Messages []*Message
