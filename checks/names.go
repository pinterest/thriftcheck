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
	"reflect"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

// Name returns an ast.Node's Name string.
func nodeName(node ast.Node) string {
	if v := reflect.ValueOf(node); v.Kind() == reflect.Ptr {
		if f := v.Elem().FieldByName("Name"); f.IsValid() {
			return f.Interface().(string)
		}
	}
	return ""
}

// CheckNamesReserved checks if a node's name is in the list of reserved names.
func CheckNamesReserved(names []string) *thriftcheck.Check {
	reserved := make(map[string]bool)
	for _, name := range names {
		reserved[name] = true
	}

	return thriftcheck.NewCheck("names.reserved", func(c *thriftcheck.C, n ast.Node) {
		if name := nodeName(n); name != "" && reserved[name] {
			c.Errorf(n, "%q is a reserved name", name)
		}
	})
}
