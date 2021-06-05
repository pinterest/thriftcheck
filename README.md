# ThriftCheck

ThriftCheck is a linter for [Thrift IDL](https://thrift.apache.org/docs/idl)
files. It provides a general Thrift linting framework, a set of checks, and a
command line tool (`thriftcheck`).

## `thriftcheck`

`thriftcheck` is a configuration-driven tool for linting Thrift IDL files from
the command line:

```
usage: thriftcheck [options] [file ...]
  -I, --include value
    	include path (can be specified multiple times)
  -c, --config string
    	configuration file path (default ".thriftcheck.toml")
  --errors-only
    	only report errors (not warnings)
  -h, --help
    	show command help
  -l, --list
    	list all available checks and exit
  --stdin-filename string
    	filename used when piping from stdin (default "stdin")
  -v, --verbose
    	enable verbose (debugging) output
  --version
    	print the version and exit
```

You can lint from standard input by passing `-` as the sole filename. You can
also use `--stdin-name` to customize the filename used in output messages.

```sh
$ thriftlint --stdin-name filename.thrift - < filename.thrift
```

Messages are reported to standard output using a familiar parseable format:

```
file.thrift:1:1:error: "py" namespace must match "^idl\\." (namespace.pattern)
file.thrift:3:1:error: unable to find include path for "bar.thrift" (include.path)
```

If you only want errors (and not warnings) to be reported, you can use the
`--errors-only` command line option.

`thriftcheck`'s exit code indicates whether it reported any warnings (**1**)
or errors (**2**). Otherwise, exit code **0** is returned.

## Configuration

Many checks are configurable via the configuration file. This file is named
`.thriftcheck.toml` and is loaded from the current directory by default, but
you can use the `--config` command line option to use a different file. If you
prefer, you can use a JSON- or YAML-formatted file instead by using a `.json`
or `.yaml` file extension, respectively. The examples used below use the
default [TOML](https://toml.io/) syntax.

[`example.toml`](cmd/example.toml) is an example configuration file that you
can use as a starting point.

## Available Checks

### `enum.size`

This check warns or errors if an enumeration's element size grows beyond a
limit.

```toml
[enum.size]
warning = 500
error = 1000
```

### `include.path`

This check ensures that each `include`'d file can be located in the set of
given include paths.

Relative paths are resolved relative to the current working directory. The
list of `includes` specified in the configuration file is used by default,
but if any paths are specified on the command line using the `-I` option,
they will be used instead.

```toml
includes = [
    'shared',
]
```

### `include.restricted`

This check restricts some files from being imported by other files using a
map of patterns: the key is a file name pattern that matches the *including*
filename and the value is a regular expression that matches the *included*
filename. When both match, the `include` is flagged as "restricted" and an
error is reported.

```toml
[checks.include]
[[checks.include.restricted]]
"*" = "(huge|massive).thrift"
```

### `map.key.type`

This check ensures that only primitive types are used for `map<>` keys.

### `namespace.patterns`

This check ensures that a namespace's name matches a regular expression
pattern. The pattern can be configured one a per-language basis.

```toml
[[namespace.patterns]]
py = "^idl\\."
```

### `set.value.type`

This check ensures that only primitive types are used for `set<>` values.

## Custom Checks

You can also implement your own checks using the `thriftcheck` package's public
interfaces. Checks are functions which receive a `*thriftheck.C` followed by a
variable number of [`ast.Node`][ast-node]-compliant argument types.

```go
check := thriftcheck.NewCheck("enum.name", func(c *thriftcheck.C, e *ast.Enum, ei *ast.EnumItem) {
	for _, r := range ei.Name {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			c.Errorf(f, "item name %q (in %q) can only contain uppercase letters", ei.Name, e.Name)
			return
		}
	}
})
```

You can pass any list of checks to `thriftcheck.NewLinter`. You will probably
want to build a custom version of the `thriftcheck` tool that is aware of your
additional checks.

[ast-node]: https://pkg.go.dev/go.uber.org/thriftrw/ast#Node

## `nolint` Annotations

You can disable one or more checks on a per-node basis using `nolint`
annotations. `nolint` annotations apply to the current node and all of its
descendents. The annotation's value can be empty, in which case linting is
entirely disabled, or it can be set to a comma-separated list of checks to
disable.

```thrift
enum State {
	STOPPED = 1
	RUNNING = 2
	PASSED = 3
	FAILED = 4
} (nolint = "enum.size")
```

## License

This software is released under the terms of the [Apache 2.0 License](LICENSE).
