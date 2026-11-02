package metrics

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeCollector struct{}

func (fakeCollector) Collect(w *Writer) {
	w.Counter("jobs_total", "Jobs done.", 3, "kind", `a"b\c`+"\n")
	w.Counter("jobs_total", "Jobs done.", 4.5, "kind", "other")
	w.Gauge("queue_size", "Items waiting.", 2)
	w.Gauge("odd", "Ignored name without value.", 1, "alone")
}

func scrape(t *testing.T, r *Registry) string {
	t.Helper()

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain; version=0.0.4") {
		t.Errorf("Content-Type = %q", got)
	}

	return recorder.Body.String()
}

func TestRegistry_ServeHTTP(t *testing.T) {
	r := NewRegistry()
	r.Register(nil)
	r.Register(fakeCollector{})

	want := `# HELP jobs_total Jobs done.
# TYPE jobs_total counter
jobs_total{kind="a\"b\\c\n"} 3
jobs_total{kind="other"} 4.5
# HELP queue_size Items waiting.
# TYPE queue_size gauge
queue_size 2
# HELP odd Ignored name without value.
# TYPE odd gauge
odd 1
`
	if got := scrape(t, r); got != want {
		t.Errorf("metrics =\n%s\nwant\n%s", got, want)
	}
}

func TestRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requests := NewRequests()
	engine := gin.New()
	engine.Use(requests.Middleware())
	engine.GET("/users/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.POST("/users/:id", func(c *gin.Context) { c.Status(http.StatusConflict) })

	for _, call := range []struct{ method, path string }{
		{http.MethodGet, "/users/1"},
		{http.MethodGet, "/users/2"},
		{http.MethodPost, "/users/1"},
		{http.MethodGet, "/nowhere"},
	} {
		engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(call.method, call.path, nil))
	}
	requests.Observe(http.MethodGet, "/users/:id", http.StatusOK, 1500*time.Millisecond)

	r := NewRegistry()
	r.Register(requests)
	got := scrape(t, r)

	for _, want := range []string{
		`http_requests_total{method="GET",route="/users/:id",status="200"} 3`,
		`http_requests_total{method="POST",route="/users/:id",status="409"} 1`,
		`http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		"# TYPE http_request_duration_seconds_total counter",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("metrics lack %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "# TYPE http_requests_total") != 1 {
		t.Errorf("the type must be written once:\n%s", got)
	}
}

func TestRegistry_Serve(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeCollector{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := r.Serve(ctx, "invalid address"); err == nil {
		t.Error("an invalid address must fail")
	}

	// Find a free port.
	probe := httptest.NewServer(http.NotFoundHandler())
	address := strings.TrimPrefix(probe.URL, "http://")
	probe.Close()

	done := make(chan error, 1)
	go func() { done <- r.Serve(ctx, address) }()

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
	if !strings.Contains(body, "queue_size 2") {
		t.Errorf("body = %q", body)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve() = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not stop")
	}
}
