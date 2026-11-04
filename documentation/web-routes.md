# Web routes

The file given by the argument `-web-routes` lists the routes of the website, by key and by locale, see
[Web server](web-server.md). `Run` loads it **once**, after the configuration (the URLs are prefixed with `hosts.web`),
and shares the same `*weburl.Configuration` with the three servers and the plugins.

| Server | Without `-web-routes` | With an unreadable file |
| --- | --- | --- |
| web | refuses to start (the argument is mandatory) | fails to start |
| API | the routes are `nil` | fails to start |
| worker | the routes are `nil`: the URLs of the emails are empty and the sitemaps are not generated | fails to start |

A `nil` `*weburl.Configuration` is safe to use: `GetURL` and `GetURLWithParams` log a warning and return an empty string.

## Where to read the routes

- **API and web handlers**: `weburl.Of(c)` (a middleware, installed by default, saves the routes in the Gin context).
- **Plugins**: `server.WebRoutes()` in `Configure`. The routes are loaded before the plugins are configured.
- **Worker**: the emails use them; the tasks and the hooks of the application read them through the server, e.g.
  `server.WebRoutes()` in a closure called once `Run` started, or `server.Worker.Routes`.
- **Web**: `weburl.Of(c)`, the same configuration as the other servers. The pages are still read from the file by the
  web server (`html.Load`).

## Building a URL

```go
routes := weburl.Of(c)

routes.GetURL("account.profile", "fr-fr")
// https://www.example.com/fr-fr/compte

routes.GetURLWithParams("order.show", "en-us", map[string]string{"id": "42", "slug": "my-order"})
// https://www.example.com/en-us/order/42/my-order   (route path: /{{.Locale}}/order/:id/:slug)
```

`GetURLWithParams` replaces each `:name` of the path by its value (the key of `params` may start with a colon). The names
are matched entirely (`id` does not alter `:identifier`), a parameter without value is left as is, and the values are
inserted as they are: escape them with `url.PathEscape` if they may contain a `/` or a `?`.

## Migration from v0.9

The argument `-routes` is **removed**, without alias: the servers refuse to start with it (`flag provided but not
defined: -routes`). Rename it `-web-routes` in the command lines, the Docker files, the compose files, the service
definitions, etc. The API and the worker may now receive it too; the web server still requires it.

In the code, `flags.Flags.RoutesFile` is now `WebRoutesFile`, `worker.Server.RoutesFile` is replaced by
`worker.Server.Routes` (a `*weburl.Configuration` set by `Run`, not a path), and the worker no longer reloads the file.
An application that loaded the file itself for the API (plugin or hook) can use `server.WebRoutes()` and `weburl.Of(c)`.
