package route

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/web/html"
	"github.com/cygy/ginamite/web/internal/apitest"
	"github.com/cygy/ginamite/web/internal/pagetest"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

// upstream is what the API server received.
type upstream struct {
	method      string
	uri         string
	contentType string
	body        string
}

// setUpstream starts an API server which records the requests and answers with the status and the body.
func setUpstream(t *testing.T, status int, body string) *upstream {
	t.Helper()

	got := &upstream{}
	testClient = apitest.Set(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got = upstream{r.Method, r.URL.RequestURI(), r.Header.Get("Content-Type"), string(b)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})

	return got
}

func serveAsync(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func TestAsyncRoutes_Forward(t *testing.T) {
	got := setUpstream(t, http.StatusCreated, `{"liked":true}`)
	rec := serveAsync(t, http.MethodPost, "/fr-fr/async/aimer/42?x=1&y=a?b", `{"a":1}`)

	if rec.Code != http.StatusCreated || strings.TrimSpace(rec.Body.String()) != `{"liked":true}` {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	want := upstream{http.MethodPost, "/v1/likes/42?x=1&y=a?b", "application/json", `{"a":1}`}
	if *got != want {
		t.Errorf("upstream = %+v, want %+v", *got, want)
	}
}

func TestAsyncRoutes_EscapesTheParameters(t *testing.T) {
	got := setUpstream(t, http.StatusOK, `{}`)

	serveAsync(t, http.MethodPost, "/en-us/async/like/a%3Fb%23c%20d", "")

	if want := "/v1/likes/a%3Fb%23c%20d"; got.uri != want {
		t.Errorf("uri = %q, want %q", got.uri, want)
	}
}

func TestAsyncRoutes_APIError(t *testing.T) {
	setUpstream(t, http.StatusBadRequest, `{"error":{"code":12,"domain":"test","message":"nope"}}`)

	rec := serveAsync(t, http.MethodPost, "/en-us/async/like/1", "")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
	for _, want := range []string{`"error"`, `"code":12`, `"message":"nope"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Errorf("body = %q, want it to contain %q", rec.Body.String(), want)
		}
	}
}

func TestAsyncRoutes_RequestError(t *testing.T) {
	// Nothing listens on this port.
	testClient = apitest.NewClient("http://127.0.0.1", 1)
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	rec := do(engine, http.MethodPost, "/en-us/async/like/1")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAsyncRoutes_ReadError(t *testing.T) {
	setUpstream(t, http.StatusOK, `{}`)
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	req := httptest.NewRequest(http.MethodPost, "/en-us/async/like/1", failingReader{})
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestAsyncRoutes_UnsupportedMethod(t *testing.T) {
	router, engine := newRouter(t)
	webtest.Hook.Reset()
	pagetest.Pages.Async["bad"] = &html.Async{Paths: map[string]string{"en-us": pagetest.WebHost + "/en-us/bad"}, Method: "head", Target: "/bad"}

	router.LoadDefaultRoutes()

	if rec := do(engine, http.MethodHead, "/en-us/bad"); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if entry := webtest.Hook.LastEntry(); entry == nil || entry.Data["async"] != "bad" {
		t.Errorf("logged entry = %+v", entry)
	}
}

func TestAsyncRoutes_Methods(t *testing.T) {
	for _, method := range []string{"get", "PUT", "Patch", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			got := setUpstream(t, http.StatusOK, `{}`)
			router, engine := newRouter(t)
			pagetest.Pages.Async["verb"] = &html.Async{Paths: map[string]string{"en-us": pagetest.WebHost + "/en-us/verb"}, Method: method, Target: "/verb"}
			router.LoadDefaultRoutes()

			rec := do(engine, strings.ToUpper(method), "/en-us/verb")

			if rec.Code != http.StatusOK || got.method != strings.ToUpper(method) {
				t.Errorf("status = %d, upstream method = %q", rec.Code, got.method)
			}
		})
	}
}

func TestExpandEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		target string
		params gin.Params
		query  string
		want   string
	}{
		{"no parameter", "/list", nil, "", "/list"},
		{"query", "/list", nil, "a=1&b=2", "/list?a=1&b=2"},
		{"parameter", "/likes/:id", gin.Params{{Key: "id", Value: "42"}}, "", "/likes/42"},
		{"escaped parameter", "/likes/:id", gin.Params{{Key: "id", Value: "a/b c"}}, "", "/likes/a%2Fb%20c"},
		{"longest name first", "/x/:identifier/:id", gin.Params{{Key: "id", Value: "1"}, {Key: "identifier", Value: "2"}}, "", "/x/2/1"},
		{"repeated parameter", "/:id/:id", gin.Params{{Key: "id", Value: "1"}}, "", "/1/1"},
		{"unused parameter", "/likes", gin.Params{{Key: "id", Value: "42"}}, "q=1", "/likes?q=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandEndpoint(tt.target, tt.params, tt.query); got != tt.want {
				t.Errorf("expandEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}
