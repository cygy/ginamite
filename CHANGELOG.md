# Changelog

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
