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

/*
Thriftcheck lints Thrift IDL files.

Usage:

	thriftcheck [options] [path ...]

Options:

	-I, --include value
		include path (can be specified multiple times)
	-c, --config string
		configuration file path (default ".thriftcheck.toml")
	--errors-only
		only report errors (not warnings)
	-h, --help
		show command help
	-l, --list
		list all available checks with their status and exit
	--stdin-filename string
		filename used when piping from stdin (default "stdin")
	-v, --verbose
		enable verbose (debugging) output
	--version
		print the version and exit
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kkyr/fig"
	"github.com/pinterest/thriftcheck"
	"github.com/pinterest/thriftcheck/checks"
	"rsc.io/getopt"
)

// Config represents all of the configurable values.
type Config struct {
	Includes []string `fig:"includes"`
	Checks   struct {
		Enabled  []string `fig:"enabled"`
		Disabled []string `fix:"disabled"`

		Enum struct {
			Size struct {
				Warning int `fig:"warning"`
				Error   int `fig:"error"`
			}
		}

		Include struct {
			Restricted map[string]*regexp.Regexp `fig:"restricted"`
		}

		Map struct {
			Key struct {
				AllowedTypes    []thriftcheck.ThriftType `fig:"allowedTypes"`
				DisallowedTypes []thriftcheck.ThriftType `fig:"disallowedTypes"`
			}
			Value struct {
				AllowedTypes    []thriftcheck.ThriftType `fig:"allowedTypes"`
				DisallowedTypes []thriftcheck.ThriftType `fig:"disallowedTypes"`
			}
		}

		Set struct {
			AllowedTypes    []thriftcheck.ThriftType `fig:"allowedTypes"`
			DisallowedTypes []thriftcheck.ThriftType `fig:"disallowedTypes"`
		}

		Names struct {
			Reserved []string `fig:"reserved"`
		}

		Namespace struct {
			Patterns map[string]*regexp.Regexp `fig:"patterns"`
		}

		Types struct {
			AllowedTypes    []thriftcheck.ThriftType `fig:"allowedTypes"`
			DisallowedTypes []thriftcheck.ThriftType `fig:"disallowedTypes"`
		}
	}
}

// Strings accumlates strings for a repeated command line flag.
type Strings []string

func (i *Strings) String() string {
	return strings.Join(*i, " ")
}

// Set adds a new value using a flag.Var-compatible interface.
func (i *Strings) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	version       = "dev"
	revision      = "dev"
	includes      Strings
	configFile    = flag.String("c", ".thriftcheck.toml", "configuration file path")
	errorsOnly    = flag.Bool("errors-only", false, "only report errors (not warnings)")
	helpFlag      = flag.Bool("h", false, "show command help")
	listFlag      = flag.Bool("l", false, "list all available checks with their status and exit")
	stdinFilename = flag.String("stdin-filename", "stdin", "filename used when piping from stdin")
	verboseFlag   = flag.Bool("v", false, "enable verbose (debugging) output")
	versionFlag   = flag.Bool("version", false, "print the version and exit")
)

func init() {
	flag.Var(&includes, "I", "include path (can be specified multiple times)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: thriftcheck [options] [path ...]\n")
		getopt.PrintDefaults()
	}
	getopt.Aliases(
		"I", "include",
		"c", "config",
		"h", "help",
		"l", "list",
		"v", "verbose")
}

func isFlagSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

func loadConfig(cfg *Config) error {
	if err := fig.Load(cfg, fig.UseStrict(), fig.File(*configFile)); err != nil {
		// Ignore FileNotFound when we're using the default configuration file.
		if errors.Is(err, fig.ErrFileNotFound) && !isFlagSet("c") {
			return nil
		}
		return err
	}
	return nil
}

func lint(l *thriftcheck.Linter, paths []string) (thriftcheck.Messages, error) {
	if len(paths) == 1 && paths[0] == "-" {
		return l.Lint(os.Stdin, *stdinFilename)
	}
	paths, err := expandPaths(paths)
	if err != nil {
		return nil, err
	}
	return l.LintFiles(paths)
}

func expandPaths(paths []string) ([]string, error) {
	var filenames []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			filenames = append(filenames, path)
			continue
		}

		err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == ".thrift" {
				filenames = append(filenames, path)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return filenames, nil
}

func main() {
	// Parse command line flags
	if err := getopt.CommandLine.Parse(os.Args[1:]); err != nil {
		os.Exit(1 << uint(thriftcheck.Error))
	}
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}
	if *versionFlag {
		fmt.Fprintf(flag.CommandLine.Output(), "thriftcheck %s (%s)\n", version, revision)
		os.Exit(0)
	}

	// Load the (optional) configuration file
	var cfg Config
	if err := loadConfig(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1 << uint(thriftcheck.Error))
	}

	if len(includes) > 0 {
		cfg.Includes = includes
	}

	// Build the set of checks we'll use for the linter
	allChecks := thriftcheck.Checks{
		checks.CheckConstantRef(),
		checks.CheckEnumSize(cfg.Checks.Enum.Size.Warning, cfg.Checks.Enum.Size.Error),
		checks.CheckFieldIDMissing(),
		checks.CheckFieldIDNegative(),
		checks.CheckFieldIDZero(),
		checks.CheckFieldOptional(),
		checks.CheckFieldRequiredness(),
		checks.CheckFieldDocMissing(),
		checks.CheckIncludeCycle(),
		checks.CheckIncludePath(),
		checks.CheckIncludeRestricted(cfg.Checks.Include.Restricted),
		checks.CheckInteger64bit(),
		checks.CheckMapKeyType(cfg.Checks.Map.Key.AllowedTypes, cfg.Checks.Map.Key.DisallowedTypes),
		checks.CheckMapValueType(cfg.Checks.Map.Value.AllowedTypes, cfg.Checks.Map.Value.DisallowedTypes),
		checks.CheckNamesReserved(cfg.Checks.Names.Reserved),
		checks.CheckNamespacePattern(cfg.Checks.Namespace.Patterns),
		checks.CheckSetValueType(cfg.Checks.Set.AllowedTypes, cfg.Checks.Set.DisallowedTypes),
		checks.CheckTypes(cfg.Checks.Types.AllowedTypes, cfg.Checks.Types.DisallowedTypes),
	}

	checks := allChecks
	if len(cfg.Checks.Disabled) > 0 {
		checks = checks.Without(cfg.Checks.Disabled)
	}
	if len(cfg.Checks.Enabled) > 0 {
		checks = checks.With(cfg.Checks.Enabled)
	}
	if *listFlag {
		enabledNames := make(map[string]bool, len(checks))
		for _, check := range checks {
			enabledNames[check.Name] = true
		}
		for _, name := range allChecks.SortedNames() {
			status := "disabled"
			if enabledNames[name] {
				status = "enabled"
			}
			fmt.Printf("%-30s %s\n", name, status)
		}
		os.Exit(0)
	}

	// Build the set of linter options
	options := []thriftcheck.Option{
		thriftcheck.WithIncludes(cfg.Includes),
	}
	if *verboseFlag {
		logger := log.New(os.Stderr, "", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		options = append(options, []thriftcheck.Option{
			thriftcheck.WithVerboseFlag(),
			thriftcheck.WithLogger(logger),
		}...)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Create the linter and run it over the input files
	linter := thriftcheck.NewLinter(checks, options...)
	messages, err := lint(linter, paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1 << uint(thriftcheck.Error))
	}

	// Print any messages reported by the linter
	status := 0
	for _, m := range messages {
		if *errorsOnly && m.Severity != thriftcheck.Error {
			continue
		}
		fmt.Println(m)
		status |= 1 << uint(m.Severity)
	}
	os.Exit(status)
}
