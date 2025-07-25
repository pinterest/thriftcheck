# ThriftCheck

ThriftCheck is a linter for [Thrift IDL](https://thrift.apache.org/docs/idl)
files. It provides a general Thrift linting framework, a set of checks, and a
command line tool (`thriftcheck`).

You can download the [latest prebuilt release][latest] or build the project
yourself using the Go compiler toolchain (e.g. `go install`).

[latest]: https://github.com/pinterest/thriftcheck/releases/latest

## `thriftcheck`

`thriftcheck` is a configuration-driven tool for linting Thrift IDL files from
the command line:

```
usage: thriftcheck [options] [path ...]
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
```

You can pass a list of filenames or directory paths. Directories will be
expanded recursively to include all nested `.thrift` files.

You also can lint from standard input by passing `-` as the sole filename.
Use `--stdin-name` to customize the filename used in output messages.

```sh
$ thriftlint --stdin-name filename.thrift - < filename.thrift
```

Messages are reported to standard output using a familiar parseable format:

```
file.thrift:1:1: error: "py" namespace must match "^idl\\." (namespace.pattern)
file.thrift:3:1: error: unable to find include path for "bar.thrift" (include.path)
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
or `.yaml` file extension, respectively. The examples shown below use the
default [TOML](https://toml.io/) syntax.

[`example.toml`](cmd/example.toml) is an example configuration file that you
can use as a starting point.

## Checks

The full list of available checks can printed using the `--list` command line
option. By default, all checks are enabled.

You can enable or disable checks using the configuration file's top-level
`enabled` and `disabled` lists. The list of `disabled` checks is subtracted
from the full list first, and then the resulting list is filtered by the list
of `enabled` checks. Either list can be empty (the default).

### `constant.ref`

This check reports an error if a referenced constant or enum value cannot be
found in either the current scope or in an included file (using dot notation).

### `enum.size`

This check warns or errors if an enumeration's element size grows beyond a
limit.

```toml
[enum.size]
warning = 500
error = 1000
```

### `field.doc.missing`

This check warns if a field is missing a documentation comment.

### `field.id.missing`

This check reports an error if a field's ID is missing (using the legacy
implicit/auto-assigning syntax).

### `field.id.negative`

This check reports an error if a field's ID is explicitly negative.

### `field.id.zero`

This check reports an error if a field's ID is explicitly zero, which is
generally unsupported by the Apache Thrift compiler. This is distinct from
the `field.id.negative` check given the existence of the `--allow-neg-keys`
Apache Thrift compiler option.

### `field.optional`

This check warns if a field isn't declared as "optional", which is considered
a best practice.

### `field.requiredness`

This check warns if a field isn't explicitly declared as "required" or
"optional".

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

### `int.64bit`

This check warns when an integer constant exceeds the 32-bit number range.
Some languages (e.g. JavaScript) don't support 64-bit integers.

### `map.key.type`

This check can be configured to allow/disallow specific types from being used as `map<>` keys.

```toml
[checks.map]
[[checks.map.key]]
allowed = []
disallowed = []
```

For a `map<>` key, if its type is in the `disallowed` list, this check will report it as an error and stop.
Otherwise, provided that the `allowed` list is not empty, the check will report an error if the
key type is not part of the `allowed` types.

### `map.key.type.primitive`

This check ensures that only primitive types are used for `map<>` keys.

### `map.value.restricted`

This check allows you to restrict specific types from being used as `map<>` values. This is useful for enforcing coding standards around map usage, such as disallowing nested maps for simplicity or preventing unions as map values for serialization compatibility.

```toml
[checks.map.value]
restricted = [
    "union",  # Disallow unions as map values
    "map",    # Disallow nested maps
    "i32",    # Disallow i32 as map values
]
```

Supported type names include:
- **Primitives**: `bool`, `i8`, `i16`, `i32`, `i64`, `double`, `string`, `binary`
- **Collections**: `map`, `list`, `set`
- **Structures**: `union`, `struct`, `exception`

The check performs semantic type matching, including resolving `typedef`s to their underlying types, so it works correctly with typedefs and other indirect type references.

### `names.reserved`

This checks allows you to extend the [default list of reserved keywords][] with
additional disallowed names.

```toml
[checks.names]
reserved = [
    "template",
]
```

[default list of reserved keywords]: https://github.com/thriftrw/thriftrw-go/blob/0cee03e01be6bbbd45303ca94663c951f0573fd0/idl/internal/lex.rl#L110-L218

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

## `nolint` Directives

You can disable one or more checks on a per-node basis using `nolint`
directives. `nolint` directives apply to the current node and all of its
descendents. The directives's value can be empty, in which case linting is
entirely disabled, or it can be set to a comma-separated list of checks to
disable.

`nolint` directives can be written as Thrift annotations:

```thrift
enum State {
	STOPPED = 1
	RUNNING = 2
	PASSED = 3
	FAILED = 4
} (nolint = "enum.size")
```

... and as `@nolint` lines in documentation blocks:

```thrift
/**
 * States
 *
 * @nolint(enum.size)
 */
enum State {
	STOPPED = 1
	RUNNING = 2
	PASSED = 3
	FAILED = 4
}
```

The annotation syntax is preferred, but the documentation block syntax is
useful for those few cases where the target node doesn't support Thrift
annotations (such as `const` declarations).

## Editor Support

* Vim, using [ALE](https://github.com/dense-analysis/ale)

## pre-commit

[pre-commit](https://pre-commit.com/) support is provided. Simply add the
following to your `.pre-commit-config.yaml` configuration file:

```yaml
- repo: https://github.com/pinterest/thriftcheck
  rev: 1.0.0 # git revision or tag
  hooks:
    - id: thriftcheck
      name: thriftcheck
```

## Development

For information on development and making code contributions, see
[`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

This software is released under the terms of the [Apache 2.0 License](LICENSE).
