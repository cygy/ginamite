# Rate limiting

The sign-in, registration and forgot-password routes of the API are limited per client IP: a refused request gets a `429`
and a `Retry-After` header. The limits are the same whatever stores the counters:

```yaml
api_rate_limiter:
  limits:
    authentication:  { requests: 10, seconds: 60 }   # the defaults
    registration:    { requests: 5,  seconds: 60 }
    forgot_password: { requests: 5,  seconds: 60 }
```

A group whose `requests` or `seconds` is `0` is not limited. The IP is the one of the connection, or the one given by a
trusted proxy, see [Security](security.md).

## Where the counters are

| Storage | How | Behavior |
| --- | --- | --- |
| `memory` | No rate limiting plugin (the default). | A token bucket in the memory of each instance: with several instances of the API a client can send the limit times the number of instances, and the counters are lost when the API restarts. |
| `redis` | `plugins/ratelimit/redis` | Counters shared by the instances, in fixed windows. |
| `mongodb` | `plugins/ratelimit/mongodb` | Counters shared by the instances, in fixed windows. |

The storage is **not** a database plugin: the rate limiting plugins are independent of `plugins/mongodb`, have their own
connection and settings, and the application can use any of them with any database plugin. Add the plugin to the server
before `Run`, and give the settings of its subsection in `api_rate_limiter`:

```go
server := ginamite.NewServer()
server.AddPlugin(redis.NewPlugin()) // github.com/cygy/ginamite/plugins/ratelimit/redis
```

```yaml
api_rate_limiter:
  limits: { … }
  redis:
    host: redis
    port: 6379          # the default
    database: 0
    username: ""
    password: !env REDIS_PASSWORD   # or !file /run/secrets/redis_password
    prefix: "ginamite:ratelimit:"   # the default
    timeout: 2          # seconds
```

```go
server.AddPlugin(mongodb.NewPlugin()) // github.com/cygy/ginamite/plugins/ratelimit/mongodb
```

```yaml
api_rate_limiter:
  mongodb:
    host: mongo
    port: 27017         # the default
    database: ratelimits
    username: ""
    password: !env RATE_LIMITER_MONGO_PASSWORD
    timeout: 10         # seconds
    collection: rateLimits   # the default
```

## Rules

- The plugins only work with the API (`ginamite.APIOnly`): the worker and the web server do not connect them.
- Only one rate limiting plugin can be loaded. With two, the API fails to start and the error names both.
- The plugins fail to start if the storage is not reachable (Redis is pinged, MongoDB too). Afterwards, if the storage
  fails, the request is let through and the error is returned to the limiter, never to the client.
- The API logs the active storage when it starts: `rate limiter store=memory|redis|mongodb`.
- Fixed windows allow a client to send up to twice the limit around the end of a window.
- The MongoDB plugin creates its collection and its index (`expireAt`, TTL 0: the counters are removed when their window
  ends) when the API starts; the operation is idempotent. The Redis plugin sets the expiration of a counter with its
  first request (Lua script, `INCR` then `PEXPIRE`, atomic).

## Another storage

Implement `request.RateLimiter` (`Allow(ctx, client) (ok, retryAfter, err)`) and register the limiters from `Configure` with
`ratelimit.Register(server, "name", newLimiter)` (`plugins/ratelimit`), which also refuses a second plugin. Implement
`ginamite.APIOnly` on a `Connector` to connect only for the API. Read the settings of your subsection with
`Configuration.RateLimiterStore("name", &settings)`.
