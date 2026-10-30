# ginamite

A framework to build an application made of an API server, a worker server and a web server.

## Database plugins

The core does not depend on a database. A database plugin, such as `plugins/mongodb`, provides the storage through
the extension points below, and is registered with `server.AddPlugin`.

| Extension point | Role |
| --- | --- |
| `ginamite.Plugin.Configure` | Sets the functions of the API (`api.Server.Functions`) and of the worker (`worker.Server.Functions`). |
| `ginamite.Connector` (optional) | `Connect` opens the connection before the plugins are configured, `Close` closes it when the server stops. Called only for the API and worker servers. |
| `api.Server.Middlewares` | Gin middlewares added before the routes, for instance to put the database in the context of the requests. |
| `api.Handlers.IsValidID` | Reports whether an identifier in a URL is well-formed. It depends on the database (an ObjectID for MongoDB, a UUID or an integer for SQL). By default every identifier is accepted. |

```go
server := ginamite.NewServer()
server.AddPlugin(mongodb.NewPlugin())
server.Start()
```

A plugin for another database implements the same extension points. `plugins/mongodb` is the reference.
