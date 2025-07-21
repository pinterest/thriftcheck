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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

// Linter is a configured Thrift linter.
type Linter struct {
	checks   Checks
	logger   *log.Logger
	includes []string
}

// Option represents a Linter option.
type Option func(*Linter)

// WithLogger is an Option that sets the log.Logger object used by the linter.
func WithLogger(logger *log.Logger) Option {
	return func(l *Linter) {
		l.logger = logger
	}
}

// WithIncludes is an Option that adds Thrift include paths to the linter.
func WithIncludes(includes []string) Option {
	return func(l *Linter) {
		l.includes = includes
	}
}

// NewLinter creates a new Linter configured with the given checks and options.
func NewLinter(checks Checks, options ...Option) *Linter {
	l := &Linter{
		checks: checks,
		logger: log.New(io.Discard, "", 0),
	}
	for _, option := range options {
		option(l)
	}
	l.logger.Printf("checks: %s\n", checks)
	l.logger.Printf("includes: %s\n", strings.Join(l.includes, " "))
	return l
}

// Lint lints a single input file.
func (l *Linter) Lint(r io.Reader, filename string) (Messages, error) {
	program, info, err := Parse(r)
	if err != nil {
		var parseError *idl.ParseError
		if errors.As(err, &parseError) {
			msgs := make(Messages, len(parseError.Errors))
			for i, err := range parseError.Errors {
				msgs[i] = Message{
					Filename: filename,
					Pos:      err.Pos,
					Check:    "parse",
					Severity: Error,
					Message:  err.Err.Error(),
				}
			}
			return msgs, nil
		}
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return l.lint(program, filename, info), nil
}

// LintFiles lints multiple files. Each is opened, parsed, and linted in
// order, and the aggregate result is returned.
func (l *Linter) LintFiles(filenames []string) (Messages, error) {
	msgs := Messages{}

	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return msgs, fmt.Errorf("%s: %w", filename, err)
		}

		m, err := l.Lint(f, filename)
		if err != nil {
			return msgs, err
		}

		msgs = append(msgs, m...)
	}

	for _, check := range l.checks {
		if check.isMultiFile {
			reflect.ValueOf(check.multiFileFn).Call([]reflect.Value{reflect.ValueOf(check.multiFileCtx)})
		}
	}

	return msgs, nil
}

func (l *Linter) lint(program *ast.Program, filename string, parseInfo *idl.Info) (messages Messages) {
	l.logger.Printf("linting %s\n", filename)

	ctx := &C{
		Filename:  filename,
		Dirs:      append([]string{filepath.Dir(filename)}, l.includes...),
		Program:   program,
		logger:    l.logger,
		parseInfo: parseInfo,
	}
	activeChecks := overridableChecks{root: &l.checks}

	var visitor VisitorFunc
	visitor = func(w ast.Walker, n ast.Node) VisitorFunc {
		nodes := append([]ast.Node{n}, w.Ancestors()...)
		checks := *activeChecks.lookup(nodes[1:])

		// Handle 'nolint' directives.
		if names, found := nolint(n); found {
			if names == nil {
				return nil
			}
			checks = checks.Without(names)
			activeChecks.add(n, &checks)
		}

		// Run all of the checks that match this part of the tree.
		for _, check := range checks {
			check.Call(ctx, nodes...)
		}

		return visitor
	}

	ast.Walk(visitor, program)
	return ctx.Messages
}

// Stores Checks overrides that apply to a node and all of its children.
type overridableChecks struct {
	root      *Checks
	overrides []override
}

type override struct {
	node   ast.Node
	checks *Checks
}

func (oc *overridableChecks) add(node ast.Node, checks *Checks) {
	oc.overrides = append(oc.overrides, override{node: node, checks: checks})
}

func (oc *overridableChecks) lookup(ancestors []ast.Node) *Checks {
	// When we don't have ancestors, use the root checks. This is also a
	// convenient time to clear any overrides from a previous hierarchy.
	if len(ancestors) == 0 {
		oc.overrides = nil
		return oc.root
	}

	// If any overrides have been assigned, find the most specific set by
	// walking up the list of ancestral nodes. We expect to find a match in
	// this loop, but we can fall back to the root if there's a logic error.
	if oc.overrides != nil {
		for _, node := range ancestors {
			for _, o := range oc.overrides {
				if o.node == node {
					return o.checks
				}
			}
		}
	}

	return oc.root
}
