# Configuration

The server configuration is a YAML file, see `common/config.Configuration` for the complete list of settings and their
default values. The main sections are:

| Section | Content |
| --- | --- |
| `server` | Host, port, gzip, SSL and log format of the HTTP server. |
| `hosts` | Public hosts of the API, the web server and the static files. |
| `database` | `service` (for instance `mongodb`), `host`, `port`, `database`, `username`, `password`, `time_out` in seconds. |
| `api_features` | The default routes of the API, by feature (see [API features](api-features.md)). |
| `jwt`, `cookie` | Settings of the authentication tokens and of the cookie. |
| `locales`, `default_locale`, `localization_path`, `templates_path` | Languages, translations and templates. |
| `smtp`, `sender_*` | Sending of the emails. |
| `facebook`, `google` | Sign-in with a social network. |
| `recurring_tasks`, `run_start_up_tasks` | Tasks executed by the workers, see [Worker](worker.md). |
| `metrics`, `rate_limits` | Prometheus endpoint and limits of the requests (see [Security](security.md) and [Observability](observability.md)). |
| `broker` | The message broker: `service` (for instance `kafka`, the plugin which connects), `host` and `port`. |
| `kafka` | Read by `plugins/kafka`: connection to Kafka (version, TLS, SASL), prefix of the topics, retries, creation of the topics, statistics. |
| `worker_consumers` | Kafka topics consumed by a worker, with their consumer group. |

The sections `broker`, `kafka` and `worker_consumers`, and the way the workers consume the messages, are described in
[Message broker and Kafka](kafka.md).

The sections the core does not know belong to the plugins: a plugin reads its own with
`Configuration.Section("name", &settings)`, which keeps the defaults of `settings` for the missing values and resolves
`!file` and `!env` like the other sections. A plugin which is a `ginamite.Commander` also adds its own arguments to the
command line with `RegisterFlags`.
