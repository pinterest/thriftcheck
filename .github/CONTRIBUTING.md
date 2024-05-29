# Contributing

First off, thanks for taking the time to contribute! This guide will answer
some common questions about how this project works.

While this is a Pinterest open source project, we welcome contributions from
everyone.

## Making Changes

1. Fork this repository to your own account
2. Make your changes and verify that `go test` passes
3. Commit your work and push to a new branch on your fork
4. Submit a [pull request](https://github.com/pinterest/thriftcheck/compare/)
5. Participate in the code review process by responding to feedback

Once there is agreement that the code is in good shape, one of the project's
maintainers will merge your contribution.

To increase the chances that your pull request will be accepted:

- Follow the style guide
- Write tests for your changes
- Write a good commit message

## Style

We use [`gofmt`][]-based code formatting. The format will be enforced by CI, so
please make sure your code is well formatted *before* you push your branch so
those checks will pass.

[`gofmt`]: https://golang.org/cmd/gofmt/

## Building

You can build the project and its dependencies using `go build .`.

You can run the `thriftcheck` command line tool using `go run ./cmd`.

## Testing

Tests are written using the [Go testing package](https://pkg.go.dev/testing).
Run them using the `go test` command.

## Releases

Releases are built using [GoReleaser](https://goreleaser.com/), triggered
using a [GitHub Actions Workflow](.github/workflows/release.yml) whenever
a new tag is pushed to the repository. Version tags are prefixed with a `v`
(e.g. `v1.0.0`).

## License

By contributing to this project, you agree that your contributions will be
licensed under its [Apache 2.0 license](LICENSE).
