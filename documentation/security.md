# Security

- **Passwords** are hashed with Argon2id (PHC format, random salt, constant-time comparison). The legacy SHA-512 crypt
  hashes (`$6$`) are still accepted, up to 1024 bytes of password, and replaced by an Argon2id hash after a successful
  login. Logging in with an
  unknown email costs as much time as with a wrong password.
- **JWT**: only HS256, HS384 and HS512 are accepted, and the `exp` claim is required. The tokens are never logged.
  A `jwt` secret shorter than 32 characters logs a warning.
- **Rate limiting**: the sign-in, registration and forgot-password routes are limited per client IP. A refused request
  gets a `429` and a `Retry-After` header. By default the counters are a token bucket in the memory of each instance,
  so the limit applies per API instance. The MongoDB plugin replaces them with counters in the database
  (`rateLimits`, fixed windows), shared by the instances; another plugin can set `api.Server.Hooks.NewRateLimiter`
  to its own `request.RateLimiter`. A failure of the limiter lets the request through.

  ```yaml
  rate_limits:
    authentication:  { requests: 10, seconds: 60 }
    registration:    { requests: 5,  seconds: 60 }
    forgot_password: { requests: 5,  seconds: 60 }
  server:
    trusted_proxies: ["10.0.0.0/8"]  # the proxies allowed to give the client IP
  ```

  No proxy is trusted by default: the client IP, hence the limit and the IP saved with the auth tokens, is the one of
  the connection. Set `server.trusted_proxies` behind a reverse proxy, otherwise every client has the IP of the proxy.
  Only the trusted proxies can give the client IP, in `X-User-IP`, `X-Forwarded-For` or `X-Real-IP`: list the web
  servers in the setting of the API, since they send the IP of the user in `X-User-IP`.

- **Request bodies** are limited to `server.max_body_size` bytes, 10 MiB by default (`0` disables the limit), on the
  API and the web server: a larger body gets a `413`, or a `412` for an uploaded file. Set it above the largest file
  the application accepts.
- **Images**: an uploaded image declaring more than 50 megapixels (`image.MaxPixels`) is refused before it is decoded,
  so that a small file can not exhaust the memory.
- **CSRF**: the web server rejects with a `403` the cross-origin requests whose method is not safe (POST, PUT, PATCH,
  DELETE), using the headers `Sec-Fetch-Site` and `Origin`; `hosts.web` is trusted. The authentication cookie is
  `SameSite=Lax` by default, see `cookie.same_site`.
- **Security headers**: every response of the web server has `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`
  and `Content-Security-Policy: frame-ancestors 'none'` (no page in a frame of another site), and
  `Referrer-Policy: strict-origin-when-cross-origin`; `Strict-Transport-Security` (one year) is added in production. A
  handler can replace a header with `c.Header`. Add a full `Content-Security-Policy` in the reverse proxy or with a
  middleware of the application, since it depends on the scripts and styles of the pages.
- **Logs**: the request logger masks the `Authorization`, `Proxy-Authorization` and `Cookie` headers, the JSON values
  and the query parameters whose name contains `password`, `token` or `secret` (such as the `auth_token` of a login
  response), and the local part of the email addresses. The multipart bodies are not logged.

## Secrets

The configuration file should hold no secret (`jwt.secret`, `database.password`, `kafka.sasl.password`, `smtp.password`,
`facebook.app_secret`, API keys): it is committed and copied into the images. Any value can be read from outside of the file with a tag:

```yaml
jwt:
  secret: !file /run/secrets/jwt_secret   # the content of the file, without its trailing line break
database:
  password: !env DB_PASSWORD              # the environment variable
```

Prefer `!file` with the secrets of Docker or Kubernetes, mounted in memory: the environment variables show up in
`docker inspect` and in the dumps of the process. The start fails, naming the key and the file or the variable but never
the value, if the file is unreadable or the variable is not set. In production it also fails if `jwt.secret` still
starts with `change-me`, the placeholder of the examples.
