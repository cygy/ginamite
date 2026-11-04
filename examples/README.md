# Examples

Small projects built with ginamite. Each one is deployed with a Dockerfile, built from the root of the repository.

| Example | Description |
| --- | --- |
| [helloworld-api](helloworld-api) | An API with a single route, `GET /v1/hello`, which answers `{"hello":"world"}`. |
| [helloworld-web](helloworld-web) | A web server (no API, no worker) with a single page, `/en-us/hello`, rendered from an HTML template. |

## helloworld-api

Run it from the root of the repository:

```bash
go run ./examples/helloworld-api -api -name hello -conf-server examples/helloworld-api/conf/server.yaml
curl localhost:8080/v1/hello
```

Or with Docker:

```bash
docker build -f examples/helloworld-api/Dockerfile -t ginamite-helloworld-api .
docker run --rm -p 8080:8080 ginamite-helloworld-api
```

The API loads the `status` feature only (`GET /v1/status`), so it needs no `jwt.secret`.

## helloworld-web

Run it from its own directory, since the paths of the configuration are relative to the working directory:

```bash
cd examples/helloworld-web
go run . -web -name helloworld-web -conf-server conf/server.yaml -web-routes conf/routes.yaml
curl -L localhost:8080
```

Or with Docker:

```bash
docker build -f examples/helloworld-web/Dockerfile -t ginamite-helloworld-web .
docker run --rm -p 8080:8080 ginamite-helloworld-web
```

The page is described in [conf/routes.yaml](helloworld-web/conf/routes.yaml), its templates are in `templates/pages`, and its title is translated by `localization/en-us.all.json`.
