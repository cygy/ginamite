// Package apitest provides clients of the API connected to a test server.
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

// NewClient returns a client of the API at host:port.
func NewClient(host string, port int) *api.Client {
	return api.NewClient(host, "v1", port, 2000, 0, false)
}

// Set returns a client of a server running the handler, closed at the end of the test.
func Set(t *testing.T, handler http.HandlerFunc) *api.Client {
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

	return NewClient("http://"+host, portNumber)
}
