package route

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cygy/ginamite/api/internal/apitest"

	"github.com/gin-gonic/gin"
)

func newTestRouter() (*gin.Engine, *Router) {
	engine := gin.New()
	router := NewRouter(engine, "v1", false, "secret", "terms-1", nil, nil, nil, nil)

	return engine, router
}

// routeSet returns the "METHOD path" of the routes registered in the engine.
func routeSet(engine *gin.Engine) map[string]bool {
	set := map[string]bool{}
	for _, r := range engine.Routes() {
		set[r.Method+" "+r.Path] = true
	}

	return set
}

// setEnabled sets every "Enabled" field of the struct to the value.
func setEnabled(v reflect.Value, enabled bool) {
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Field(i)
			if v.Type().Field(i).Name == "Enabled" && field.Kind() == reflect.Bool {
				field.SetBool(enabled)
				continue
			}
			setEnabled(field, enabled)
		}
	default:
	}
}

func TestNewRouter(t *testing.T) {
	engine, router := newTestRouter()

	if router.Group.BasePath() != "/v1" {
		t.Errorf("BasePath() = %q, want /v1", router.Group.BasePath())
	}
	if router.ProductionEnvironment {
		t.Error("ProductionEnvironment = true, want false")
	}

	handlers := reflect.ValueOf(router.Handlers)
	for i := range handlers.NumField() {
		if handlers.Field(i).IsNil() {
			t.Errorf("handler %s is nil", handlers.Type().Field(i).Name)
		}
	}

	if !router.DefaultRoutes.Status.Enabled || router.DefaultRoutes.Status.Paths.Get.Path != "/status" {
		t.Errorf("status route = %+v, want it enabled on /status", router.DefaultRoutes.Status)
	}
	if len(engine.Routes()) != 0 {
		t.Errorf("NewRouter() registered %d routes, want none before LoadDefaultRoutes", len(engine.Routes()))
	}
}

func TestRouter_Rights(t *testing.T) {
	_, router := newTestRouter()

	if router.Rights(true, nil) == nil {
		t.Error("Rights() = nil, want a handler")
	}
}

func TestRouter_LoadDefaultRoutes(t *testing.T) {
	engine, router := newTestRouter()
	router.LoadDefaultRoutes()

	routes := routeSet(engine)
	for _, want := range []string{
		"GET /v1/status",
		"GET /v1/locales",
		"GET /v1/timezones",
		"GET /v1/notifications",
		"GET /v1/terms",
		"POST /v1/registration",
		"POST /v1/forgotPassword",
		"POST /v1/auth/login",
		"DELETE /v1/auth/token/:id",
		"GET /v1/account/notifications",
		"GET /v1/users/enabled",
		"GET /v1/contact",
	} {
		if !routes[want] {
			t.Errorf("route %q is not registered", want)
		}
	}
	if len(routes) < 50 {
		t.Errorf("%d routes registered, want at least 50", len(routes))
	}
}

func TestRouter_LoadDefaultRoutes_Disabled(t *testing.T) {
	engine, router := newTestRouter()
	setEnabled(reflect.ValueOf(&router.DefaultRoutes).Elem(), false)
	router.LoadDefaultRoutes()

	// Only the debug route remains, it is not a default route.
	routes := routeSet(engine)
	if len(routes) != 1 || !routes["GET /v1/debug/expiredToken"] {
		t.Errorf("routes = %v, want only the debug route when every default route is disabled", routes)
	}
}

func TestRouter_LoadDefaultRoutes_Production(t *testing.T) {
	engine := gin.New()
	router := NewRouter(engine, "v1", true, "secret", "terms-1", nil, nil, nil, nil)
	router.LoadDefaultRoutes()

	if routeSet(engine)["GET /v1/debug/expiredToken"] {
		t.Error("the debug route is registered in production")
	}
}

func TestRouter_LoadDefaultRoutes_PathsDisabled(t *testing.T) {
	engine, router := newTestRouter()
	router.DefaultRoutes.Status.Paths.Get.Enabled = false
	router.LoadDefaultRoutes()

	if routeSet(engine)["GET /v1/status"] {
		t.Error("the status route is registered, want it disabled")
	}
}

func TestRouter_StatusRoute(t *testing.T) {
	engine, router := newTestRouter()
	router.LoadDefaultRoutes()

	req, _ := http.NewRequest(http.MethodGet, "/v1/status", nil)
	recorder := apitest.Serve(engine, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestLimited(t *testing.T) {
	handler := func(*gin.Context) {}
	limit := func(*gin.Context) {}

	if got := limited(nil, handler); len(got) != 1 {
		t.Errorf("without limit: %d handlers, want 1", len(got))
	}
	if got := limited(limit, handler, handler); len(got) != 3 {
		t.Errorf("with a limit: %d handlers, want 3", len(got))
	}
}

func TestRouter_LoadDefaultRoutes_Limits(t *testing.T) {
	engine, router := newTestRouter()
	calls := map[string]int{}
	limit := func(name string) gin.HandlerFunc {
		return func(c *gin.Context) {
			calls[name]++
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
	router.Limits = RateLimits{Authentication: limit("authentication"), Registration: limit("registration"), ForgotPassword: limit("forgot_password")}
	router.LoadDefaultRoutes()

	for _, tt := range []struct{ method, path, group string }{
		{http.MethodPost, "/v1/auth/login", "authentication"},
		{http.MethodPost, "/v1/registration", "registration"},
		{http.MethodPost, "/v1/forgotPassword", "forgot_password"},
	} {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))

		if recorder.Code != http.StatusTooManyRequests || calls[tt.group] != 1 {
			t.Errorf("%s %s: status = %d, limit calls = %v", tt.method, tt.path, recorder.Code, calls)
		}
	}
}
