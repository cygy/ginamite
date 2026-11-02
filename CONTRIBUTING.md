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
make check       # build, formatting, go vet, golangci-lint and the tests with the race detector
make fmt         # format the code
make cover       # tests and total coverage
make fuzz        # each fuzz test for 30 seconds, FUZZTIME=5m for longer
make integration # the tests needing a real MongoDB and Kafka, see below
```

A failing input found by `make fuzz` is saved under `testdata/fuzz`: commit it with the fix, it becomes a regression
test.

`make integration` needs a MongoDB and a Kafka reachable at `localhost`, `MONGO_HOST`/`MONGO_PORT` and
`KAFKA_HOST`/`KAFKA_PORT` otherwise:

```bash
docker run -d -p 27017:27017 mongo:9
docker run -d -p 9092:9092 \
  -e KAFKA_NODE_ID=0 -e KAFKA_PROCESS_ROLES=controller,broker \
  -e KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=0@localhost:9093 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
  -e KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  -e KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 -e KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
  -e CLUSTER_ID=ciWo7IWazngRchmPES6q5A \
  apache/kafka:4.4.0
```

The low replication factors are only needed because this is a single broker: a real cluster uses its own defaults.
The CI runs the same images as a separate job.

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
