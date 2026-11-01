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
- A message broker, through a plugin, to send messages between the servers. The only plugin for now is `plugins/kafka`.

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
	"github.com/cygy/ginamite/plugins/kafka"
	"github.com/cygy/ginamite/plugins/mongodb"
)

func main() {
	server := ginamite.NewServer()
	server.AddPlugin(mongodb.NewPlugin())
	server.AddPlugin(kafka.NewPlugin())

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
| `-routes` | Path to the routes file, required by the web server only. |
| `-help` | Print the help. |

The functions of `api.Server.Functions` and `worker.Server.Functions` are provided by the plugins or by the application.
An API only exposes the default routes whose functions are defined: without `CreateContact` there is no contact route,
and an API without any function only has `/status`, `/locales`, `/timezones`, `/notifications` and `/terms`. A route
can still be enabled in `RouterConfigureDefaultRoutes`.

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
| `smtp`, `sender_*` | Sending of the emails. |
| `facebook`, `google` | Sign-in with a social network. |
| `recurring_tasks`, `run_start_up_tasks` | Tasks executed by the workers. |
| `metrics`, `rate_limits` | Prometheus endpoint and limits of the requests (see Security and Metrics). |
| `broker` | The message broker: `service` (for instance `kafka`, the plugin which connects), `host` and `port`. |
| `kafka` | Connection to Kafka (version, TLS, SASL), prefix of the topics, attempts and delay of the retries, creation of the topics, statistics (see below). |
| `worker_consumers` | Kafka topics consumed by a worker, with their consumer group (see below). |

### Security

- **Passwords** are hashed with Argon2id (PHC format, random salt, constant-time comparison). The legacy SHA-512 crypt
  hashes (`$6$`) are still accepted, and replaced by an Argon2id hash after a successful login when the database plugin
  defines `UpdateEncryptedPasswordByUserID` (a warning is logged at start-up if it is undefined). Logging in with an
  unknown email costs as much time as with a wrong password.
- **JWT**: only HS256, HS384 and HS512 are accepted, and the `exp` claim is required. The tokens are never logged.
  A `jwt` secret shorter than 32 characters logs a warning.
- **Rate limiting**: the sign-in, registration and forgot-password routes are limited per client IP (token bucket in the
  memory of each instance, so the limit applies per API instance). A refused request gets a `429` and a `Retry-After`
  header.

  ```yaml
  rate_limits:
    authentication:  { requests: 10, seconds: 60 }
    registration:    { requests: 5,  seconds: 60 }
    forgot_password: { requests: 5,  seconds: 60 }
  server:
    trusted_proxies: ["10.0.0.0/8"]  # the proxies allowed to set X-Forwarded-For
  ```

  No proxy is trusted by default: the client IP, hence the limit, is the one of the connection. Set
  `server.trusted_proxies` behind a reverse proxy, otherwise every client has the IP of the proxy.

### Metrics

With `metrics.address`, the server listens there for Prometheus and serves `/metrics`, on its own port: do not expose
it to the Internet. It is disabled when the address is empty.

```yaml
metrics:
  address: ":9100"
```

- `http_requests_total` and `http_request_duration_seconds_total`, by method, route pattern and status (API).
- `kafka_messages_{received,handled,retried,dead_lettered,duplicates,sent,dropped}_total`, `kafka_send_errors_total`,
  `kafka_outbox_messages` and `kafka_consumer_lag_messages{topic,partition}`.

A custom broker appears in the metrics if it implements `metrics.Collector` (`common/metrics`).

### Workers and Kafka consumer groups

The messages are consumed through Kafka consumer groups, and a message is committed once it is handled, so a worker
stopped in the middle of a message does not lose it (the message is handled again: handlers must tolerate it).

- The workers of the **same group** share the messages of a topic: each message is handled by one worker. Run several
  identical workers to share the load.
- The workers of **different groups** each receive every message.
- A worker consumes all the topics (`task`, `task_location`, `task_user`, `task_registration`, `group_notification`,
  `user_notification`, `mail`) unless `worker_consumers` lists the ones it consumes. The group is the topic by default,
  so two workers listing the same topic share it. The tasks have a topic per kind: `task_location` (the location of the
  IP addresses, slow because of ipstack), `task_user` (deleting, disabling, enabling a user, social networks, ordered by
  user), `task_registration`, and `task` for your own tasks (`Queue.CreateTask`). All of them handle the built-in
  tasks, so the ones still waiting in `task` after an upgrade are handled. A worker which only runs the location can
  list only `task_location`:

```yaml
# worker 1: the mails, shared with the worker 2, and the users
worker_consumers:
  - topic: mail
  - topic: task_user
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

- Failures are retried up to `kafka.max_attempts` times (3 by default), waiting `kafka.retry_delay` milliseconds
  (1000 by default) before the second attempt, then twice as long each time, a minute at most. The partition waits
  meanwhile, which keeps the order of its messages.
- The failed message is then published, as it is, to the topic named like its own with the suffix `_dlq`
  (`<prefix>mail_dlq`), and the consumer goes on. The headers give the reason: `error`, `attempts`, `group`, `topic`,
  `partition` and `offset`. No worker consumes this topic: use the command below to inspect the failures and to send
  the messages again once the cause is fixed.
- A message which can not be decoded is dead-lettered at once, since no retry can fix it. A panic is recovered and
  counts as a failure. A message of an unknown type is only logged.
- The functions of `worker.Handlers` handling a message (`DeleteUserByID`, `SaveIPAddressDetails`, `HandleMail`, etc.)
  return an error, and the custom ones return `(handled bool, err error)`; the plugin `plugins/mongodb` returns the
  errors of its database. An email which can not be sent (SMTP error) fails the same way. The functions must tolerate
  being called again for the same message, and `broker.Permanent(err)` marks the error which must not be retried. A
  handler of your own (`broker.MessageConsumer.Action`) follows the same rule.
- The hooks of the plugin `plugins/mongodb` which update the documents referencing a user
  (`UpdateDocumentsReferencingTo...`) return an error, and the message is handled again when they fail, so they must be
  safe to run again. The temporary failures of the ipstack API (network, `429`, `5xx`) are retried the same way; an
  unknown location or a client error is logged and not retried.
- If the producer is not connected, the failed message is dropped and logged as an error.

##### Dead letters command

The binary lists, replays and purges the dead letters of a topic, without starting any server. It only needs the
server configuration (`-conf-server`) and the name of the process, and the sections `broker` and `kafka` of the configuration:

```bash
myapp -name dlq -conf-server server.yaml -dlq list -dlq-topic mail
myapp -name dlq -conf-server server.yaml -dlq replay -dlq-topic mail -dlq-limit 10
myapp -name dlq -conf-server server.yaml -dlq purge -dlq-topic mail
```

- `-dlq-topic` is the topic without the prefix. `-dlq-limit` limits the count of dead letters (all if absent).
- `list` prints where each dead letter is, when it failed, the topic, group and number of attempts, the error and the
  size of the message; `-dlq-body` prints the body too, which holds secrets.
- `replay` sends the dead letters again to the topic they came from, partition by partition, and stops at the first
  failure. A message replayed again and failing again is dead-lettered again. The replayed message has a header
  `replayed-from`.
- `purge` forgets the dead letters without sending them again.
- The position of the command is saved in Kafka, in the consumer group `<prefix><topic>_dlq_tool`: a dead letter
  replayed or purged is never visited again, and `list` shows only the others. The messages themselves stay in the
  topic until their retention ends (`dead_letter_retention_hours`).

#### Shutdown and logs

- When the process is interrupted or terminated, the consumers finish the message they are handling, commit it and
  leave their group (30 seconds at most), then the producer sends the messages it still holds, then the plugins close
  their connections. `Start` returns, and the exit code is 0.
- The logs give the topic, the partition, the offset and the size of the messages, never their body, which holds the
  keys of the confirmation links: it is logged at the debug level only.

#### Ordering and duplicates

- The messages about a same user (`user_id` in the payload, else `email`) are sent with that key: with several
  partitions they go to the same one, so a worker handles them in the order they were sent. A payload can choose its
  key by implementing `queue.Keyed`. The messages without key go to any partition.
- Each message carries an identifier in the `message-id` header, kept by the dead letters. A group ignores a message
  it already handled (counted in `Stats().Duplicates`), through `Server.Deduplicator`: in memory by default (the last
  10 000 messages, which covers the redeliveries of a rebalance), and in the `processedMessages` collection with the
  MongoDB plugin, shared by the workers and kept after a restart (the documents expire after 7 days; the TTL index is
  created by the worker at start up). A consumer whose handler gives the same result twice, such as the
  invalidation of a cache, sets `Idempotent: true` and costs nothing.
- A message is remembered once it is handled: a crash between the handling and this record handles it twice. The
  handlers of the database (deleting, disabling, enabling a user, saving a location) are idempotent. The mails and the
  notifications are not: a duplicate is possible after such a crash, which is preferred to losing the message.

### Using another broker

The servers only depend on the interfaces of `common/broker`: a `Producer` (`SendMessageToTopic`, and
`SendKeyedMessageToTopic` for the ordering), the `MessageConsumer`s of the topics, and `Broker` (`Start(consumers)`,
`Close()`). The plugin `plugins/kafka` is the implementation for Kafka: when it is configured, it sets the Kafka client as
the broker, see `server.UseBroker`, if the configuration has `broker.service: kafka` with a host and a port. Without a broker, the messages are not
sent and an error is logged for each. To use another broker, implement `broker.Broker` and make a plugin, or the
application before `Start`, call `server.UseBroker(b)`: the queue of the API and of the worker send through it, and
`Start` runs it with the consumers of the chosen server. A broker which implements `broker.DeduplicatorUser` receives
the deduplicator of the server before it starts. A plugin which implements `ginamite.Commander` can run a command of the
command line instead of a server: `plugins/kafka` runs the one on the dead letters. The retries, the dead letters and the duplicates are the business of the
broker: Kafka does them itself (see above), another one can rely on its own mechanisms. `broker.Permanent(err)` marks
an error which must not be retried. `broker.NewMemory()` is a broker for the tests: it records the messages and
delivers them to the consumers at once.

### Kafka connection and topics

The address of the broker is in the section `broker`, with `service: kafka`; the other settings are in the section
`kafka` of the configuration file; `kafka.topics_prefix` is prepended to the names of the topics and of the consumer groups.

```yaml
broker:
  service: kafka
  host: kafka.example.com
  port: 9093

kafka:
  topics_prefix: myapp_
  max_attempts: 3
  retry_delay: 1000            # milliseconds, doubled at each attempt
  version: 3.6.0               # the version of the brokers, the one of the client library if absent
  tls:
    enabled: true
    ca_file: /etc/kafka/ca.pem  # the system authorities if absent
    cert_file: /etc/kafka/client.pem   # for a mutual authentication
    key_file: /etc/kafka/client.key
  sasl:
    mechanism: SCRAM-SHA-512   # PLAIN, SCRAM-SHA-256 or SCRAM-SHA-512
    username: myapp
    password: secret
  topics:
    create: true               # create the missing topics when a server starts
    partitions: 6
    replication_factor: 3
    retention_hours: 168
    dead_letter_retention_hours: 720
  stats_interval: 60           # seconds between two logs of the statistics, none if absent
```

- **Producer.** A message is acknowledged by all the in-sync replicas, and sent again if it fails without being
  duplicated nor reordered (idempotent producer). The messages are sent in the order of the calls; the ones sent before
  the producer is connected wait for it (1024 at most, the others are dropped and logged).
- **Topics.** With `topics.create`, the servers create the topics, and the topics of their dead letters, which do not
  exist yet, and never change the ones which exist. Without it, create them yourself, or let the brokers do it.
- **Partitions.** The workers of a group share the partitions of a topic: with one partition, a single worker
  works and the others wait. The messages are only ordered within a partition, and the messages have no key, so with
  several partitions two messages sent one after the other may be handled in any order.
- **Statistics.** `Client.Stats()` gives the counters (messages received, handled, retried, dead-lettered, sent, refused,
  dropped) and the lag (messages waiting) of each partition. `kafka.stats_interval` logs them.

### Shared services are injected

There is no global service. `ginamite.Server` holds the configuration (`Config`), the translations (`Localization`),
the message broker (`Broker`, set by a plugin) and the queue built on it (`Queue`), loads them in `Start`, and shares them with `Server.API`,
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
| `worker/` | The worker server: message consumers, notifications, mails, async and recurring tasks. |
| `web/` | The web server: HTML routes, pages, cookies, calls to the API, `robots.txt`. |
| `common/` | Shared packages: configuration, flags, logger, localization, templates, broker interfaces, JWT, timezones. |
| `plugins/` | The plugins: the database (`mongodb`) and the message broker (`kafka`). |
| `examples/` | Small projects built with ginamite, deployed with a Dockerfile. See [examples](examples/README.md). |

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
