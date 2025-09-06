// Copyright 2025 Pinterest
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

package checks

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

type config struct {
	maxDepth      int
	maxDepthUnset bool
	allowCycles   bool
}

type structNode struct {
	id     string
	fields []*ast.Field
	source *source
}

type source struct {
	filename string
	program  *ast.Program
}

type typeNode struct {
	isBaseType bool
	sourceNode structNode
	ref        typeRef
}

type typeRef struct {
	filename string
	name     string
	line     int
	col      int
	depth    int
}

func NewStructNode(s *ast.Struct, f string, p *ast.Program) structNode {
	return structNode{id: f + s.Name, fields: s.Fields, source: &source{filename: f, program: p}}
}

// CheckDepth returns a thriftcheck.Check that reports an error
// if a Struct, Union, or Exception exceeds a specified depth.
func CheckDepth(maxDepth int, allowCycles bool) thriftcheck.Check {
	structIdToTypes := make(map[string]map[string]*typeNode)

	return thriftcheck.NewCheck("depth", func(c *thriftcheck.C, s *ast.Struct) {
		for _, a := range s.Annotations {
			if a.Name == "maxDepth" {
				i, err := strconv.Atoi(a.Value)
				if err != nil {
					c.Errorf(s, `value of %q for "maxDepth" annotation could not be parsed into an integer`, a.Value)
					return
				}
				if i < 1 {
					c.Errorf(s, `"maxDepth" annotations should be positive, but got %q`, i)
					return
				}
				maxDepth = i
				break
			}
		}

		maxDepthUnset := maxDepth == 0

		if maxDepthUnset && allowCycles {
			return
		}

		depth, cycle, path := getDepth(
			NewStructNode(s, c.Filename, c.Program), 1, 1,
			make(map[string]bool), []*typeNode{}, structIdToTypes,
			config{maxDepth: maxDepth, maxDepthUnset: maxDepthUnset, allowCycles: allowCycles}, c)

		if (!maxDepthUnset && depth > maxDepth) || (cycle && !allowCycles) {
			pathDetails := []string{}
			accD := 1
			for _, e := range path {
				accD += e.ref.depth
				pathDetails = append(
					pathDetails,
					fmt.Sprintf("\t%s:%d:%d (%s) +%d (%d)", e.ref.filename, e.ref.line, e.ref.col, e.ref.name, e.ref.depth, accD))
			}

			m := fmt.Sprintf("exceeded maximum depth of %d", maxDepth)
			if cycle && !allowCycles {
				m = "led to a cycle"
			}
			c.Errorf(s, "%s %s\n%s", s.Name, m, strings.Join(pathDetails, "\n"))
		}
	})
}

func getDepth(
	s structNode, curD, maxD int, vis map[string]bool, path []*typeNode,
	structIdToTypes map[string]map[string]*typeNode, cfg config, c *thriftcheck.C,
) (int, bool, []*typeNode) {
	if vis[s.id] {
		return curD, true, path
	}

	vis[s.id] = true

	maxD = max(maxD, curD)
	if !cfg.maxDepthUnset && maxD > cfg.maxDepth {
		return maxD, false, path
	}

	expandStructFields(s, structIdToTypes, c)

	var cycle bool
	for _, t := range structIdToTypes[s.id] {
		if t.isBaseType {
			if newD := curD + t.ref.depth; !cfg.maxDepthUnset && newD > cfg.maxDepth {
				return newD, cycle, append(path, t)
			}
			continue
		}

		d, c, path := getDepth(t.sourceNode, curD+t.ref.depth, maxD, vis, append(path, t), structIdToTypes, cfg, c)
		cycle = cycle || c
		maxD = max(maxD, d)
		if (!cfg.maxDepthUnset && maxD > cfg.maxDepth) || (cycle && !cfg.allowCycles) {
			return maxD, cycle, path
		}
	}

	vis[s.id] = false
	return maxD, false, []*typeNode{}
}

func expandStructFields(s structNode, structIdToTypes map[string]map[string]*typeNode, c *thriftcheck.C) {
	if structIdToTypes[s.id] == nil {
		structIdToTypes[s.id] = make(map[string]*typeNode)
		for _, f := range s.fields {
			expandType(f.Type, 1, s.source, make(map[string]bool), structIdToTypes[s.id], c)
		}
	}
}

// expandType traverses an ast.Type, resolving references when needed,
// to store the deepest ast.BaseType and *ast.Struct types
// relative to the parent struct.
func expandType(t ast.Type, depth int, src *source, vis map[string]bool, deepestTypes map[string]*typeNode, c *thriftcheck.C) {
	switch v := t.(type) {
	case ast.BaseType:
		updateTypeIfDeepest(v.String(), v.String(), src, v.Line, v.Column, depth-1,
			&typeNode{isBaseType: true}, deepestTypes)
	case ast.TypeReference:
		name := v.String()
		n, rInfo, err := thriftcheck.Resolve(name, src.program, []string{filepath.Dir(src.filename)}, c.ParseCache)
		if err != nil {
			return
		}
		newSrc := src
		if strings.Contains(name, ".") {
			newSrc = &source{filename: rInfo.Filename, program: rInfo.Program}
		}

		switch n := n.(type) {
		case *ast.Constant:
			expandType(n.Type, depth, newSrc, vis, deepestTypes, c)
		case *ast.Typedef:
			key := src.filename + n.Name
			if vis[key] {
				c.Warningf(t, "found a cycle resolving typedef %q", n.Name)
				return
			}
			vis[key] = true
			expandType(n.Type, depth, newSrc, vis, deepestTypes, c)
			vis[key] = false
		default:
			if s, ok := n.(*ast.Struct); ok {
				updateTypeIfDeepest(newSrc.filename+s.Name, name, src, v.Line, v.Column, depth,
					&typeNode{sourceNode: NewStructNode(s, newSrc.filename, newSrc.program)}, deepestTypes)
			}
		}
	case ast.MapType:
		expandType(v.KeyType, depth+1, src, vis, deepestTypes, c)
		expandType(v.ValueType, depth+1, src, vis, deepestTypes, c)
	case ast.ListType:
		expandType(v.ValueType, depth+1, src, vis, deepestTypes, c)
	case ast.SetType:
		expandType(v.ValueType, depth+1, src, vis, deepestTypes, c)
	}
}

func updateTypeIfDeepest(key, name string, src *source, line, col, depth int, baseT *typeNode, deepestTypes map[string]*typeNode) {
	if deepestTypes[key] == nil {
		deepestTypes[key] = baseT
	}
	if depth > deepestTypes[key].ref.depth {
		deepestTypes[key].ref = typeRef{name: name, filename: src.filename, line: line, col: col, depth: depth}
	}
}
