package html

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/response"
	"github.com/cygy/ginamite/web/internal/webtest"

	"github.com/gin-gonic/gin"
)

func TestPage_Render(t *testing.T) {
	engine := setup(t)

	renderers := map[string]struct {
		render func(page *Page, c *gin.Context)
		status int
	}{
		"/ok":           {(*Page).RenderOk, http.StatusOK},
		"/error":        {(*Page).RenderError, http.StatusInternalServerError},
		"/unauthorized": {(*Page).RenderUnauthorized, http.StatusUnauthorized},
		"/notfound":     {(*Page).RenderNotFound, http.StatusNotFound},
	}

	for path, tt := range renderers {
		t.Run(path, func(t *testing.T) {
			var aborted bool
			engine.GET(path, func(c *gin.Context) {
				c.Set("_locale", "fr-fr")
				page := testConf.PageByKey(c, "home")
				page.SetVar("args", map[string]string{"Name": "Marie"})
				tt.render(page, c)
				aborted = c.IsAborted()
			})

			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if want := "<html><h1>Accueil</h1>Bonjour Marie</html>"; rec.Body.String() != want {
				t.Errorf("body = %q, want %q", rec.Body.String(), want)
			}
			if !aborted {
				t.Error("the following handlers must be aborted")
			}
		})
	}
}

func TestPage_Process(t *testing.T) {
	setup(t)

	page := testConf.PageByKey(requestFor("en-us"), "home")
	page.Content.Title = "Hello {{.Vars.name}}"
	page.Meta.Description = "{{.Vars.name}} page"
	page.Meta.OpenGraph.Title = ""
	page.Meta.OpenGraph.Description = "OG {{.Vars.name}}"
	page.Meta.HrefLang[0].Href = "{{.Vars.name}}/a"
	page.Meta.Keywords = []string{"{{.Vars.name}}"}
	page.SetVar("name", "Bob")
	page.SetVar("args", map[string]string{"Name": "Bob"})

	page.Process()

	tests := []struct{ name, got, want string }{
		{"title", page.Content.Title, "Hello Bob"},
		{"description", page.Meta.Description, "Bob page"},
		{"canonical", page.Meta.CanonicalURL, "Bob/a"},
		{"open graph URL", page.Meta.OpenGraph.URL, "Bob/a"},
		{"open graph title falls back on the title", page.Meta.OpenGraph.Title, page.Meta.Title},
		{"open graph description", page.Meta.OpenGraph.Description, "OG Bob"},
		{"keyword", page.Meta.Keywords[0], "Bob"},
		{"main template", string(page.IncludeMainTemplate()), "<h1>Hello Bob</h1>Hello Bob"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestPage_Process_OpenGraphOverrides(t *testing.T) {
	setup(t)

	page := testConf.PageByKey(requestFor("en-us"), "home")
	page.Meta.OpenGraph.Title = "OG {{.Vars.name}}"
	page.SetVar("name", "Bob")

	page.Process()

	if page.Meta.OpenGraph.Title != "OG Bob" {
		t.Errorf("open graph title = %q, want %q", page.Meta.OpenGraph.Title, "OG Bob")
	}
}

func TestPage_Process_TemplateError(t *testing.T) {
	setup(t)
	webtest.Hook.Reset()

	page := testConf.PageByKey(requestFor("en-us"), "home")
	page.Template.Main = "missing.html"

	page.Process()

	if got := page.IncludeMainTemplate(); got != "" {
		t.Errorf("IncludeMainTemplate() = %q, want empty", got)
	}
	entry := webtest.Hook.LastEntry()
	if entry == nil || entry.Message != "unable to execute the template" || entry.Data["template"] != "missing.html" || entry.Data["page"] != "home" {
		t.Errorf("logged entry = %+v", entry)
	}
	if _, leaked := entry.Data["vars"]; leaked {
		t.Error("the variables of the page must not be logged")
	}
}

func TestPage_SetError(t *testing.T) {
	page := NewPage()

	page.SetError(response.Error{Message: "message", Reason: "reason", Recovery: "recovery"})

	if want := (Error{Message: "message", Reason: "reason", Recovery: "recovery"}); page.Content.Error != want {
		t.Errorf("Error = %+v, want %+v", page.Content.Error, want)
	}
	if !page.Content.HasError() {
		t.Error("HasError() = false")
	}
}

func TestPage_SetInfo(t *testing.T) {
	page := NewPage()

	page.SetInfo(Info{Title: "title", Message: "message"})

	if want := (Info{Title: "title", Message: "message"}); page.Content.Info != want {
		t.Errorf("Info = %+v, want %+v", page.Content.Info, want)
	}
	if !strings.Contains(page.Content.Info.Message, "message") || !page.Content.HasInfo() {
		t.Error("HasInfo() = false")
	}
}
