// Copyright 2022 Pinterest
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

// CheckConstantRef returns a thriftcheck.Check that ensures that a constant
// reference's target can be resolved.
func CheckConstantRef() thriftcheck.Check {
	return thriftcheck.NewCheck("constant.ref", func(c *thriftcheck.C, ref ast.ConstantReference) {
		if c.ResolveConstant(ref) == nil {
			c.Errorf(ref, "unable to find a constant or enum value named %q", ref.Name)
		}
	})
}
