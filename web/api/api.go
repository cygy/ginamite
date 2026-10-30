// Package api is the client of the internal API, used by the web server.
package api

var (
	// Main is the main API client of the application, set by Initialize.
	Main *Client

	// MainCache is the main cache for the results of some API calls.
	MainCache = NewCache()
)

// Status is the result of an API call returning only a status.
type Status struct {
	Status string `json:"status"`
}

// Initialize creates the main API client. The timeout is in milliseconds.
func Initialize(host, version string, port int, timeout, retryCount int, debug bool) {
	Main = NewClient(host, version, port, timeout, retryCount, debug)
}

// AddResponseHandlers adds some response handlers to the main API client.
func AddResponseHandlers(handlers ...ResponseHandlerFunc) {
	Main.AddResponseHandlers(handlers...)
}
