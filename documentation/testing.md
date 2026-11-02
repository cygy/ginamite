# End-to-end tests

The package `api/e2e` tests the endpoints of a running API: each request is sent over HTTP, and the status code, the
status, the error and the renewed auth token of its response are compared with the expected ones.

```go
var api = e2e.FromEnv("v1") // http://$API_HOST:$API_PORT/v1

func TestLogin(t *testing.T) {
	req := api.NewRequest(t)
	req.Method = http.MethodPost
	req.Endpoint = "/auth/login"
	req.Body = e2e.H{"identifier": "john", "password": "wrong"}
	req.Unauthorized()
}
```

## Logs in tests

`logtest.Replace(&log.Main, slog.LevelInfo)` (`common/log/logtest`) replaces the main logger with one collecting its
records, and returns the hook to read them with `AllEntries` and `LastEntry`, and the function restoring the logger.
Every record carries `server=test`; the integers in `Data` are `int` and `uint`.
