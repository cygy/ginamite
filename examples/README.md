# Examples

Small projects built with ginamite. Each one is deployed with a Dockerfile, built from the root of the repository.

| Example | Description |
| --- | --- |
| [hello](hello) | An API with a single route, `GET /v1/hello`, which answers `{"hello":"world"}`. |

## hello

Run it from the root of the repository:

```bash
go run ./examples/hello -api -name hello -conf-server examples/hello/conf/server.yaml
curl localhost:8080/v1/hello
```

Or with Docker:

```bash
docker build -f examples/hello/Dockerfile -t ginamite-hello .
docker run --rm -p 8080:8080 ginamite-hello
```

Change `jwt.secret` in [conf/server.yaml](hello/conf/server.yaml) before any real deployment.
