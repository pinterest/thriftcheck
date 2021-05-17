package checks

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pinterest/thriftcheck"
	"go.uber.org/thriftrw/ast"
)

func CheckIncludes(includes []string) thriftcheck.Check {
	return thriftcheck.NewCheck("includes", func(c *thriftcheck.C, i *ast.Include) {
		// Always check the file's directory first to match `thrift`s behavior.
		dirs := append([]string{filepath.Dir(c.Filename)}, includes...)

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
