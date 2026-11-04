# Changelog

## v0.10.0

### Breaking changes

- The command line argument `-routes` is **removed** and replaced by `-web-routes` (`flags.Flags.WebRoutesFile`), with
  no alias: a server started with `-routes` fails with `flag provided but not defined: -routes`. It is mandatory for
  the web server, and optional for the API and the worker.
- `worker.Server.RoutesFile` is replaced by `worker.Server.Routes`, the routes already loaded; the worker no longer
  reads the file itself.
- A web routes file that can not be read now makes the API and the worker fail to start, when given.

### Added

- The web routes file is loaded once by `Run`, after the configuration, and shared by the three servers and the
  plugins: `Server.WebRoutes()` (available in `Plugin.Configure`), `api.Server.WebRoutes`, `worker.Server.Routes` and
  `web.Server.Routes`. Without the argument, the routes are `nil`.
- A middleware, installed by default in the API and web servers, saves the routes in the Gin context; read them with
  `weburl.Of(c)`.
- `(*weburl.Configuration).GetURLWithParams(key, locale, params)` replaces the `:id`, `:slug`... of a path.
- `(*weburl.Configuration).GetURL` accepts a `nil` receiver: it logs a warning and returns an empty string.

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

See [Web routes](documentation/web-routes.md).
