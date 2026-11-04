# ginamite

[![CI](https://github.com/cygy/ginamite/actions/workflows/ci.yml/badge.svg)](https://github.com/cygy/ginamite/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cygy/ginamite.svg)](https://pkg.go.dev/github.com/cygy/ginamite)
[![Go Report Card](https://goreportcard.com/badge/github.com/cygy/ginamite)](https://goreportcard.com/report/github.com/cygy/ginamite)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Go framework to build an application made of three servers, built on [Gin](https://github.com/gin-gonic/gin):

- an **API server**: a JSON REST API with the accounts, the authentication (JWT, Facebook and Google sign-in), the
  notifications and the contacts;
- a **worker server**: asynchronous tasks consumed from Kafka, and recurring tasks (cleaning of the expired tokens, of
  the inactive accounts, sitemap generation, etc.);
- a **web server**: HTML pages rendered from templates, localized, which call the API.

The three servers share one binary: the command line arguments choose which one runs. The storage is provided by a
plugin, so the core does not depend on a database.

## Why ginamite

Most web applications need the same foundations before their first feature: accounts, sign-in, tokens, emails,
notifications, translations, background tasks. ginamite provides them, tested and secured, so that an application only
writes what is its own.

| ginamite provides | The application writes |
| --- | --- |
| The servers, their start-up, their shutdown, the logs and the metrics. | Its routes: `AddCustomRoutes` on the API and the web server. |
| The default routes of the API: accounts, authentication, notifications, contacts, terms (see [API features](documentation/api-features.md)). | The configuration file: the features it enables, its hosts, its secrets. |
| The storage of these routes, through the plugin `plugins/mongodb`. | Its own storage, for its own data. |
| The messages between the servers, through the plugin `plugins/kafka`, and the worker handling them. | Its own tasks, if any, sent with `Queue.CreateTask`. |
| The web pages described by a routes file, localized, with their templates. | Its templates and translations. |

## Requirements

- Go 1.26 or later.
- Optional: a database, through a plugin, for the default routes which store data. The only plugin for now is
  `plugins/mongodb`.
- Optional: a message broker, through a plugin, to send messages between the servers. The only plugin for now is
  `plugins/kafka`. Without a broker, the messages are not sent and an error is logged for each.

## Installation

```bash
go get github.com/cygy/ginamite
```

## Quick start

The smallest API, an [example](examples/README.md) which needs neither database nor broker:

```go
func main() {
	server := ginamite.NewServer()

	server.API.AddCustomRoutes(func(group *gin.RouterGroup, _ route.Middlewares, _ bool, _ func(bool, []string) gin.HandlerFunc) {
		group.GET("/hello", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"hello": "world"}) })
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

```bash
go run ./examples/helloworld-api -api -name hello -conf-server examples/helloworld-api/conf/server.yaml
curl localhost:8080/v1/hello
```

## Usage

Create a server, register the plugins, add the functions and routes of the application, then run it:

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/plugins/kafka"
	"github.com/cygy/ginamite/plugins/mongodb"
)

func main() {
	server := ginamite.NewServer()
	server.AddPlugin(mongodb.NewPlugin())
	server.AddPlugin(kafka.NewPlugin())

	// Customize the servers before running: server.API, server.Worker and server.Web.

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
```

`Run` serves until its context is done, here on SIGINT or SIGTERM. It returns an error, without exiting the process, if
the arguments, the configuration or a plugin are invalid, or if the server stops before its context is done; the
connections already opened are closed first. It returns nil once the server stopped gracefully, or at once if `-help`
is given.

Run the same binary once per server:

```bash
go run . -api    -name api    -env production -conf-server conf/server.yaml
go run . -worker -name worker -env production -conf-server conf/server.yaml
go run . -web    -name web    -env production -conf-server conf/server.yaml -web-routes conf/routes.yaml
```

| Argument | Description |
| --- | --- |
| `-api`, `-worker`, `-web` | The server to run. If several are given, the API wins over the worker, which wins over the web server. |
| `-name` | Name of the server, added to the logs. |
| `-env` | Environment: `development` (default), `production` or `test`. |
| `-version` | Version: `final` (default), `beta` or `alpha`. |
| `-conf-server` | Path to the server configuration file (YAML). |
| `-conf-app` | Path to the application configuration file, see `ApplicationConfiguration`. |
| `-web-routes` | Path to the web routes file: required by the web server, optional for the API and the worker, whose URLs of the website are then empty. Loaded once, shared by the three servers and the plugins, see [Web routes](documentation/web-routes.md). |
| `-help` | Print the help, with the arguments added by the plugins. |

The arguments `-dlq*`, added by `plugins/kafka`, run a command on the dead letters of Kafka instead of a server, see
[Message broker and Kafka](documentation/kafka.md#dead-letters-command).

## Documentation

| Page | Content |
| --- | --- |
| [Architecture](documentation/architecture.md) | The three servers, the plugins, the injected services, the packages. |
| [Configuration](documentation/configuration.md) | The configuration file and its sections. |
| [API features](documentation/api-features.md) | The default routes of the API, by feature. |
| [Web routes](documentation/web-routes.md) | Build the URLs of the website from the API, the worker and the plugins. |
| [Custom routes](documentation/custom-routes.md) | Add the routes of the application to the API and to the web server. |
| [Web server](documentation/web-server.md) | The routes file, the templates and the translations. |
| [Worker](documentation/worker.md) | The messages consumed, the recurring tasks, the storage and the hooks. |
| [Message broker and Kafka](documentation/kafka.md) | The broker, the Kafka plugin, the consumer groups and the dead letters. |
| [Payments with Stripe](documentation/stripe.md) | The Stripe plugin: checkout sessions, payment intents, the webhook and its store. |
| [Database plugins](documentation/database-plugins.md) | The interfaces a database plugin implements, and the MongoDB plugin. |
| [Security](documentation/security.md) | Passwords, JWT, rate limiting, trusted proxies and secrets. |
| [Observability](documentation/observability.md) | The Prometheus metrics and the logs. |
| [Tests](documentation/testing.md) | End-to-end tests of an API, and the logs in the tests. |
| [Examples](examples/README.md) | Small projects built with ginamite. |

## Development

The `Makefile` lists its targets with `make help`:

```bash
make check   # build, formatting, go vet, golangci-lint and the tests with the race detector
make cover   # tests and total coverage
make fmt     # gofumpt and goimports
```

The code is formatted with `gofumpt`, `golangci-lint` must report no issue, and the code is tested along with it. The
CI runs `make check` and `govulncheck` on every pull request. See [CONTRIBUTING.md](CONTRIBUTING.md) to contribute, and
[SECURITY.md](SECURITY.md) to report a vulnerability.

### Release

The release targets check the code, bump the `VERSION` file, commit it, tag it and push:

```bash
make bump_patch   # or bump_minor, bump_major
```

## License

[MIT](LICENSE)
