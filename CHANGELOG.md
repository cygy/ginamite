# Changelog

## Unreleased

### Changed

- Dependencies updated: `go.mongodb.org/mongo-driver/v2` v2.9.3, `github.com/xdg-go/scram` v1.4.0 and the indirect
  modules (`gin-contrib/sse`, `validator/v10`, `klauspost/cpuid/v2`, `quic-go`, `shopspring/decimal`). The integration
  tests pass against MongoDB 9 and Kafka 4.4.0.
- `make integration` also runs the integration test of the MongoDB rate limiter (`plugins/ratelimit/mongodb`), which
  its documentation announced but the target did not run.

## v0.11.1

### Fixed

- The translations accept the unsigned integers (`uint`, `uint8` to `uint64`) as plural counts, like the signed ones.
  go-i18n does not accept them as they are: they are now converted to strings, as the floats already were.

## v0.11.0

### Added

- The package `common/yaml` is public: `yaml.Unmarshal(source, outs...)` decodes a YAML document after replacing the
  scalars tagged `!file` (content of a file, without its trailing line break) and `!env` (environment variable), and
  `yaml.Resolve(node, key)` does it on a `yaml.Node`. The applications use it to read their own configuration file
  with the same secrets as ginamite. The behaviour of the server configuration is unchanged.
- `plugins/stripe` adds the API routes creating the payments: `POST intent_path` (payment intent, for Stripe.js) and
  `POST checkout_path` (checkout session), behind the middlewares checking the authentication and the acceptance of the
  latest terms. The application says what is bought with `Plugin.OrderFunc` (`func(c *gin.Context) (stripe.Order, bool)`):
  the items and their prices, the customer, the metadata, and the keys of the web routes where Stripe sends the user back.
  The function answers the request itself when it refuses the order. The plugin builds the success and cancel URLs from
  the web routes of the server, in the locale of the request. The handlers are also exported
  (`Service.IntentHandler`, `Service.CheckoutHandler`).
- New settings of the section `stripe`: `intent_enabled`, `checkout_enabled` (both `false` by default), `intent_path`
  (default `/stripe/intent`) and `checkout_path` (default `/stripe/checkout`), which may hold path parameters such as
  `/store/packs/:id/purchase/intent`.
- `Plugin.WebVars()` gives the variables of the pages paying with Stripe (`StripePublishableKey`, `IntentEnabled`,
  `CheckoutEnabled`) from the settings, to set in the templates of the application.
- `ginamite.Server.Role()` (`RoleAPI`, `RoleWorker`, `RoleWeb`, `RoleNone`) tells a plugin which server runs, set by `Run`
  before the plugins are configured. The Stripe plugin uses it: the web server only reads the settings, so its section
  `stripe` needs no secret, store or `OrderFunc`.

### Migration

- Nothing changes for an application which only uses the webhook: both new routes are disabled until
  `intent_enabled` or `checkout_enabled` is set.
- An application which had its own routes creating payments with `Plugin.Service()` can keep them, or move to
  `Plugin.OrderFunc` and delete its handlers. In that case, set `intent_path` and `checkout_path` to its current
  paths so the clients do not change, check that the checks its route middlewares did (identifying the item, its
  availability) are in the `OrderFunc`, and move the settings it kept elsewhere to the section `stripe` (or keep
  `Plugin.SettingsFunc`).
- Enabling the intents or the checkouts without an `OrderFunc` makes the API fail at startup.
- The web server now runs the plugin with the section `stripe` of **its own** configuration file (before, an
  application using `SettingsFunc` could share one source): add `enabled`, `publishable_api_key` and the two
  `*_enabled` flags to it; the secrets are not needed there.

## v0.10.0

### Breaking changes

- The command line argument `-routes` is **removed** and replaced by `-web-routes` (`flags.Flags.WebRoutesFile`), with
  no alias: a server started with `-routes` fails with `flag provided but not defined: -routes`. It is mandatory for
  the web server, and optional for the API and the worker.
- `worker.Server.RoutesFile` is replaced by `worker.Server.Routes`, the routes already loaded; the worker no longer
  reads the file itself.
- A web routes file that can not be read now makes the API and the worker fail to start, when given.
- The rate limiter of the API is **no longer provided by the MongoDB plugin**: `plugins/mongodb` does not set
  `api.Server.Hooks.NewRateLimiter` any more, and the worker no longer creates the collection `rateLimits` and its index
  (`model.RateLimitCollection` and `apiFunctions.NewRateLimiters` are removed). An application which relied on it to share
  the counters now counts in memory, by instance, until it adds a rate limiting plugin.
- The configuration section `rate_limits` is **replaced by `api_rate_limiter.limits`** (same keys `authentication`,
  `registration`, `forgot_password`), with no alias: a file which still has `rate_limits` is refused at startup, so that
  the default limits are not applied silently. `config.Configuration.RateLimits` becomes
  `config.Configuration.APIRateLimiter.Limits`.
- `connectPlugins` skips the connectors which implement the new `ginamite.APIOnly` when the API does not run.

### Added

- The web routes file is loaded once by `Run`, after the configuration, and shared by the three servers and the
  plugins: `Server.WebRoutes()` (available in `Plugin.Configure`), `api.Server.WebRoutes`, `worker.Server.Routes` and
  `web.Server.Routes`. Without the argument, the routes are `nil`.
- A middleware, installed by default in the API and web servers, saves the routes in the Gin context; read them with
  `weburl.Of(c)`.
- `(*weburl.Configuration).GetURLWithParams(key, locale, params)` replaces the `:id`, `:slug`... of a path.
- `(*weburl.Configuration).GetURL` accepts a `nil` receiver: it logs a warning and returns an empty string.
- The rate limiting plugins `plugins/ratelimit/redis` (Redis, atomic Lua script `INCR` + `PEXPIRE`) and
  `plugins/ratelimit/mongodb` (the former limiter, with its own connection and its collection and expiration index
  created when the API starts). They are independent of the database plugins, only run for the API, and read the
  subsections `api_rate_limiter.redis` and `api_rate_limiter.mongodb` (passwords with `!env` or `!file`).
- `plugins/ratelimit.Register` sets the limiters and refuses a second rate limiting plugin: the API fails to start.
- `config.Configuration.RateLimiterStore` decodes the subsection of a rate limiting plugin.
- The API logs the storage of its rate limiter when it starts (`memory`, `redis` or `mongodb`), see
  `api.Server.Hooks.RateLimiterStore`.
- [Rate limiting](documentation/rate-limiting.md).

### Changed

- Without web routes, the worker does not schedule the generation of the sitemaps.

### Migration from v0.9

1. Rename `-routes` into `-web-routes` everywhere the servers are started (compose files, Dockerfiles, services,
   Makefiles, deployment tools).
2. Replace `worker.Server.RoutesFile` and `flags.Flags.RoutesFile` by `Routes` and `WebRoutesFile`, if the application
   uses them.
3. Drop the plugins or the code that loaded the routes file to give the URLs of the website to the API: use
   `server.WebRoutes()` in a plugin, `weburl.Of(c)` in a handler, and `GetURLWithParams` instead of replacing `:id` by
   hand. Pass `-web-routes` to the API, or the URLs are empty.
4. Rename `rate_limits` into `api_rate_limiter.limits` in the configuration files:

   ```yaml
   api_rate_limiter:
     limits:
       authentication: { requests: 10, seconds: 60 }
   ```

5. With a single instance of the API, nothing else: the counters are in memory.
6. With several instances, add a plugin before `Run` and its subsection: `server.AddPlugin(redis.NewPlugin())`
   (`github.com/cygy/ginamite/plugins/ratelimit/redis`) with `api_rate_limiter.redis`, or
   `mongodb.NewPlugin()` (`github.com/cygy/ginamite/plugins/ratelimit/mongodb`) with `api_rate_limiter.mongodb`.
7. The collection `rateLimits` of the database of the MongoDB plugin is no longer used: drop it
   (`db.rateLimits.drop()`). The MongoDB rate limiting plugin has its own settings, and can use another database.

See [Web routes](documentation/web-routes.md) and [Rate limiting](documentation/rate-limiting.md).
