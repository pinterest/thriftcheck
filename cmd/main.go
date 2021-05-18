package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kkyr/fig"
	"github.com/pinterest/thriftcheck"
	"github.com/pinterest/thriftcheck/checks"
)

type Options struct {
	ConfigFile string
}

type Config struct {
	Checks struct {
		Enabled  []string `fig:"enabled"`
		Disabled []string `fix:"disabled"`
	}
	Includes []string `fig:"includes"`
	Enum     struct {
		Size struct {
			Warning int `fig:"warning"`
			Error   int `fig:"error"`
		}
	}
	Namespace struct {
		Patterns map[string]string `fig:"patterns"`
	}
}

type Includes []string

func (i *Includes) String() string {
	return strings.Join(*i, " ")
}

func (i *Includes) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var configFile = flag.String("c", "thriftcheck.toml", "configuration file path")
var listFlag = flag.Bool("l", false, "list all available checks")
var verboseFlag = flag.Bool("v", false, "enable verbose (debugging) output")

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
	if err := fig.Load(cfg, fig.File(*configFile)); err != nil {
		// Ignore FileNotFound when we're using the default configuration file.
		if errors.Is(err, fig.ErrFileNotFound) && !isFlagSet("c") {
			return nil
		}
		return err
	}
	return nil
}

func main() {
	var includes Includes

	// Parse command line flags
	flag.Var(&includes, "I", "include path (can be specified multiple times)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: thriftcheck [options] [file ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Load the (optional) configuration file
	var cfg Config
	if err := loadConfig(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1 << uint(thriftcheck.Error))
	}
	if isFlagSet("I") {
		cfg.Includes = includes
	}

	// Build the set of checks we'll use for the linter
	checks := &thriftcheck.Checks{
		checks.CheckIncludes(cfg.Includes),
		checks.CheckNamespacePattern(cfg.Namespace.Patterns),
		checks.CheckEnumSize(cfg.Enum.Size.Warning, cfg.Enum.Size.Error),
	}
	if *listFlag {
		fmt.Println(strings.Join(checks.SortedKeys(), "\n"))
		os.Exit(0)
	}
	if len(cfg.Checks.Disabled) > 0 {
		checks = checks.Without(cfg.Checks.Disabled)
	}
	if len(cfg.Checks.Enabled) > 0 {
		checks = checks.With(cfg.Checks.Enabled)
	}

	// Build the set of linter options
	options := []thriftcheck.Option{}
	if *verboseFlag {
		logger := log.New(os.Stderr, "", log.Ltime|log.Lmicroseconds|log.Lshortfile)
		options = append(options, thriftcheck.WithLogger(logger))
	}

	// Create the linter and run it over the input files
	linter := thriftcheck.NewLinter(*checks, options...)
	messages, err := linter.Lint(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1 << uint(thriftcheck.Error))
	}

	// Print any messages reported by the linter
	status := 0
	for _, m := range messages {
		fmt.Println(m)
		status |= 1 << uint(m.Severity)
	}
	os.Exit(status)
}
