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
	"log"
	"reflect"
	"sort"
	"strings"

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

// Check is a named check function.
type Check struct {
	Name string
	fn   interface{}
}

// Checks is a list of checks.
type Checks []*Check

// NewCheck creates a new Check.
func NewCheck(name string, fn interface{}) *Check {
	if fn == nil {
		panic("check function must be a Func; got nil")
	}

	f := reflect.TypeOf(fn)
	if f.Kind() != reflect.Func {
		panic("check function must be a Func; got " + f.String())
	}
	if f.NumIn() < 2 {
		panic("check function receive at least two arguments")
	}
	if f.In(0) != reflect.TypeOf(&C{}) {
		panic("check function must receive C as its first argument")
	}
	for i := 1; i < f.NumIn(); i++ {
		if !f.In(i).Implements(nodeInterface) {
			panic("all additional arguments must implement ast.Node")
		}
	}

	return &Check{Name: name, fn: fn}
}

// Call the check function if its arguments end with the current node in the
// hierarchy and all other variable arguments are its strictly ordered parents.
//
// The first argument is always a *C instance.
//
// The nodes are ordered from the current node through it ancestors.
//
//	hierarchy = {*ast.EnumItem, *ast.Enum, *ast.Program}
//
// The following functions would match:
//
// 		f(*C, *ast.Program, *ast.Enum, *ast.EnumItem)
// 		f(*C, *ast.Enum, *ast.EnumItem)
// 		f(*C, *ast.EnumItem)
//
// But these would not:
//
// 		f(*C, *ast.Program)
// 		f(*C, *ast.Enum)
// 		f(*C, *ast.EnumItem, *ast.Enum)
// 		f(*C, *ast.Program, *ast.EnumItem)
//
// Function arguments can also use the generic ast.Node interface type:
//
//		f(*C, ast.Node)
//		f(*C, *ast.Program, ast.Node)
//		f(*C, parent, node ast.Node)
//
func (c *Check) Call(ctx *C, nodes ...ast.Node) bool {
	if len(nodes) < 1 {
		panic("expected at least one node")
	}

	f := reflect.TypeOf(c.fn)

	// Ensure that the current node is compatible with the last argument.
	if !reflect.TypeOf(nodes[0]).AssignableTo(f.In(f.NumIn() - 1)) {
		return false
	}

	if len(nodes) < f.NumIn()-1 {
		return false
	}

	args := []reflect.Value{reflect.ValueOf(ctx)}
	for i := 1; i < f.NumIn() && i <= len(nodes); i++ {
		node := nodes[f.NumIn()-i-1]
		if arg := reflect.ValueOf(node); arg.Type().AssignableTo(f.In(i)) {
			args = append(args, arg)
		}
	}

	// Ensire the arguments match.
	if len(args) != f.NumIn() {
		return false
	}

	ctx.Check = c.Name
	reflect.ValueOf(c.fn).Call(args)
	return true
}

func (c Checks) String() string {
	return strings.Join(c.SortedNames(), " ")
}

// SortedNames returns a sorted list of the checks' names.
func (c Checks) SortedNames() []string {
	keys := make([]string, 0, len(c))
	for _, check := range c {
		keys = append(keys, check.Name)
	}
	sort.Strings(keys)
	return keys
}

// With returns a copy with only those checks whose names match the given prefixes.
func (c Checks) With(prefixes []string) Checks {
	checks := make(Checks, 0)
	for _, check := range c {
		for _, prefix := range prefixes {
			if check.Name == prefix || strings.HasPrefix(check.Name, prefix+".") {
				checks = append(checks, check)
				break
			}
		}
	}
	return checks
}

// Without returns a copy without those checks whose names match the given prefixes.
func (c Checks) Without(prefixes []string) Checks {
	checks := make(Checks, 0)
next:
	for _, check := range c {
		for _, prefix := range prefixes {
			if check.Name == prefix || strings.HasPrefix(check.Name, prefix+".") {
				continue next
			}
		}
		checks = append(checks, check)
	}
	return checks
}

// C is a type passed to all check functions to provide context.
type C struct {
	Filename  string
	Dirs      []string
	Program   *ast.Program
	Check     string
	Messages  Messages
	logger    *log.Logger
	parseInfo *idl.Info
}

func (c *C) pos(n ast.Node) ast.Position {
	if c.parseInfo != nil {
		return c.parseInfo.Pos(n)
	}
	pos, _ := ast.Pos(n)
	return pos
}

// Logf prints a formatted message to the verbose output logger.
func (c *C) Logf(message string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(message, args...)
	}
}

// Warningf records a new message for the given node with Warning severity.
func (c *C) Warningf(node ast.Node, message string, args ...interface{}) {
	m := Message{Filename: c.Filename, Pos: c.pos(node), Node: node, Check: c.Check, Severity: Warning, Message: fmt.Sprintf(message, args...)}
	c.Messages = append(c.Messages, m)
}

// Errorf records a new message for the given node with Error severity.
func (c *C) Errorf(node ast.Node, message string, args ...interface{}) {
	m := Message{Filename: c.Filename, Pos: c.pos(node), Node: node, Check: c.Check, Severity: Error, Message: fmt.Sprintf(message, args...)}
	c.Messages = append(c.Messages, m)
}

// Resolve resolves a name.
func (c *C) Resolve(name string) ast.Node {
	if n, err := Resolve(name, c.Program, c.Dirs); err == nil {
		return n
	}
	return nil
}

// ResolveConstant resolves a constant reference to its target.
func (c *C) ResolveConstant(ref ast.ConstantReference) ast.Node {
	if n, err := ResolveConstant(ref, c.Program, c.Dirs); err == nil {
		return n
	}
	return nil
}

// ResolveType resolves a type reference to its target type.
func (c *C) ResolveType(ref ast.TypeReference) ast.Node {
	if n, err := ResolveType(ref, c.Program, c.Dirs); err == nil {
		return n
	}
	return nil
}
