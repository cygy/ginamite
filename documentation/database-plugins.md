# Database plugins

The core does not depend on a database. A database plugin, such as `plugins/mongodb`, provides the storage through
the extension points below, and is registered with `server.AddPlugin`.

| Extension point | Role |
| --- | --- |
| `ginamite.Plugin.Configure` | Sets the storages and the hooks of the API (`api.Server.Contacts`, `Account`, `Hooks`, etc.) and the storages and the hooks of the worker (`worker.Server.Storage`, `Hooks`). An error stops `Run`. |
| `ginamite.Connector` (optional) | `Connect` opens the connection before the plugins are configured, and returns an error if it fails; `Close` closes it when the server stops. Called only for the API and worker servers, and for the API alone if the plugin also implements `ginamite.APIOnly`. |
| `api.Server.Middlewares` | Gin middlewares added before the routes, for instance to add a request ID or a CORS policy. A database plugin does not need them. |
| `api.Server.Hooks.IsValidID` | Reports whether an identifier in a URL is well-formed. It depends on the database (an ObjectID for MongoDB, a UUID or an integer for SQL). By default every identifier is accepted. |

A plugin for another database implements the same extension points. `plugins/mongodb` is the reference.

The interfaces a plugin implements, by feature. A feature whose storage is left `nil` is disabled, with a warning at
startup. Each method takes the `context.Context` of the request (or of the task, in the worker) and returns an error:
`errors.ErrNotFound` when the document does not exist.

| Server | Field | Interface |
| --- | --- | --- |
| API | `Contacts` | `contact.Storage` |
| API | `Terms` | `terms.Storage` |
| API | `Users` | `user.Storage` (administration of the users) |
| API | `Account` | `account.Storage`: `Accounts`, `Tokens`, `Notifications` and `Processes` (`AccountsStorage`, `TokensStorage`, `NotificationsStorage`, `ProcessesStorage`) |
| API | `Hooks` | `IsValidID`, `MiddlewareIsValidAuthToken`, `MiddlewareGetUserAndLocaleFromAuthToken`, `MiddlewareGetUserAbilities`, `BuildTokenFromID`, `ExtendTokenExpirationDateFromID`, etc. |
| Worker | `Storage` | `worker.Storage`: `Tokens`, `Accounts`, `Users`, `Notifications`, `Locations`, `Emails` and `Sitemap` |
| Worker | `Hooks` | `worker.Hooks`: the tasks, mails and notifications of the application |

The rate limiters of the requests are not stored by the database plugins: they keep their counters in memory by default,
and a rate limiting plugin of its own shares them between the instances, see [Rate limiting](rate-limiting.md).

## MongoDB

The packages `plugins/mongodb/mongo/session`, `store`, `document` and `bson` are the MongoDB helpers (session, driver,
generic documents, BSON builders): an application can use them for its own collections.

`plugins/mongodb` opens its own MongoDB session in `Connect`, with no global state: the functions of the API and of the
worker work on the database of the plugin, which they receive when they are created, not through the context of the
requests. Its settings (`Plugin.UserMaxImageSize`, etc., see `Plugin.Settings`) are given the same way. It stores the users, their auth tokens, their notifications and the contacts. Create it
with `mongodb.NewPlugin()`, set the properties that differ from the defaults (abilities, size of the images, etc.), then
register it. Its hooks let the application add its own collections, indexes and default documents, and update the
documents which reference the users.

At its start, the worker creates the indexes of the plugin and of the application. When the name of an index is already
used by an index with another definition (other keys, partial filter or options, left by an older version), MongoDB
refuses it (`IndexKeySpecsConflict` or `IndexOptionsConflict`): the worker then drops the old index and creates the new
one, and logs `mongoDB index replaced`. Only the indexes with a name set in their options can be replaced. If the new
definition cannot be created (for example a unique index on duplicated values), the error is logged and the collection
is left without this index.
