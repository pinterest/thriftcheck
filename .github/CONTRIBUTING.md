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

## License

By contributing to this project, you agree that your contributions will be
licensed under its [Apache 2.0 license](LICENSE).
