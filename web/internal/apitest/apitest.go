// Package apitest replaces the main API client with a client of a test server.
package apitest

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/cygy/ginamite/web/api"
)

// SetAddress replaces the main API client, for the duration of the test, with a client of the API at host:port.
func SetAddress(t *testing.T, host string, port int) {
	t.Helper()

	old := api.Main
	api.Main = api.NewClient(host, "v1", port, 2000, 0, false)
	t.Cleanup(func() { api.Main = old })
}

// Set replaces the main API client, for the duration of the test, with a client of a server running the handler.
func Set(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	host, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	SetAddress(t, "http://"+host, portNumber)
}
