# Contributing

Thank you for helping. Bug reports, fixes, documentation and features are welcome.

## Before you start

- A bug: open an issue with the steps to reproduce it, unless the fix is obvious.
- A feature or an incompatible change: open an issue first, to agree on the design before writing it.
- A vulnerability: do not open an issue, see [SECURITY.md](SECURITY.md).

## Development

Requirements: Go (see `go.mod`), [gofumpt](https://github.com/mvdan/gofumpt) and
[golangci-lint](https://golangci-lint.run) v2.

```bash
make check   # build, formatting, go vet, golangci-lint and the tests with the race detector
make fmt     # format the code
make cover   # tests and total coverage
make fuzz    # each fuzz test for 30 seconds, FUZZTIME=5m for longer
```

A failing input found by `make fuzz` is saved under `testdata/fuzz`: commit it with the fix, it becomes a regression
test.

A pull request is merged once `make check` passes, as the CI runs it.

## Rules of the code

- Format with `gofumpt`; `golangci-lint` reports no issue. A `//nolint` names its linter and says why.
- Write the tests with the code: a fix comes with the test that fails without it.
- Document every exported name, and the documentation pages in `documentation/` the change affects.
- Pass a `context.Context` to what does I/O, and return the errors, wrapped with `%w`: the library never exits nor
  panics on a runtime error.
- Log with `common/log`, as key-value pairs, and never a secret.

## Pull requests

- One subject per pull request, rebased on `master`.
- Commit messages: a short subject in the imperative ("Limit the size of the request bodies"), then why the change is
  made.
- Describe the incompatible changes in the pull request: they are listed in the notes of the next release.

By contributing, you agree that your contribution is licensed under the [MIT License](LICENSE).
