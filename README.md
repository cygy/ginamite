# ginamite

A Go framework to build an application made of three servers, built on [Gin](https://github.com/gin-gonic/gin):

- an **API server**: a JSON REST API with the accounts, the authentication (JWT, Facebook and Google sign-in), the
  notifications and the contacts;
- a **worker server**: asynchronous tasks consumed from Kafka, and recurring tasks (cleaning of the expired tokens, of
  the inactive accounts, sitemap generation, etc.);
- a **web server**: HTML pages rendered from templates, localized, which call the API.

The three servers share one binary: the command line arguments choose which one runs. The storage is provided by a
plugin, so the core does not depend on a database.

## Requirements

- Go 1.26 or later.
- A database, through a plugin. The only plugin for now is `plugins/mongodb`.
- Kafka, to send messages between the servers (optional).

## Installation

```bash
go get github.com/cygy/ginamite
```

## Usage

Create a server, register the plugins, add the functions and routes of the application, then start it:

```go
package main

import (
	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/plugins/mongodb"
)

func main() {
	server := ginamite.NewServer()
	server.AddPlugin(mongodb.NewPlugin())

	// Customize the servers before starting: server.API, server.Worker and server.Web.

	server.Start()
}
```

Run the same binary once per server:

```bash
go run . -api    -name api    -env production -conf-server conf/server.yaml
go run . -worker -name worker -env production -conf-server conf/server.yaml
go run . -web    -name web    -env production -conf-server conf/server.yaml -routes conf/routes.yaml
```

| Argument | Description |
| --- | --- |
| `-api`, `-worker`, `-web` | The server to run. If several are given, the API wins over the worker, which wins over the web server. |
| `-name` | Name of the server, added to the logs. |
| `-env` | Environment: `development` (default), `production` or `test`. |
| `-version` | Version: `final` (default), `beta` or `alpha`. |
| `-conf-server` | Path to the server configuration file (YAML). |
| `-conf-app` | Path to the application configuration file, see `ApplicationConfiguration`. |
| `-routes` | Path to the routes file of the web server. |
| `-help` | Print the help. |

## Configuration

The server configuration is a YAML file, see `common/config.Configuration` for the complete list of settings and their
default values. The main sections are:

| Section | Content |
| --- | --- |
| `server` | Host, port, gzip, SSL and log format of the HTTP server. |
| `hosts` | Public hosts of the API, the web server and the static files. |
| `database` | `service` (for instance `mongodb`), `host`, `port`, `database`, `username`, `password`, `time_out` in seconds. |
| `jwt`, `cookie` | Settings of the authentication tokens and of the cookie. |
| `locales`, `default_locale`, `localization_path`, `templates_path` | Languages, translations and templates. |
| `services` | Addresses of the other services, such as `kafka`. |
| `smtp`, `sender_*` | Sending of the emails. |
| `facebook`, `google` | Sign-in with a social network. |
| `recurring_tasks`, `run_start_up_tasks` | Tasks executed by the workers. |

### Configuration is injected

There is no global configuration. `ginamite.Server.Config` is loaded by `Start` from the configuration file, then shared
with `Server.API`, `Server.Worker`, `Server.Web` (their `Config` field) and the plugins (`Connector.Connect` receives it).

- The handlers of the API and of the web server read it from the Gin context with `ctxconfig.Get(c)`; the servers
  save it there with the middleware `ctxconfig.Inject`.
- The packages without a request, such as the tasks of the worker, receive the settings they need as arguments.
- In the tests, set it with `ctxconfig.Set(c, configuration)`.

## Layout

| Directory | Content |
| --- | --- |
| `.` | `Server`, `Plugin`, `Connector` and `Start`, which sets up the configuration and starts the chosen server. |
| `api/` | The API server: routes, middlewares, request validation, responses, authentication cache. |
| `worker/` | The worker server: Kafka consumers, notifications, mails, async and recurring tasks. |
| `web/` | The web server: HTML routes, pages, cookies, calls to the API, `robots.txt`. |
| `common/` | Shared packages: configuration, flags, logger, localization, templates, Kafka, JWT, timezones. |
| `plugins/` | The database plugins. |

## Database plugins

The core does not depend on a database. A database plugin, such as `plugins/mongodb`, provides the storage through
the extension points below, and is registered with `server.AddPlugin`.

| Extension point | Role |
| --- | --- |
| `ginamite.Plugin.Configure` | Sets the functions of the API (`api.Server.Functions`) and of the worker (`worker.Server.Functions`). |
| `ginamite.Connector` (optional) | `Connect` opens the connection before the plugins are configured, `Close` closes it when the server stops. Called only for the API and worker servers. |
| `api.Server.Middlewares` | Gin middlewares added before the routes, for instance to put the database in the context of the requests. |
| `api.Handlers.IsValidID` | Reports whether an identifier in a URL is well-formed. It depends on the database (an ObjectID for MongoDB, a UUID or an integer for SQL). By default every identifier is accepted. |

A plugin for another database implements the same extension points. `plugins/mongodb` is the reference.

### MongoDB

`plugins/mongodb` stores the users, their auth tokens, their notifications and the contacts. Create it with
`mongodb.NewPlugin()`, set the properties that differ from the defaults (abilities, size of the images, etc.), then
register it. Its hooks let the application add its own collections, indexes and default documents, and update the
documents which reference the users.

## Development

The `Makefile` lists its targets with `make help`:

```bash
make check   # build, formatting, go vet, golangci-lint and the tests with the race detector
make cover   # tests and total coverage
make fmt     # gofumpt and goimports
```

The code is formatted with `gofumpt`, and `golangci-lint` must report no issue. Every source file has its test file.

### Release

The release targets check the code, bump the `VERSION` file, commit it, tag it and push:

```bash
make bump_patch   # or bump_minor, bump_major
```

## License

[MIT](LICENSE)
