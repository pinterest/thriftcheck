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
	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckEnumSize returns a thriftcheck.Check that warns or errors if an
// enumeration's element size grows beyond a limit.
func CheckEnumSize(warningLimit, errorLimit int) *thriftcheck.Check {
	return thriftcheck.NewCheck("enum.size", func(c *thriftcheck.C, e *ast.Enum) {
		size := len(e.Items)
		if errorLimit > 0 && size > errorLimit {
			c.Errorf(e, "enumeration %q has more than %d items", e.Name, errorLimit)
		} else if warningLimit > 0 && size > warningLimit {
			c.Warningf(e, "enumeration %q has more than %d items", e.Name, warningLimit)
		}
	})
}
