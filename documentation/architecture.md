# Architecture

An application is one binary which runs as one of three servers, chosen by the command line (`-api`, `-worker`,
`-web`). They share the configuration and the packages of `common/`, and exchange messages through a broker.

```text
browser ──► web server ──► API server ──► database
                              │              ▲
                              │ messages     │
                              ▼              │
                           broker ──► worker server
```

| Server | Role | Reads |
| --- | --- | --- |
| API (`api/`) | The JSON REST API: accounts, authentication, notifications, contacts. Sends messages for what can wait. | The database, through the storages. |
| Worker (`worker/`) | Consumes the messages: tasks, emails, notifications. Runs the recurring tasks. See [Worker](worker.md). | The database, through the storages. |
| Web (`web/`) | The HTML pages, localized, which call the API. See [Web server](web-server.md). | The API only. |

## Plugins

The core depends on no database nor broker: a **plugin** provides them, registered with `server.AddPlugin`.

- A **database plugin** (`plugins/mongodb`) sets the storages and the hooks of the API and of the worker, see
  [Database plugins](database-plugins.md).
- A **broker plugin** (`plugins/kafka`) connects the servers, see [Message broker and Kafka](kafka.md).

The storages are interfaces taking the `context.Context` and returning errors. A feature whose storage is not set is
disabled, so an application can use the core without a database.

### Shared services are injected

There is no global service. `ginamite.Server` holds the configuration (`Config`), the translations (`Localization`),
the message broker (`Broker`, set by a plugin) and the queue built on it (`Queue`), loads them in `Run`, and shares them with `Server.API`,
`Server.Worker`, `Server.Web` (their fields of the same names) and the plugins (`Connector.Connect` receives the
configuration).

- The API and web servers save them in the Gin context of the requests, and the handlers read them with
  `ctxkey.Config.Get(c)`, `ctxkey.Localization.Get(c)` and `ctxkey.Queue.Get(c)`. The storages of the application, such as the ones
  of the storage plugin, are saved the same way: `account.StorageOf(c)`, `ctxAuthToken.HooksOf(c)`, etc.
- The web server also saves its pages, `html.Of(c)`, and the worker builds a `mail.Mailer` from its services.
- The code without a request receives what it needs as arguments: for instance the worker builds a `recurring.Scheduler`
  and a `mail.Mailer` from its services.
- The only process-wide state left is the logger `log.Main`, like the default logger of the standard library, and a few
  constants and caches of immutable values.
- In the tests, use `Set` on the key, for instance `ctxkey.Config.Set(c, configuration)`. `ctxkey.New[T](name)` creates
  the key of a service of your own.

## Packages

| Directory | Content |
| --- | --- |
| `.` | `Server`, `Plugin`, `Connector` and `Run`, which sets up the configuration and runs the chosen server. |
| `api/` | The API server: routes, middlewares, request validation, responses, authentication cache. |
| `worker/` | The worker server: message consumers, notifications, mails, async and recurring tasks. |
| `web/` | The web server: HTML routes, pages, cookies, calls to the API, `robots.txt`. |
| `common/` | Shared packages: configuration, flags, logger, localization, templates, broker interfaces, JWT, timezones. |
| `plugins/` | The plugins: the database (`mongodb`) and the message broker (`kafka`). |
| `examples/` | Small projects built with ginamite, deployed with a Dockerfile. See [examples](../examples/README.md). |

The packages under an `internal/` directory are the implementation of the servers: the HTTP server, the handlers of
the built-in routes and tasks, the dead letters tool of Kafka. An application can not import them; the other packages
are the API of ginamite, which an application and a third-party plugin use.
