package location

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/worker/internal/workertest"
)

// useAPI starts a server for the handler for the duration of the test, and returns the configuration of the API to it.
func useAPI(t *testing.T, handler http.HandlerFunc) config.GeoIPConfiguration {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return config.GeoIPConfiguration{Host: server.URL}
}

func TestGetIPInfo(t *testing.T) {
	// configured is a geoip configuration with a host, which the tests below never reach.
	configured := config.GeoIPConfiguration{Host: "http://geoip.invalid"}

	t.Run("saves the location", func(t *testing.T) {
		var gotIP, gotToken string
		var gotCtx context.Context
		var gotFunc DetailsFunc

		err := GetIPInfo([]byte(`{"ip":"1.2.3.4","token_id":"tid"}`), configured, func(ctx context.Context, IPAddress, tokenID string, get DetailsFunc) error {
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

	t.Run("no host: the function is not called", func(t *testing.T) {
		err := GetIPInfo([]byte(`{"ip":"1.2.3.4","token_id":"tid"}`), config.GeoIPConfiguration{}, func(context.Context, string, string, DetailsFunc) error {
			t.Error("the function must not be called: the database would be read for nothing")

			return nil
		})
		if err != nil {
			t.Errorf("GetIPInfo() error = %v", err)
		}
	})

	t.Run("the function uses the API", func(t *testing.T) {
		geoIP := useAPI(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"country_code":"FR","country_name":"France","region":"Occitanie","city":"Toulouse"}`))
		})

		err := GetIPInfo([]byte(`{"ip":"1.2.3.4","token_id":"tid"}`), geoIP, func(_ context.Context, IPAddress, _ string, get DetailsFunc) error {
			details, err := get(IPAddress)
			if err != nil || details == nil || details.City != "Toulouse" {
				t.Errorf("details = %+v, error = %v", details, err)
			}

			return nil
		})
		if err != nil {
			t.Errorf("GetIPInfo() error = %v", err)
		}
	})

	t.Run("the error of the function", func(t *testing.T) {
		boom := errors.New("boom")
		err := GetIPInfo([]byte(`{"ip":"1.2.3.4"}`), configured, func(context.Context, string, string, DetailsFunc) error {
			return boom
		})
		if !errors.Is(err, boom) || broker.IsPermanent(err) {
			t.Errorf("GetIPInfo() error = %v, want the error of the function, which a retry can fix", err)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		err := GetIPInfo([]byte("{"), configured, func(context.Context, string, string, DetailsFunc) error {
			t.Error("the function must not be called")

			return nil
		})
		if !broker.IsPermanent(err) {
			t.Errorf("GetIPInfo() error = %v, want a permanent error", err)
		}
	})

	t.Run("undefined function", func(t *testing.T) {
		if err := GetIPInfo([]byte(`{"ip":"1.2.3.4"}`), config.GeoIPConfiguration{}, nil); err != nil {
			t.Errorf("GetIPInfo() error = %v", err)
		}
	})
}

func TestGetIPAddressDetails(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		var gotPath, gotAccept string
		geoIP := useAPI(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath, gotAccept = r.URL.Path, r.Header.Get("Accept")
			_, _ = w.Write([]byte(`{"ip":"2a01:cb00::1","country_code":"FR","country_name":"France","flag":"🇫🇷",` +
				`"continent_code":"EU","region":"Occitanie","region_code":"FR-OCC","city":"Toulouse","source":"dbip"}`))
		})
		geoIP.Host += "/"

		details := GetIPAddressDetails(geoIP, "2a01:cb00::1")

		want := IPAddressDetails{
			Region: "Occitanie", RegionCode: "FR-OCC", City: "Toulouse", CountryCode: "FR", CountryName: "France",
			Source: "dbip",
		}
		if details == nil || *details != want {
			t.Errorf("details = %+v, want %+v", details, want)
		}
		if gotPath != "/v1/ip/2a01:cb00::1" || gotAccept != "application/json" {
			t.Errorf("path = %q, Accept = %q", gotPath, gotAccept)
		}
	})

	t.Run("no host: no request, no log", func(t *testing.T) {
		workertest.Hook.Reset()

		if details := GetIPAddressDetails(config.GeoIPConfiguration{}, "1.2.3.4"); details != nil {
			t.Errorf("details = %+v, want nil", details)
		}
		if entry := workertest.Hook.LastEntry(); entry != nil {
			t.Errorf("log = %+v, want none", entry)
		}
	})

	t.Run("unknown address: no log", func(t *testing.T) {
		workertest.Hook.Reset()
		geoIP := useAPI(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) })

		if details := GetIPAddressDetails(geoIP, "1.2.3.4"); details != nil {
			t.Errorf("details = %+v, want nil", details)
		}
		if entry := workertest.Hook.LastEntry(); entry != nil {
			t.Errorf("log = %+v, want none", entry)
		}
	})

	unreachable := useAPI(t, func(http.ResponseWriter, *http.Request) {})
	unreachable.Host = "http://127.0.0.1:1"
	invalidURL := config.GeoIPConfiguration{Host: "http://[::1"}
	slow := useAPI(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
		w.WriteHeader(http.StatusOK)
	})
	slow.Timeout = 10 * time.Millisecond

	tests := []struct {
		name    string
		geoIP   config.GeoIPConfiguration
		wantLog string
	}{
		{"unreachable", unreachable, "connect"},
		{"invalid URL", invalidURL, "missing ']'"},
		{"timeout", slow, "deadline exceeded"},
		{"server error", useAPI(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) }), "unexpected status 502"},
		{"bad request", useAPI(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest) }), "unexpected status 400"},
		{"invalid JSON", useAPI(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{")) }), "unexpected end of JSON input"},
		{"no location", useAPI(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"city":"Toulouse"}`)) }), "get no location from the API"},
		{
			"truncated body",
			useAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "100")
				_, _ = w.Write([]byte("{"))
			}),
			"unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" is logged, not retried", func(t *testing.T) {
			workertest.Hook.Reset()

			if details := GetIPAddressDetails(tt.geoIP, "1.2.3.4"); details != nil {
				t.Errorf("details = %+v, want nil", details)
			}

			entry := workertest.Hook.LastEntry()
			if entry == nil || !strings.Contains(entry.Data["error"].(string), tt.wantLog) || entry.Data["IP"] != "1.2.3.4" {
				t.Errorf("log = %+v, want %q", entry, tt.wantLog)
			}
		})
	}
}
