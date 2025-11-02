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

package thriftcheck

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

// Parse parses Thrift document content.
func Parse(r io.Reader) (*ast.Program, *idl.Info, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	cfg := idl.Config{Info: &idl.Info{}}
	prog, err := cfg.Parse(b)
	return prog, cfg.Info, err
}

type parsedFile struct {
	prog *ast.Program
	info *idl.Info
}

// FileParser caches parsed Thrift files to avoid re-parsing included files.
// It is safe for concurrent use by multiple goroutines.
type FileParser struct {
	mu    sync.RWMutex
	cache map[string]parsedFile
	dirs  []string
}

// NewParser returns a new Parser with a list of directories that will be
// searched for files.
func NewFileParser(dirs []string) *FileParser {
	return &FileParser{
		cache: make(map[string]parsedFile),
		dirs:  dirs,
	}
}

// Parse parses Thrift document content with the given filename.
func (p *FileParser) Parse(r io.Reader, filename string) (*ast.Program, *idl.Info, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, nil, err
	}

	p.mu.RLock()
	cached, ok := p.cache[filename]
	p.mu.RUnlock()

	if ok {
		return cached.prog, cached.info, nil
	}

	prog, info, err := Parse(r)
	if err == nil {
		p.mu.Lock()
		p.cache[filename] = parsedFile{prog: prog, info: info}
		p.mu.Unlock()
	}

	return prog, info, err
}

// ParseFile parses a Thrift file from its filename.
func (p *FileParser) ParseFile(filename string) (*ast.Program, *idl.Info, error) {
	if filepath.IsAbs(filename) {
		if f, err := os.Open(filename); err == nil {
			defer f.Close()
			return p.Parse(f, filename)
		}
		return nil, nil, fmt.Errorf("%s not found", filename)
	}

	dirs := append([]string{filepath.Dir(filename)}, p.dirs...)
	for _, dir := range dirs {
		if f, err := os.Open(filepath.Join(dir, filename)); err == nil {
			defer f.Close()
			return p.Parse(f, f.Name())
		}
	}

	return nil, nil, fmt.Errorf("%s not found in %s", filename, p.dirs)
}
