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

// CheckFieldIDMissing reports an error if a field's ID is missing.
func CheckFieldIDMissing() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.id.missing", func(c *thriftcheck.C, f *ast.Field) {
		if f.IDUnset {
			c.Errorf(f, "field ID for %q is missing", f.Name)
		}
	})
}

// CheckFieldIDNegative reports an error if a field's ID is explicitly negative.
func CheckFieldIDNegative() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.id.negative", func(c *thriftcheck.C, f *ast.Field) {
		if !f.IDUnset && f.ID < 0 {
			c.Errorf(f, "field ID for %q (%d) is negative", f.Name, f.ID)
		}
	})
}

// CheckFieldIDZero reports an error if a field's ID is explicitly zero.
func CheckFieldIDZero() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.id.zero", func(c *thriftcheck.C, f *ast.Field) {
		if !f.IDUnset && f.ID == 0 {
			c.Errorf(f, "field ID for %q is zero", f.Name)
		}
	})
}

// CheckFieldOptional warns if a field isn't declared as "optional".
func CheckFieldOptional() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.optional", func(c *thriftcheck.C, f *ast.Field) {
		if f.Requiredness != ast.Optional {
			c.Warningf(f, `field %q (%d) should be "optional"`, f.Name, f.ID)
		}
	})
}

// CheckFieldRequiredness warns if a field isn't explicitly declared as "required" or "optional".
func CheckFieldRequiredness() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.requiredness", func(c *thriftcheck.C, f *ast.Field) {
		if f.Requiredness == ast.Unspecified {
			c.Warningf(f, `field %q (%d) should be explicitly "required" or "optional"`, f.Name, f.ID)
		}
	})
}

// CheckFieldCommentMissing warns if a field doesn't have comments.
func CheckFieldCommentMissing() *thriftcheck.Check {
	return thriftcheck.NewCheck("field.comment.missing", func(c *thriftcheck.C, f *ast.Field) {
		if f.Doc == "" {
			c.Warningf(f, `field %q (%d) don't have comments`, f.Name, f.ID)
		}
	})
}
