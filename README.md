# ginamite

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
| The default routes of the API: accounts, authentication, notifications, contacts, terms (see API features). | The configuration file: the features it enables, its hosts, its secrets. |
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

	if err := server.Run(); err != nil {
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
	"log"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/plugins/kafka"
	"github.com/cygy/ginamite/plugins/mongodb"
)

func main() {
	server := ginamite.NewServer()
	server.AddPlugin(mongodb.NewPlugin())
	server.AddPlugin(kafka.NewPlugin())

	// Customize the servers before running: server.API, server.Worker and server.Web.

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
```

`Run` returns an error, without exiting the process, if the arguments, the configuration or a plugin are invalid, or
if the server stops for another reason than SIGINT or SIGTERM; the connections already opened are closed first. It
returns nil once the server stopped gracefully, or at once if `-help` is given.

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

The arguments `-dlq*` run a command on the dead letters of Kafka instead of a server, see
[Message broker and Kafka](documentation/kafka.md#dead-letters-command).

## Custom routes

The routes of the application are added before `Run`, by a function called when the server sets up its routes.

**API.** The function receives the group of the version of the API (`/v1`), the built-in middlewares, whether the
environment is the production, and a function returning the middleware which checks the rights of the user:

```go
server.API.AddCustomRoutes(func(group *gin.RouterGroup, handlers route.Middlewares, isProd bool, rights func(admin bool, abilities []string) gin.HandlerFunc) {
	group.GET("/public", public)
	group.GET("/me", handlers.MustBeAuthenticated, handlers.MustAuthenticationBeValid, me)
	group.DELETE("/items/:id", handlers.MustBeAuthenticated, rights(false, []string{"manage_items"}), deleteItem)
})
```

The middlewares which authenticate need `jwt.secret`, see API features.

**Web server.** The function is called once per locale, with the group of the locale (`/en-us`). `getRelativePath`
returns, for the key of a route of the routes file, its path relative to the group, whether the page is partial, the
middleware saving the key of the page in the context, and whether the route exists for the locale:

```go
server.Web.AddCustomRoutes(func(g *gin.RouterGroup, locale string, handlers route.Middlewares, getRelativePath func(key string) (string, bool, gin.HandlerFunc, bool)) {
	if path, _, savePageKey, ok := getRelativePath("index"); ok {
		g.GET(path, savePageKey, func(c *gin.Context) {
			page := html.Of(c).PageByKey(c, "index")
			page.Content.Title = page.Title
			page.RenderOk(c)
		})
	}
})
```

## Web server

The pages are described by the routes file (`-routes`), rendered with the Go templates of `templates_path`, and
translated with the files of `localization_path`. See the example [helloworld-web](examples/helloworld-web).

```yaml
default:                      # the properties of every page, overridden by each route
  template:
    decorator: decorator.html # the layout, which includes the main template with {{.IncludeMainTemplate}}
    main: main.html
routes:
  index:                      # the key of the route, used by getRelativePath and PageByKey
    title: hello.title        # a key of the translations
    paths:
      en-us: "/{{.Locale}}/hello"
    sitemap: { included: true, priority: 0.5, update_frequency: weekly }
```

- The templates are the files of the sub-directories of `templates_path` (for instance `templates/pages/`), named by
  their file name. A page also has `meta`, `scripts`, `access` and `partial`, see `web/html.Page`.
- The translations are in `<localization_path>/<locale>.all.json`, one file per locale of `locales`, which lists the
  locales served.
- The web server serves `robots.txt` (`robots_file_path`) and the error pages, whose keys are `error.not_found`,
  `error.internal_server`, `error.unauthorized`, etc. Without a locale, a request is redirected to the default locale.
- The section `async` of the routes file forwards calls of the browser to the API (`api.host`, `api.port`).

## End-to-end tests

The package `api/e2e` tests the endpoints of a running API: each request is sent over HTTP, and the status code, the
status, the error and the renewed auth token of its response are compared with the expected ones.

```go
var api = e2e.FromEnv("v1") // http://$API_HOST:$API_PORT/v1

func TestLogin(t *testing.T) {
	req := api.NewRequest(t)
	req.Method = http.MethodPost
	req.Endpoint = "/auth/login"
	req.Body = e2e.H{"identifier": "john", "password": "wrong"}
	req.Unauthorized()
}
```

## Configuration

The server configuration is a YAML file, see `common/config.Configuration` for the complete list of settings and their
default values. The main sections are:

| Section | Content |
| --- | --- |
| `server` | Host, port, gzip, SSL and log format of the HTTP server. |
| `hosts` | Public hosts of the API, the web server and the static files. |
| `database` | `service` (for instance `mongodb`), `host`, `port`, `database`, `username`, `password`, `time_out` in seconds. |
| `api_features` | The default routes of the API, by feature (see API features). |
| `jwt`, `cookie` | Settings of the authentication tokens and of the cookie. |
| `locales`, `default_locale`, `localization_path`, `templates_path` | Languages, translations and templates. |
| `smtp`, `sender_*` | Sending of the emails. |
| `facebook`, `google` | Sign-in with a social network. |
| `recurring_tasks`, `run_start_up_tasks` | Tasks executed by the workers. |
| `metrics`, `rate_limits` | Prometheus endpoint and limits of the requests (see Security and Metrics). |
| `broker` | The message broker: `service` (for instance `kafka`, the plugin which connects), `host` and `port`. |
| `kafka` | Connection to Kafka (version, TLS, SASL), prefix of the topics, retries, creation of the topics, statistics. |
| `worker_consumers` | Kafka topics consumed by a worker, with their consumer group. |

The sections `broker`, `kafka` and `worker_consumers`, and the way the workers consume the messages, are described in
[Message broker and Kafka](documentation/kafka.md).

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

### Shared services are injected

There is no global service. `ginamite.Server` holds the configuration (`Config`), the translations (`Localization`),
the message broker (`Broker`, set by a plugin) and the queue built on it (`Queue`), loads them in `Run`, and shares them with `Server.API`,
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

### API features

The default routes of the API are grouped by feature, chosen with `api_features`. Each feature owns its routes of
administration, and needs the functions of `api.Server.Functions` they call (a plugin provides them): the start fails,
naming the undefined functions, if an enabled feature misses some. No route is loaded without feature.

```yaml
api_features: [status, locales, timezones, auth, account, contact]
```

| Feature | Routes |
| --- | --- |
| `status`, `locales`, `timezones` | `GET /status`, the supported locales, the timezones of the countries. They need nothing from the application. |
| `terms` | The terms of use, their version, their acceptance. |
| `notifications` | The notifications of the users: list, center, settings, unsubscription, and their administration. |
| `registration` | The registration with a password, and the availability of the usernames. |
| `forgot_password` | The reset of a forgotten password. |
| `auth` | The login with a password, the auth tokens of the users, and their administration. |
| `facebook`, `google` | The registration and the login with the provider. |
| `account` | The account of the users, and the administration of the users: list, update, disable, delete, abilities. |
| `contact` | The contact messages, and their administration. |

`api.Features()` returns the names. A route can still be enabled or moved in `RouterConfigureDefaultRoutes`.

The features other than `status`, `locales` and `timezones` sign and verify the auth tokens, so they need `jwt.secret`:
the start fails without it. Without a secret, the middlewares which authenticate (`MustBeAuthenticated`, `rights`,
etc.) of the custom routes answer an error: set `jwt.secret` as soon as a route authenticates.

### Secrets

The configuration file should hold no secret (`jwt.secret`, `database.password`, `kafka.sasl.password`, `smtp.password`,
API keys): it is committed and copied into the images. Any value can be read from outside of the file with a tag:

```yaml
jwt:
  secret: !file /run/secrets/jwt_secret   # the content of the file, without its trailing line break
database:
  password: !env DB_PASSWORD              # the environment variable
```

Prefer `!file` with the secrets of Docker or Kubernetes, mounted in memory: the environment variables show up in
`docker inspect` and in the dumps of the process. The start fails, naming the key and the file or the variable but never
the value, if the file is unreadable or the variable is not set. In production it also fails if `jwt.secret` still
starts with `change-me`, the placeholder of the examples.

## Layout

| Directory | Content |
| --- | --- |
| `.` | `Server`, `Plugin`, `Connector` and `Run`, which sets up the configuration and runs the chosen server. |
| `api/` | The API server: routes, middlewares, request validation, responses, authentication cache. |
| `worker/` | The worker server: message consumers, notifications, mails, async and recurring tasks. |
| `web/` | The web server: HTML routes, pages, cookies, calls to the API, `robots.txt`. |
| `common/` | Shared packages: configuration, flags, logger, localization, templates, broker interfaces, JWT, timezones. |
| `plugins/` | The plugins: the database (`mongodb`) and the message broker (`kafka`). |
| `examples/` | Small projects built with ginamite, deployed with a Dockerfile. See [examples](examples/README.md). |
| `documentation/` | The detailed documentation: [message broker and Kafka](documentation/kafka.md). |

## Database plugins

The core does not depend on a database. A database plugin, such as `plugins/mongodb`, provides the storage through
the extension points below, and is registered with `server.AddPlugin`.

| Extension point | Role |
| --- | --- |
| `ginamite.Plugin.Configure` | Sets the functions of the API (`api.Server.Functions`) and of the worker (`worker.Server.Functions`). An error stops `Run`. |
| `ginamite.Connector` (optional) | `Connect` opens the connection before the plugins are configured, and returns an error if it fails; `Close` closes it when the server stops. Called only for the API and worker servers. |
| `api.Server.Middlewares` | Gin middlewares added before the routes, for instance to put the database in the context of the requests. |
| `api.Handlers.IsValidID` | Reports whether an identifier in a URL is well-formed. It depends on the database (an ObjectID for MongoDB, a UUID or an integer for SQL). By default every identifier is accepted. |

A plugin for another database implements the same extension points. `plugins/mongodb` is the reference.

### MongoDB

The packages `plugins/mongodb/mongo/session`, `store`, `document` and `bson` are the MongoDB helpers (session, driver,
generic documents, BSON builders): an application can use them for its own collections.

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

The code is formatted with `gofumpt`, `golangci-lint` must report no issue, and the code is tested along with it.

### Release

The release targets check the code, bump the `VERSION` file, commit it, tag it and push:

```bash
make bump_patch   # or bump_minor, bump_major
```

## License

[MIT](LICENSE)
