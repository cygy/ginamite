package api

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/web/internal/webtest"
)

// newTestClient returns a client of an API server running the handler, closed at the end of the test.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return clientOf(t, server.URL)
}

// clientOf returns a client of the API server at the URL "http://host:port".
func clientOf(t *testing.T, url string) *Client {
	t.Helper()

	host, port, err := net.SplitHostPort(strings.TrimPrefix(url, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	return NewClient("http://"+host, "v1", portNumber, 2000, 0, false)
}

func setCookieConfig(t *testing.T) {
	t.Helper()

	webtest.SetConfig(t, &config.Configuration{
		Cookie: config.CookieConfiguration{Name: "auth", Path: "/", TTL: 60},
	})
}
