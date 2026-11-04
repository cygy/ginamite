# Configuration

The server configuration is a YAML file, see `common/config.Configuration` for the complete list of settings and their
default values. The main sections are:

| Section | Content |
| --- | --- |
| `server` | Host, port, trusted proxies, maximum body size (`max_body_size`, 10 MiB by default), gzip, SSL and log format of the HTTP server. |
| `hosts` | Public hosts of the API, the web server and the static files. `hosts.web` is also the origin trusted by the web server for the cross-origin requests. |
| `database` | `service` (for instance `mongodb`), `host`, `port`, `database`, `username`, `password`, `time_out` in seconds. |
| `api_features` | The default routes of the API, by feature (see [API features](api-features.md)). |
| `jwt`, `cookie` | Settings of the authentication tokens and of the cookie (`same_site`: `lax` by default, `strict` or `none`, which requires `secure`). |
| `locales`, `default_locale`, `localization_path`, `templates_path` | Languages, translations and templates. |
| `smtp`, `sender_*` | Sending of the emails. |
| `facebook`, `google` | Sign-in with a social network. |
| `recurring_tasks`, `run_start_up_tasks` | Tasks executed by the workers, see [Worker](worker.md). |
| `metrics` | Prometheus endpoint, see [Observability](observability.md). |
| `api_rate_limiter` | `limits` of the requests (`authentication`, `registration`, `forgot_password`, each with `requests` and `seconds`), and the subsections `mongodb` and `redis` read by the rate limiting plugins, see [Rate limiting](rate-limiting.md). `rate_limits` is removed. |
| `broker` | The message broker: `service` (for instance `kafka`, the plugin which connects), `host` and `port`. |
| `kafka` | Read by `plugins/kafka`: connection to Kafka (version, TLS, SASL), prefix of the topics, retries, creation of the topics, statistics. |
| `stripe` | Read by `plugins/stripe`: the keys of Stripe, the currency, the payment methods, the routes of the webhook, of the intents and of the checkouts, see [Payments with Stripe](stripe.md). |
| `worker_consumers` | Kafka topics consumed by a worker, with their consumer group. |

The sections `broker`, `kafka` and `worker_consumers`, and the way the workers consume the messages, are described in
[Message broker and Kafka](kafka.md).

The sections the core does not know belong to the plugins: a plugin reads its own with
`Configuration.Section("name", &settings)`, which keeps the defaults of `settings` for the missing values and resolves
`!file` and `!env` like the other sections. A plugin which is a `ginamite.Commander` also adds its own arguments to the
command line with `RegisterFlags`.
