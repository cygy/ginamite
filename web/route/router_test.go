package route

import (
	"net/http"
	"strings"
	"testing"

	"github.com/cygy/ginamite/web/internal/pagetest"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

func TestNewRouter(t *testing.T) {
	router, _ := newRouter(t)

	robots := router.DefaultRoutes.Robots
	if !robots.Enabled || !robots.Paths.Get.Enabled || robots.Paths.Get.Path != "/robots.txt" {
		t.Errorf("robots = %+v, want enabled on /robots.txt", robots)
	}
	if router.Handlers.MustBeAuthenticated == nil {
		t.Error("MustBeAuthenticated is not defined")
	}
}

func TestRouter_LoadDefaultRoutes(t *testing.T) {
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	tests := []struct {
		name     string
		target   string
		status   int
		contains string
		location string
	}{
		{"robots", "/robots.txt", http.StatusOK, "User-agent: *", ""},
		{"index redirects to the default locale", "/", http.StatusPermanentRedirect, "", "/en-us/home"},
		{"not found page", "/en-us/error/not-found", http.StatusNotFound, "<html>", ""},
		{"not found page in french", "/fr-fr/erreur/introuvable", http.StatusNotFound, "<html>", ""},
		{"internal server error page", "/en-us/error/internal-server", http.StatusInternalServerError, "<html>", ""},
		{"unauthorized page", "/en-us/error/unauthorized", http.StatusUnauthorized, "<html>", ""},
		{"invalid authentication page", "/en-us/error/invalid-authentication", http.StatusUnauthorized, "<html>", ""},
		{"expired authentication page", "/en-us/error/expired-authentication", http.StatusUnauthorized, "<html>", ""},
		{"revoked authentication page", "/en-us/error/revoked-authentication", http.StatusUnauthorized, "<html>", ""},
		{"unknown route", "/nowhere", http.StatusNotFound, "<html>", ""},
		{"unknown route of a locale", "/fr-fr/nowhere", http.StatusNotFound, "<html>", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(engine, http.MethodGet, tt.target)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if !strings.Contains(rec.Body.String(), tt.contains) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.contains)
			}
			if got := rec.Header().Get("Location"); got != tt.location {
				t.Errorf("location = %q, want %q", got, tt.location)
			}
		})
	}
}

func TestRouter_LoadDefaultRoutes_UnknownRoutePageIsLocalized(t *testing.T) {
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	rec := do(engine, http.MethodGet, "/fr-fr/nowhere/")

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "<html>") {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}

func TestRouter_LoadDefaultRoutes_DisabledRobots(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(r *Router)
	}{
		{"robots disabled", func(r *Router) { r.DefaultRoutes.Robots.Enabled = false }},
		{"get disabled", func(r *Router) { r.DefaultRoutes.Robots.Paths.Get.Enabled = false }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, engine := newRouter(t)
			tt.mutate(router)
			router.LoadDefaultRoutes()

			if rec := do(engine, http.MethodGet, "/robots.txt"); rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}

func TestRouter_LoadDefaultRoutes_CustomRobotsPath(t *testing.T) {
	router, engine := newRouter(t)
	router.DefaultRoutes.Robots.Paths.Get.Path = "/custom-robots.txt"
	router.LoadDefaultRoutes()

	if rec := do(engine, http.MethodGet, "/custom-robots.txt"); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouter_LoadCustomRoutes(t *testing.T) {
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	type call struct {
		locale  string
		path    string
		partial bool
		ok      bool
	}
	var calls []call

	router.LoadCustomRoutes(func(g *gin.RouterGroup, locale string, handlers Middlewares, getRelativePath func(key string) (string, bool, gin.HandlerFunc, bool)) {
		path, partial, savePageKey, ok := getRelativePath("page")
		calls = append(calls, call{locale, path, partial, ok})
		g.GET(path, savePageKey, func(c *gin.Context) {
			c.String(http.StatusOK, "page "+c.Param("id"))
		})

		_, _, _, ok = getRelativePath("unknown")
		calls = append(calls, call{locale, "", false, ok})
	})

	want := []call{
		{"en-us", "/page/:id", false, true},
		{"en-us", "", false, false},
		{"fr-fr", "/page/:id", false, true},
		{"fr-fr", "", false, false},
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("call %d = %+v, want %+v", i, calls[i], want[i])
		}
	}

	if rec := do(engine, http.MethodGet, "/fr-fr/page/7"); rec.Code != http.StatusOK || rec.Body.String() != "page 7" {
		t.Errorf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
}

func TestRouter_LoadCustomRoutes_Nil(t *testing.T) {
	router, _ := newRouter(t)

	router.LoadCustomRoutes(nil)
}

func TestRouter_MustBeAuthenticated(t *testing.T) {
	router, engine := newRouter(t)
	router.LoadDefaultRoutes()

	router.LoadCustomRoutes(func(g *gin.RouterGroup, locale string, handlers Middlewares, _ func(key string) (string, bool, gin.HandlerFunc, bool)) {
		ok := func(c *gin.Context) { c.String(http.StatusOK, "private") }
		g.GET("/private", handlers.MustBeAuthenticated(false), ok)
		g.GET("/private-partial", handlers.MustBeAuthenticated(true), ok)
	})

	token := webtest.SignToken(t, pagetest.Secret, map[string]any{"sub": "user1", "id": "token1"})
	withToken := func(r *http.Request) { r.AddCookie(&http.Cookie{Name: pagetest.CookieName, Value: token}) }
	noID := webtest.SignToken(t, pagetest.Secret, map[string]any{"sub": "user1"})
	withoutID := func(r *http.Request) { r.AddCookie(&http.Cookie{Name: pagetest.CookieName, Value: noID}) }

	tests := []struct {
		name     string
		target   string
		mutate   []func(*http.Request)
		status   int
		contains string
	}{
		{"authenticated", "/en-us/private", []func(*http.Request){withToken}, http.StatusOK, "private"},
		{"authenticated partial", "/en-us/private-partial", []func(*http.Request){withToken}, http.StatusOK, "private"},
		{"token without ID", "/en-us/private", []func(*http.Request){withoutID}, http.StatusUnauthorized, "<html>"},
		{"anonymous", "/en-us/private", nil, http.StatusUnauthorized, "<html>"},
		{"anonymous partial", "/en-us/private-partial", nil, http.StatusUnauthorized, `"message":"Not allowed"`},
		{"anonymous partial in french", "/fr-fr/private-partial", nil, http.StatusUnauthorized, `"message":"Interdit"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(engine, http.MethodGet, tt.target, tt.mutate...)

			if rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.contains) {
				t.Errorf("status = %d, body = %q, want %d containing %q", rec.Code, rec.Body.String(), tt.status, tt.contains)
			}
			if tt.status == http.StatusUnauthorized && strings.Contains(rec.Body.String(), "private") && !strings.Contains(tt.contains, "private") {
				t.Error("the private content was sent")
			}
		})
	}
}
