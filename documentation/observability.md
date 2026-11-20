# Metrics

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

## Logs

The logs use `log/slog`: `log.Main` is a `*slog.Logger` (JSON in production, text otherwise) with the name of the server
on each record. `log.Info("message", "key", value)`, `log.Warn`, `log.Error`, `log.Debug` and `log.With` log with it.
In a test, `logtest.Replace(&log.Main, slog.LevelInfo)` collects the records of the code under test.

In production, each request is logged with its headers and body, and the API logs the response body too. The secrets
are masked, see [Security](security.md).

### Application version

Set `Server.ApplicationVersion` to the release of the application, such as `"1.2.3"`, usually at the build with
`-ldflags "-X main.version=1.2.3"`: each record then carries `version`, so that a log tells which release wrote it. It is not the argument `-version` (`final`, `beta` or `alpha`), which
chooses the kind of instance.

### Database errors

The MongoDB plugin logs a failed operation as `database error`, with enough to find the code to fix:

| Attribute | Content |
| --- | --- |
| `operation`, `collection` | The driver operation (`Aggregate`, `FindOne`, `UpdateMany`...) and the collection. |
| `caller`, `source` | The function which called the code reporting the error, with its file and line: the function of the application for a call of the plugin, or the function of a plugin model for the ones it makes itself. |
| `request_id`, `method`, `route` | The request being served: the ID (the header `X-Request-ID`), the method and the route pattern. |
| `error` | The error of the driver. |

`log.InjectRequestLogger` saves the request attributes in the context of the request, `c.Request.Context()`: pass it
to the storage calls. A handler adds what identifies its work, such as the ID of a puzzle, with
`ctx = log.ContextWithAttributes(ctx, "puzzle_id", id)`; `log.DatabaseOperationError(ctx, err, operation, collection)`
logs an error the same way. Outside a request, such as in a recurring task, there is no request attribute.
