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
	"io"
	"io/ioutil"
	"os"
	"path"

	"go.uber.org/thriftrw/ast"
	"go.uber.org/thriftrw/idl"
)

// Parse parses Thrift document content.
func Parse(r io.Reader) (*ast.Program, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return idl.Parse(b)
}

// ParseFile parses a Thrift file. The filename must appear in one of the
// given directories.
func ParseFile(filename string, dirs []string) (*ast.Program, error) {
	for _, dir := range dirs {
		if f, err := os.Open(path.Join(dir, filename)); err == nil {
			return Parse(f)
		}
	}
	return nil, fmt.Errorf("%s not found in %s", filename, dirs)
}
