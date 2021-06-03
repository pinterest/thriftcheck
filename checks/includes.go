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

package checks

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckIncludeExists returns a thriftcheck.Check that verifies that all of the
// files `include`'d by a Thrift file can be found in the includes paths.
func CheckIncludeExists() thriftcheck.Check {
	return thriftcheck.NewCheck("includes.exist", func(c *thriftcheck.C, i *ast.Include) {
		// Always check the file's directory first to match `thrift`s behavior.
		dirs := append([]string{filepath.Dir(c.Filename)}, c.Includes...)

		found := false
		for _, dir := range dirs {
			if _, err := os.Stat(path.Join(dir, i.Path)); err == nil {
				found = true
				break
			}
		}
		if !found {
			c.Errorf(i, "unable to find include path for '%s'", i.Path)
		}
	})
}
