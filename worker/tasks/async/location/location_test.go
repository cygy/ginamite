package location

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

// useAPI starts a server for the handler for the duration of the test, and returns the configuration of the API to it.
func useAPI(t *testing.T, handler http.HandlerFunc) config.IPStackConfiguration {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return config.IPStackConfiguration{Host: server.URL, Key: "secret key"}
}

func TestGetIPInfo(t *testing.T) {
	t.Run("saves the location", func(t *testing.T) {
		var gotIP, gotToken string
		var gotCtx context.Context
		var gotFunc DetailsFunc

		err := GetIPInfo([]byte(`{"ip":"1.2.3.4","token_id":"tid"}`), config.IPStackConfiguration{}, func(ctx context.Context, IPAddress, tokenID string, get DetailsFunc) error {
			gotCtx, gotIP, gotToken, gotFunc = ctx, IPAddress, tokenID, get

			return nil
		})
		if err != nil {
			t.Errorf("GetIPInfo() error = %v", err)
		}

		if gotCtx == nil || gotIP != "1.2.3.4" || gotToken != "tid" || gotFunc == nil {
			t.Errorf("got %v, %q, %q, func defined: %v", gotCtx, gotIP, gotToken, gotFunc != nil)
		}
	})

	t.Run("the error of the function", func(t *testing.T) {
		boom := errors.New("boom")
		err := GetIPInfo([]byte(`{"ip":"1.2.3.4"}`), config.IPStackConfiguration{}, func(context.Context, string, string, DetailsFunc) error {
			return boom
		})
		if !errors.Is(err, boom) || broker.IsPermanent(err) {
			t.Errorf("GetIPInfo() error = %v, want the error of the function, which a retry can fix", err)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		err := GetIPInfo([]byte("{"), config.IPStackConfiguration{}, func(context.Context, string, string, DetailsFunc) error {
			t.Error("the function must not be called")

			return nil
		})
		if !broker.IsPermanent(err) {
			t.Errorf("GetIPInfo() error = %v, want a permanent error", err)
		}
	})

	t.Run("undefined function", func(t *testing.T) {
		if err := GetIPInfo([]byte(`{"ip":"1.2.3.4"}`), config.IPStackConfiguration{}, nil); err != nil {
			t.Errorf("GetIPInfo() error = %v", err)
		}
	})
}

func TestGetIPAddressDetails(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		var gotPath, gotKey string
		ipStack := useAPI(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotKey = r.URL.Path, r.URL.Query().Get("access_key")
			_, _ = w.Write([]byte(`{"region_name":"Occitanie","city":"Toulouse","country_code":"FR","country_name":"France"}`))
		})

		details, err := GetIPAddressDetails(ipStack, "1.2.3.4")
		if err != nil {
			t.Fatalf("error = %v", err)
		}

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
		{"client error", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }, "unexpected status 403"},
		{"invalid JSON", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{")) }, "unexpected end of JSON input"},
		{"no location", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"city":"Toulouse"}`)) }, "get no location from the API"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workertest.Hook.Reset()
			ipStack := useAPI(t, tt.handler)

			if details, err := GetIPAddressDetails(ipStack, "1.2.3.4"); details != nil || err != nil {
				t.Errorf("details = %+v, error = %v, want neither: asking again would not help", details, err)
			}

			entry := workertest.Hook.LastEntry()
			if entry == nil || !strings.Contains(entry.Data["error"].(string), tt.wantLog) || entry.Data["IP"] != "1.2.3.4" {
				t.Errorf("log = %+v, want %q", entry, tt.wantLog)
			}
		})
	}

	temporary := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{"server error", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) }, "unexpected status 502"},
		{"too many requests", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTooManyRequests) }, "unexpected status 429"},
		{
			"truncated body",
			func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "100")
				_, _ = w.Write([]byte("{"))
			},
			"unexpected EOF",
		},
	}

	for _, tt := range temporary {
		t.Run(tt.name+" is an error to retry", func(t *testing.T) {
			ipStack := useAPI(t, tt.handler)

			details, err := GetIPAddressDetails(ipStack, "1.2.3.4")
			if details != nil || err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("details = %+v, error = %v, want %q", details, err, tt.wantErr)
			}
		})
	}

	t.Run("API unreachable, the key is not in the error", func(t *testing.T) {
		ipStack := useAPI(t, func(http.ResponseWriter, *http.Request) {})
		ipStack.Host = "http://127.0.0.1:1"

		details, err := GetIPAddressDetails(ipStack, "1.2.3.4")
		if details != nil || err == nil || strings.Contains(err.Error(), "secret") {
			t.Errorf("details = %+v, error = %v", details, err)
		}
	})

	t.Run("invalid URL, the key is not in the error", func(t *testing.T) {
		ipStack := useAPI(t, func(http.ResponseWriter, *http.Request) {})
		ipStack.Host = "http://[::1"

		details, err := GetIPAddressDetails(ipStack, "1.2.3.4")
		if details != nil || err == nil || strings.Contains(err.Error(), "secret") {
			t.Errorf("details = %+v, error = %v", details, err)
		}
	})
}
