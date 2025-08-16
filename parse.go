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

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

type ParseRes struct {
	program *ast.Program
	info    *idl.Info
	err     error
}

var filenameToParseRes = make(map[string]*ParseRes)

// Parse parses Thrift document content.
func Parse(r io.Reader, filename string) (*ast.Program, *idl.Info, error) {
	if res, ok := filenameToParseRes[filename]; ok {
		return res.program, res.info, res.err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	cfg := idl.Config{Info: &idl.Info{}}
	prog, err := cfg.Parse(b)

	filenameToParseRes[filename] = &ParseRes{program: prog, info: cfg.Info, err: err}

	return prog, cfg.Info, err
}

// ParseFile parses a Thrift file. The filename must appear in one of the
// given directories. If found, it also returns the opened file's `Name()`.
func ParseFile(filename string, dirs []string) (*ast.Program, *idl.Info, string, error) {
	if filepath.IsAbs(filename) {
		if f, err := os.Open(filename); err == nil {
			defer f.Close()
			program, info, err := Parse(f, f.Name())
			return program, info, f.Name(), err
		}
		return nil, nil, "", fmt.Errorf("%s not found", filename)
	}

	for _, dir := range dirs {
		if f, err := os.Open(filepath.Join(dir, filename)); err == nil {
			defer f.Close()
			program, info, err := Parse(f, f.Name())
			return program, info, f.Name(), err
		}
	}

	return nil, nil, "", fmt.Errorf("%s not found in %s", filename, dirs)
}
