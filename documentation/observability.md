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
