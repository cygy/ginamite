package web

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/web/html"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/config/environment"
	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/api"
	"github.com/cygy/ginamite/web/internal/webtest"
	"github.com/cygy/ginamite/web/route"

	"github.com/gin-gonic/gin"
)

const testRoutes = `
default:
  template:
    decorator: decorator.html
    main: main.html
routes:
  index:
    paths:
      en-us: "/{{.Locale}}/home"
  error.not_found:
    paths:
      en-us: "/{{.Locale}}/error/not-found"
  custom:
    paths:
      en-us: "/{{.Locale}}/custom"
`

// testConfiguration returns a configuration of a web server whose files are in a temporary directory.
func testConfiguration(t *testing.T, environmentName string) *config.Configuration {
	t.Helper()

	return &config.Configuration{
		Environment:      environmentName,
		Version:          "final",
		DefaultLocale:    "en-us",
		DefaultTimezone:  "UTC",
		SupportedLocales: []string{"en-us"},
		FallbackLocales:  map[string]string{"en": "en-us"},
		APIVersion:       "v1",
		TemplatesPath:    t.TempDir(),
		Hosts:            config.HostConfiguration{API: "https://api.test", Web: "https://web.test", Static: "https://static.test"},
		Cookie:           config.CookieConfiguration{Name: "auth"},
		JWT:              config.JWTConfiguration{Secret: "secret"},
		API:              config.APIConfiguration{Host: "http://127.0.0.1", Port: 1, Version: "v1", TimeOut: 200},
	}
}

// newTestServer returns a server whose pages are described by the test routes, and its configuration.
func newTestServer(t *testing.T, environmentName string) (*Server, *config.Configuration) {
	t.Helper()

	conf := testConfiguration(t, environmentName)
	webtest.SetConfig(t, conf)
	webtest.SetTranslations(t, map[string]map[string]string{"en-us": {"a": "b"}})
	webtest.WriteFile(t, conf.TemplatesPath, "pages/decorator.html", `<html>{{.IncludeMainTemplate}}</html>`)
	webtest.WriteFile(t, conf.TemplatesPath, "pages/main.html", `{{.Content.Title}}`)

	server := NewServer()
	server.Config = conf
	server.Robots = "User-agent: *"
	server.RoutesFile = webtest.WriteFile(t, t.TempDir(), "routes.yaml", testRoutes)

	return server, conf
}

func TestNewServer(t *testing.T) {
	server := NewServer()

	if server.customRoutes == nil || len(server.customRoutes) != 0 {
		t.Errorf("customRoutes = %v, want empty", server.customRoutes)
	}
	if consumers := server.CreateConsumers(); consumers == nil || len(consumers) != 0 {
		t.Errorf("CreateConsumers() = %v, want none", consumers)
	}
}

func TestServer_AddCustomRoutes(t *testing.T) {
	server := NewServer()

	server.AddCustomRoutes(func(*gin.RouterGroup, string, route.Middlewares, func(string) (string, bool, gin.HandlerFunc, bool)) {
	})
	server.AddCustomRoutes(func(*gin.RouterGroup, string, route.Middlewares, func(string) (string, bool, gin.HandlerFunc, bool)) {
	})

	if len(server.customRoutes) != 2 {
		t.Errorf("%d custom routes, want 2", len(server.customRoutes))
	}
}

func TestServer_SetUp(t *testing.T) {
	server, _ := newTestServer(t, environment.Test)

	var apiHandled bool
	server.Functions.APIMiddlewares = []api.ResponseHandlerFunc{
		func(*gin.Context, int, error, response.Error) bool { apiHandled = true; return false },
	}
	server.AddCustomRoutes(func(g *gin.RouterGroup, locale string, _ route.Middlewares, getRelativePath func(string) (string, bool, gin.HandlerFunc, bool)) {
		path, _, savePageKey, _ := getRelativePath("custom")
		g.GET(path, savePageKey, func(c *gin.Context) { c.String(http.StatusOK, "custom "+locale) })
	})

	engine := gin.New()
	if err := server.setUp(engine); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		target string
		status int
		body   string
	}{
		{"/robots.txt", http.StatusOK, "User-agent: *"},
		{"/en-us/custom", http.StatusOK, "custom en-us"},
		{"/en-us/error/not-found", http.StatusNotFound, "<html>"},
		{"/nowhere", http.StatusNotFound, "<html>"},
		{"/", http.StatusPermanentRedirect, ""},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.target, nil))

		if rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.body) {
			t.Errorf("%s: status = %d, body = %q, want %d containing %q", tt.target, rec.Code, rec.Body.String(), tt.status, tt.body)
		}
	}

	t.Run("API client", func(t *testing.T) {
		c, _ := webtest.Context(http.MethodGet, "/", nil)
		html.Set(c, server.pages)

		// Nothing listens on the port of the API: the built-in handler of the server errors renders the error page, and
		// the handlers of the application come after it.
		result, ok := server.apiClient.Call(c, api.NewRequest())

		if ok || result.RequestError == nil || !result.Handled || apiHandled {
			t.Errorf("result = %+v, ok = %v, application handler called = %v", result, ok, apiHandled)
		}
	})
}

func TestServer_SetUp_AppliesTheConfigurationOfTheDefaultRoutes(t *testing.T) {
	server, _ := newTestServer(t, environment.Test)
	server.Functions.RouterConfigureDefaultRoutes = func(routes *route.DefaultRoutes) {
		routes.Robots.Enabled = false
	}

	engine := gin.New()
	if err := server.setUp(engine); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if rec.Code != http.StatusNotFound || strings.Contains(rec.Body.String(), "User-agent") {
		t.Errorf("status = %d, body = %q, want the robots route to be disabled", rec.Code, rec.Body.String())
	}
}

func TestNewEngine(t *testing.T) {
	tests := []struct {
		name           string
		environment    string
		gzip           bool
		wantMode       string
		wantRequestLog bool
	}{
		{"development", environment.Development, false, gin.DebugMode, false},
		{"production", environment.Production, false, gin.ReleaseMode, true},
		{"test with gzip", environment.Test, true, gin.ReleaseMode, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { gin.SetMode(gin.TestMode) })
			conf := testConfiguration(t, tt.environment)
			conf.Server.GZip.Enabled = tt.gzip
			conf.Server.GZip.Level = gzip.DefaultCompression
			webtest.Hook.Reset()

			engine := (&Server{Config: conf}).newEngine()
			engine.GET("/", func(c *gin.Context) { c.String(http.StatusOK, strings.Repeat("hello ", 100)) })

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if gin.Mode() != tt.wantMode {
				t.Errorf("mode = %q, want %q", gin.Mode(), tt.wantMode)
			}
			if got := rec.Header().Get("Content-Encoding") == "gzip"; got != tt.gzip {
				t.Errorf("gzipped = %v, want %v", got, tt.gzip)
			}
			if tt.gzip {
				reader, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatal(err)
				}
				if body, _ := io.ReadAll(reader); !strings.HasPrefix(string(body), "hello hello") {
					t.Errorf("body = %q", body)
				}
			}

			logged := false
			for _, entry := range webtest.Hook.AllEntries() {
				logged = logged || entry.Message == "request"
			}
			if logged != tt.wantRequestLog {
				t.Errorf("request logged = %v, want %v", logged, tt.wantRequestLog)
			}
		})
	}
}

func TestServer_Run(t *testing.T) {
	type call struct {
		tls               bool
		address           string
		certPath, keyPath string
		handler           http.Handler
	}

	tests := []struct {
		name    string
		ssl     bool
		err     error
		wantTLS bool
	}{
		{"http", false, nil, false},
		{"https", true, nil, true},
		{"server error", false, errors.New("address already in use"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, conf := newTestServer(t, environment.Test)
			conf.Server.Port = 8443
			conf.Server.SSL = config.SSLConfiguration{Enabled: tt.ssl, CertPath: "cert.pem", KeyPath: "key.pem"}
			var got *call
			oldServe, oldServeTLS := listenAndServe, listenAndServeTLS
			t.Cleanup(func() { listenAndServe, listenAndServeTLS = oldServe, oldServeTLS })
			listenAndServe = func(address string, handler http.Handler) error {
				got = &call{false, address, "", "", handler}
				return tt.err
			}
			listenAndServeTLS = func(address, certPath, keyPath string, handler http.Handler) error {
				got = &call{true, address, certPath, keyPath, handler}
				return tt.err
			}

			err := server.Run()

			if !errors.Is(err, tt.err) || (err == nil) != (tt.err == nil) {
				t.Errorf("Run() = %v, want %v", err, tt.err)
			}
			if got == nil {
				t.Fatal("the server did not listen")
			}
			if got.tls != tt.wantTLS || got.address != ":8443" || (tt.wantTLS && (got.certPath != "cert.pem" || got.keyPath != "key.pem")) {
				t.Errorf("call = %+v", *got)
			}

			rec := httptest.NewRecorder()
			got.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("the routes are not loaded: status = %d", rec.Code)
			}
		})
	}
}

func TestServer_Run_SetUpErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Server, *config.Configuration)
		want  string
	}{
		{"missing templates", func(_ *Server, conf *config.Configuration) { conf.TemplatesPath = t.TempDir() }, "templates"},
		{"missing routes", func(server *Server, _ *config.Configuration) { server.RoutesFile = "missing.yaml" }, "routes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, conf := newTestServer(t, environment.Test)
			tt.setup(server, conf)

			if err := server.Run(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Run() = %v, want an error about %q", err, tt.want)
			}
		})
	}
}
