package ginamite

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/broker"
)

func TestServer_startMetrics(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		s := NewServer()
		s.startMetrics()()

		if s.API.Requests != nil {
			t.Error("the requests must not be counted when the metrics are disabled")
		}
	})

	t.Run("served", func(t *testing.T) {
		probe := httptest.NewServer(http.NotFoundHandler())
		address := strings.TrimPrefix(probe.URL, "http://")
		probe.Close()

		s := NewServer()
		s.Broker = broker.NewMemory() // not a collector
		s.Config.Metrics.Address = address
		stop := s.startMetrics()

		if s.API.Requests == nil {
			t.Fatal("the requests must be counted")
		}
		s.API.Requests.Observe(http.MethodGet, "/v1/ping", http.StatusOK, time.Millisecond)

		var body string
		for range 100 {
			response, err := http.Get("http://" + address + "/metrics") //nolint:noctx // a test against a local server
			if err != nil {
				time.Sleep(20 * time.Millisecond)
				continue
			}
			content, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			body = string(content)

			break
		}
		stop()

		if !strings.Contains(body, `http_requests_total{method="GET",route="/v1/ping",status="200"} 1`) {
			t.Errorf("body = %q", body)
		}
	})

	t.Run("address in use", func(t *testing.T) {
		probe := httptest.NewServer(http.NotFoundHandler())
		defer probe.Close()

		s := NewServer()
		s.Config.Metrics.Address = strings.TrimPrefix(probe.URL, "http://")
		s.startMetrics()() // logs the error, does not block
	})
}
