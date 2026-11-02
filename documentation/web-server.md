# Web server

The pages are described by the routes file (`-routes`), rendered with the Go templates of `templates_path`, and
translated with the files of `localization_path`. See the example [helloworld-web](../examples/helloworld-web).

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
