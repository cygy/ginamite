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
| `kafka_max_attempts`, `kafka_retry_delay` | Attempts to handle a Kafka message and delay in milliseconds between them (see below). |
| `worker_consumers` | Kafka topics consumed by a worker, with their consumer group (see below). |

### Workers and Kafka consumer groups

The messages are consumed through Kafka consumer groups, and a message is committed once it is handled, so a worker
stopped in the middle of a message does not lose it (the message is handled again: handlers must tolerate it).

- The workers of the **same group** share the messages of a topic: each message is handled by one worker. Run several
  identical workers to share the load.
- The workers of **different groups** each receive every message.
- A worker consumes all the topics (`task`, `group_notification`, `user_notification`, `mail`) unless `worker_consumers`
  lists the ones it consumes. The group is the topic by default, so two workers listing the same topic share it:

```yaml
# worker 1: the mails, shared with the worker 2, and the tasks
worker_consumers:
  - topic: mail
  - topic: task
```

```yaml
# worker 2: the mails, shared with the worker 1, and the user notifications
worker_consumers:
  - topic: mail
  - topic: user_notification
```

- A worker cannot ignore some messages of a topic: it would take them away from the other workers of the group. To
  split the work, give each kind of work its own topic.
- The API servers use a group of their own per instance for the `cache` topic, because every instance must receive
  every invalidation.
- A group created for the first time handles the messages sent from then on.

#### Retries and dead letters

A message whose handling fails is handled again, then sent to a **dead-letter topic** instead of blocking the others.

- Failures are retried up to `kafka_max_attempts` times (3 by default), waiting `kafka_retry_delay` milliseconds
  (1000 by default) before the second attempt, then twice as long each time, a minute at most. The partition waits
  meanwhile, which keeps the order of its messages.
- The failed message is then published, as it is, to the topic named like its own with the suffix `_dlq`
  (`<prefix>mail_dlq`), and the consumer goes on. The headers give the reason: `error`, `attempts`, `group`, `topic`,
  `partition` and `offset`. Nothing consumes this topic: read it to inspect the failures, and send the messages again
  to their topic once the cause is fixed.
- A message which can not be decoded is dead-lettered at once, since no retry can fix it. A panic is recovered and
  counts as a failure. A message of an unknown type is only logged.
- The functions of `worker.Handlers` handling a message (`DeleteUserByID`, `SaveIPAddressDetails`, `HandleMail`, etc.)
  return an error, and the custom ones return `(handled bool, err error)`; the plugin `plugins/mongodb` returns the
  errors of its database. An email which can not be sent (SMTP error) fails the same way. The functions must tolerate
  being called again for the same message, and `kafka.Permanent(err)` marks the error which must not be retried. A
  handler of your own (`kafka.MessageConsumer.Action`) follows the same rule.
- Not covered: the functions the plugin lets you plug in to update the documents referencing a user
  (`UpdateDocumentsReferencingTo...`) have no error to return, and a failing lookup of the location of an IP address
  (the ipstack API) is logged but not retried.
- If the producer is not connected, the failed message is dropped and logged as an error.

#### Shutdown and logs

- When the process is interrupted or terminated, the consumers finish the message they are handling, commit it and
  leave their group (30 seconds at most), then the producer sends the messages it still holds, then the plugins close
  their connections. `Start` returns, and the exit code is 0.
- The logs give the topic, the partition, the offset and the size of the messages, never their body, which holds the
  keys of the confirmation links: it is logged at the debug level only.

### Shared services are injected

There is no global service. `ginamite.Server` holds the configuration (`Config`), the translations (`Localization`),
the Kafka client (`Kafka`) and the queue built on it (`Queue`), loads them in `Start`, and shares them with `Server.API`,
`Server.Worker`, `Server.Web` (their fields of the same names) and the plugins (`Connector.Connect` receives the
configuration).

- The API and web servers save them in the Gin context of the requests, and the handlers read them with
  `ctxconfig.Get(c)`, `ctxlocalization.Get(c)` and `ctxqueue.Get(c)`. The functions of the application, such as the ones
  of the storage plugin, are saved the same way: `account.FunctionsOf(c)`, `ctxAuthToken.HooksOf(c)`, etc.
- The web server also saves its pages, `html.Of(c)`, and the worker builds a `mail.Mailer` from its services.
- The code without a request receives what it needs as arguments: for instance the worker builds a `recurring.Scheduler`
  and a `mail.Mailer` from its services.
- The only process-wide state left is the logger `log.Main`, like the default logger of the standard library, and a few
  constants and caches of immutable values.
- In the tests, use the `Set` function of the package, for instance `ctxconfig.Set(c, configuration)`.

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

`plugins/mongodb` opens its own MongoDB session in `Connect`, with no global state: the functions of the API and of the
worker work on the database of the plugin, and the middleware of the plugin saves it in the context of the requests.
Its settings (`Plugin.UserMaxImageSize`, etc., see `Plugin.Settings`) are saved in the context of the requests with
`ctxsettings`. It stores the users, their auth tokens, their notifications and the contacts. Create it with
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
