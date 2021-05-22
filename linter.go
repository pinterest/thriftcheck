package thriftcheck

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

type Linter struct {
	checks Checks
	logger *log.Logger
}

type Option func(*Linter)

// WithLogger is an Option that sets the logger object used by the linter.
func WithLogger(logger *log.Logger) Option {
	return func(l *Linter) {
		l.logger = logger
	}
}

// NewLinter creates a new Linter configured with the given checks and options.
func NewLinter(checks Checks, options ...Option) *Linter {
	l := &Linter{
		checks: checks,
		logger: log.New(ioutil.Discard, "", 0),
	}
	for _, option := range options {
		option(l)
	}
	l.logger.Printf("checks: %s\n", strings.Join(checks.SortedNames(), ", "))
	return l
}

// Lint lints a single input file.
func (l *Linter) Lint(r io.Reader, filename string) (Messages, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	program, err := idl.Parse(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	return l.lint(program, filename), nil
}

// LintFiles lints multiple files. Each is opened, parsed, and linted in
// order, and the aggregate result is returned.
func (l *Linter) LintFiles(filenames []string) (Messages, error) {
	messages := Messages{}

	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return messages, fmt.Errorf("%s: %w", filename, err)
		}

		m, err := l.Lint(f, filename)
		if err != nil {
			return messages, fmt.Errorf("%s: %w", filename, err)
		}

		messages = append(messages, m...)
	}

	return messages, nil
}

func (l *Linter) lint(n ast.Node, filename string) (messages Messages) {
	l.logger.Printf("linting %s\n", filename)

	ctx := &C{Filename: filename, Logger: l.logger}
	activeChecks := overridableChecks{root: &l.checks}

	var visitor VisitorFunc
	visitor = func(w ast.Walker, n ast.Node) VisitorFunc {
		nodes := append([]ast.Node{n}, w.Ancestors()...)
		checks := activeChecks.lookup(nodes[1:])

		// Handle 'nolint' annotations
		if annotations := Annotations(n); annotations != nil {
			for _, annotation := range annotations {
				if annotation.Name == "nolint" {
					if annotation.Value == "" {
						return nil
					}

					values := strings.Split(annotation.Value, ",")
					for i := range values {
						values[i] = strings.TrimSpace(values[i])
					}

					checks = checks.Without(values)
					activeChecks.add(n, checks)
				}
			}
		}

		// Run all of the checks that match this part of the tree.
		for _, check := range *checks {
			check.Call(ctx, nodes...)
		}

		return visitor
	}

	ast.Walk(visitor, n)
	return ctx.Messages
}

// Stores Checks overrides that apply to a node and all of its children.
type overridableChecks struct {
	root      *Checks
	overrides map[ast.Node]*Checks
}

func (oc *overridableChecks) add(node ast.Node, checks *Checks) {
	if oc.overrides == nil {
		oc.overrides = make(map[ast.Node]*Checks)
	}
	oc.overrides[node] = checks
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
			if checks, ok := oc.overrides[node]; ok {
				return checks
			}
		}
	}

	return oc.root
}
