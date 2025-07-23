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
	"math"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// CheckInteger64bit warns when an integer constant exceeds the 32-bit number range.
func CheckInteger64bit() thriftcheck.Check {
	return thriftcheck.NewCheck("int.64bit", func(c *thriftcheck.C, i ast.ConstantInteger) {
		if i < math.MinInt32 || i > math.MaxInt32 {
			c.Warningf(i, "64-bit integer constant %d may not work in all languages", i)
		}
	})
}
