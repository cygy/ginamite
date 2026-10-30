package location

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

// useAPI sets the configuration of the API to the server, for the duration of the test.
func useAPI(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	old := config.Main
	t.Cleanup(func() { config.Main = old })
	config.Main = &config.Configuration{}
	config.Main.IPStack.Host = server.URL
	config.Main.IPStack.Key = "secret key"
}

func TestGetIPInfo(t *testing.T) {
	t.Run("saves the location", func(t *testing.T) {
		var gotIP, gotToken string
		var gotCtx context.Context
		var gotFunc func(string) *IPAddressDetails

		GetIPInfo([]byte(`{"ip":"1.2.3.4","token_id":"tid"}`), func(ctx context.Context, IPAddress, tokenID string, get func(string) *IPAddressDetails) {
			gotCtx, gotIP, gotToken, gotFunc = ctx, IPAddress, tokenID, get
		})

		if gotCtx == nil || gotIP != "1.2.3.4" || gotToken != "tid" || gotFunc == nil {
			t.Errorf("got %v, %q, %q, func defined: %v", gotCtx, gotIP, gotToken, gotFunc != nil)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		GetIPInfo([]byte("{"), func(context.Context, string, string, func(string) *IPAddressDetails) {
			t.Error("the function must not be called")
		})
	})

	t.Run("undefined function", func(t *testing.T) {
		GetIPInfo([]byte(`{"ip":"1.2.3.4"}`), nil)
	})
}

func TestGetIPAddressDetails(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		var gotPath, gotKey string
		useAPI(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotKey = r.URL.Path, r.URL.Query().Get("access_key")
			_, _ = w.Write([]byte(`{"region_name":"Occitanie","city":"Toulouse","country_code":"FR","country_name":"France"}`))
		})

		details := GetIPAddressDetails("1.2.3.4")

		want := IPAddressDetails{Region: "Occitanie", City: "Toulouse", CountryCode: "FR", CountryName: "France"}
		if details == nil || *details != want {
			t.Errorf("details = %+v, want %+v", details, want)
		}
		if gotPath != "/1.2.3.4" || gotKey != "secret key" {
			t.Errorf("path = %q, key = %q", gotPath, gotKey)
		}
	})

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantLog string
	}{
		{"unexpected status", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) }, "unexpected status 502"},
		{"invalid JSON", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{")) }, "unexpected end of JSON input"},
		{"no location", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"city":"Toulouse"}`)) }, "get no location from the API"},
		{
			"truncated body",
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "100")
				_, _ = w.Write([]byte("{"))
			},
			"unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workertest.Hook.Reset()
			useAPI(t, tt.handler)

			if details := GetIPAddressDetails("1.2.3.4"); details != nil {
				t.Errorf("details = %+v, want nil", details)
			}

			entry := workertest.Hook.LastEntry()
			if entry == nil || !strings.Contains(entry.Data["error"].(string), tt.wantLog) || entry.Data["IP"] != "1.2.3.4" {
				t.Errorf("log = %+v, want %q", entry, tt.wantLog)
			}
		})
	}

	t.Run("API unreachable, the key is not logged", func(t *testing.T) {
		workertest.Hook.Reset()
		useAPI(t, func(http.ResponseWriter, *http.Request) {})
		config.Main.IPStack.Host = "http://127.0.0.1:1"

		if details := GetIPAddressDetails("1.2.3.4"); details != nil {
			t.Errorf("details = %+v, want nil", details)
		}

		entry := workertest.Hook.LastEntry()
		if entry == nil || strings.Contains(entry.Data["error"].(string), "secret") {
			t.Errorf("log = %+v", entry)
		}
	})

	t.Run("invalid URL, the key is not logged", func(t *testing.T) {
		workertest.Hook.Reset()
		useAPI(t, func(http.ResponseWriter, *http.Request) {})
		config.Main.IPStack.Host = "http://[::1"

		if details := GetIPAddressDetails("1.2.3.4"); details != nil {
			t.Errorf("details = %+v, want nil", details)
		}

		entry := workertest.Hook.LastEntry()
		if entry == nil || strings.Contains(entry.Data["error"].(string), "secret") {
			t.Errorf("log = %+v", entry)
		}
	})
}
