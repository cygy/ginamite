# Examples

Small projects built with ginamite. Each one is deployed with a Dockerfile, built from the root of the repository.

| Example | Description |
| --- | --- |
| [helloworld-api](helloworld-api) | An API with a single route, `GET /v1/hello`, which answers `{"hello":"world"}`. |
| [helloworld-web](helloworld-web) | A web server (no API, no worker) with a single page, `/en-us/hello`, rendered from an HTML template. |

## helloworld-api

Run it from the root of the repository:

```bash
export JWT_SECRET=$(openssl rand -hex 32)
go run ./examples/helloworld-api -api -name hello -conf-server examples/helloworld-api/conf/server.yaml
curl localhost:8080/v1/hello
```

Or with Docker:

```bash
docker build -f examples/helloworld-api/Dockerfile -t ginamite-helloworld-api .
docker run --rm -p 8080:8080 -e JWT_SECRET=$(openssl rand -hex 32) ginamite-helloworld-api
```

The secret of the tokens is read from the environment variable `JWT_SECRET` (`!env`): the configuration holds no secret. See the README of ginamite to read it from a file.

## helloworld-web

Run it from its own directory, since the paths of the configuration are relative to the working directory:

```bash
cd examples/helloworld-web
export JWT_SECRET=$(openssl rand -hex 32)
go run . -web -name helloworld-web -conf-server conf/server.yaml -routes conf/routes.yaml
curl -L localhost:8080
```

Or with Docker:

```bash
docker build -f examples/helloworld-web/Dockerfile -t ginamite-helloworld-web .
docker run --rm -p 8080:8080 -e JWT_SECRET=$(openssl rand -hex 32) ginamite-helloworld-web
```

The page is described in [conf/routes.yaml](helloworld-web/conf/routes.yaml), its templates are in `templates/pages`, and its title is translated by `localization/en-us.all.json`.
