# ThriftCheck

ThriftCheck is a linter for [Thrift IDL](https://thrift.apache.org/docs/idl)
files. It provides a general Thrift linting framework, a set of checks, and a
command line tool (`thriftcheck`).

## `thriftcheck`

`thriftcheck` is a configuration-driven tool for linting Thrift IDL files from
the command line:

```sh
usage: thriftcheck [options] [file ...]
  -I value
    	include path (can be specified multiple times)
  -c string
    	configuration file path (default "thriftcheck.toml")
  -l	list all available checks
  -v	enable verbose (debugging) output
  ```

## Available Checks

### `enum.size`

This check warns or errors if an enumeration's element size grows beyond a
limit.

```toml
[enum]
warning = 500
error = 1000
```

### `includes`

This check ensures that each `include`'d file can be located in the set of
given include paths.

```toml
includes = [
    'shared',
]
```

### `namespace.pattern`

This check ensures that a namespace's name matches a regular expression
pattern. The pattern can be configured one a per-language basis.

```toml
[[namespace.patterns]]
py = "^idl\\."
```

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

# License

This software is released under the terms of the [Apache 2.0 License](LICENSE).