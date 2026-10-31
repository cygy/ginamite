// Package api is the client of the internal API, used by the web server.
package api

// Status is the result of an API call returning only a status.
type Status struct {
	Status string `json:"status"`
}
