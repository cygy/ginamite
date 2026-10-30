package cookie

import (
	"net/http"
	"testing"

	"github.com/cygy/ginamite/web/context"
	"github.com/cygy/ginamite/web/internal/pagetest"
	"github.com/cygy/ginamite/web/internal/webtest"
)

// responseCookie returns the authentication cookie of the response, or nil if there is none.
func responseCookie(t *testing.T, header http.Header) *http.Cookie {
	t.Helper()

	for _, c := range (&http.Response{Header: header}).Cookies() {
		if c.Name == pagetest.CookieName {
			return c
		}
	}

	return nil
}

func TestCreate(t *testing.T) {
	pagetest.Setup(t)
	c, rec := webtest.Context(http.MethodGet, "/", nil)

	Create(c, "token")

	got := responseCookie(t, rec.Header())
	if got == nil {
		t.Fatal("no cookie sent")
	}
	if got.Value != "token" || got.MaxAge != 3600 || got.Path != "/" || got.Domain != "web.test" || !got.Secure || !got.HttpOnly {
		t.Errorf("cookie = %+v", got)
	}
	if rec.Header().Get("Location") != "" {
		t.Error("Create() must not redirect")
	}
}

func TestCreateAndRedirect(t *testing.T) {
	pagetest.Setup(t)
	c, rec := webtest.Context(http.MethodGet, "/", nil)
	context.SaveLocale("fr-fr")(c)

	CreateAndRedirect(c, "token", "index")

	if got := responseCookie(t, rec.Header()); got == nil || got.Value != "token" {
		t.Errorf("cookie = %+v", got)
	}
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != pagetest.WebHost+"/fr-fr/accueil" {
		t.Errorf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestInvalidate(t *testing.T) {
	pagetest.Setup(t)
	c, rec := webtest.Context(http.MethodGet, "/", nil)

	Invalidate(c)

	got := responseCookie(t, rec.Header())
	if got == nil || got.Value != "" || got.MaxAge >= 0 {
		t.Errorf("cookie = %+v, want an expired cookie", got)
	}
	if rec.Header().Get("Location") != "" {
		t.Error("Invalidate() must not redirect")
	}
}

func TestInvalidateAndRedirect(t *testing.T) {
	pagetest.Setup(t)
	c, rec := webtest.Context(http.MethodGet, "/", nil)
	context.SaveLocale("en-us")(c)

	InvalidateAndRedirect(c, "index")

	if got := responseCookie(t, rec.Header()); got == nil || got.MaxAge >= 0 {
		t.Errorf("cookie = %+v, want an expired cookie", got)
	}
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != pagetest.WebHost+"/en-us/home" {
		t.Errorf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGetValue(t *testing.T) {
	pagetest.Setup(t)

	c, _ := webtest.Context(http.MethodGet, "/", nil)
	if _, err := GetValue(c); err == nil {
		t.Error("GetValue() without cookie must return an error")
	}

	c.Request.AddCookie(&http.Cookie{Name: pagetest.CookieName, Value: "token"})
	if got, err := GetValue(c); err != nil || got != "token" {
		t.Errorf("GetValue() = %q, %v", got, err)
	}
}
