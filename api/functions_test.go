package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/cygy/ginamite/common/log/logtest"
	"github.com/gin-gonic/gin"
)

func TestNewServer(t *testing.T) {
	s := NewServer()

	if s == nil || s.customRoutes == nil || len(s.customRoutes) != 0 {
		t.Errorf("NewServer() = %+v, want a server without custom routes", s)
	}
}

func TestServer_idValidator(t *testing.T) {
	validity := func(s *Server) (valid, invalid bool) {
		router := gin.New()
		router.Use(s.idValidator())
		router.GET("/", func(c *gin.Context) { valid, invalid = request.IsValidID(c, "ok"), request.IsValidID(c, "anything") })
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		return valid, invalid
	}

	s := NewServer()
	if valid, invalid := validity(s); !valid || !invalid {
		t.Error("without a function, every ID must stay valid")
	}

	s.Hooks.IsValidID = func(id string) bool { return id == "ok" }
	if valid, invalid := validity(s); !valid || invalid {
		t.Error("the function of the server must validate the IDs")
	}
}

func TestServer_AddCustomRoutes(t *testing.T) {
	s := NewServer()
	calls := 0
	configure := func(*gin.RouterGroup, route.Middlewares, bool, func(bool, []string) gin.HandlerFunc) { calls++ }

	s.AddCustomRoutes(configure)
	s.AddCustomRoutes(configure)

	if len(s.customRoutes) != 2 {
		t.Fatalf("len(customRoutes) = %d, want 2", len(s.customRoutes))
	}
	for _, r := range s.customRoutes {
		r(nil, route.Middlewares{}, false, nil)
	}
	if calls != 2 {
		t.Errorf("custom routes called %d times, want 2", calls)
	}
}

// cacheMessage returns the raw kafka message of the type with the payload.
func cacheMessage(t *testing.T, messageType string, payload any) []byte {
	t.Helper()

	value, err := json.Marshal(queue.Message{Type: messageType, Payload: payload})
	if err != nil {
		t.Fatalf("unable to encode the message: %v", err)
	}

	return value
}

func TestServer_CreateConsumers(t *testing.T) {
	expiration := time.Now().Add(time.Hour)

	tests := []struct {
		name         string
		handler      func(string) bool
		value        []byte
		wantRevoked  bool
		wantWarnings int
		wantFailure  bool // the message can not be decoded: it is dead-lettered without a retry
	}{
		{
			name:        "invalidated auth token",
			value:       cacheMessage(t, queue.MessageCacheInvalidAuthToken, queue.CacheAuthToken{Tokens: map[string]time.Time{"revoked-token": expiration}}),
			wantRevoked: true,
		},
		{name: "undecodable message", value: []byte("nope"), wantFailure: true},
		{name: "undecodable payload", value: cacheMessage(t, queue.MessageCacheInvalidAuthToken, "nope"), wantFailure: true},
		{name: "unknown message without handler", value: cacheMessage(t, "unknown", nil), wantWarnings: 1},
		{name: "unknown message ignored by the handler", handler: func(string) bool { return false }, value: cacheMessage(t, "unknown", nil), wantWarnings: 1},
		{name: "unknown message handled", handler: func(string) bool { return true }, value: cacheMessage(t, "unknown", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook, restore := logtest.Replace(&log.Main, slog.LevelInfo)
			defer restore()
			s := NewServer()
			s.Hooks.NewMessageFromQueueCache = tt.handler

			consumers := s.CreateConsumers()
			if len(consumers) != 1 || consumers[0].Topic != queue.TopicCache {
				t.Fatalf("consumers = %+v, want one consumer of the topic %q", consumers, queue.TopicCache)
			}
			if !consumers[0].Idempotent {
				t.Error("the invalidation of the cache can be handled twice: no deduplication needed")
			}
			if other := s.CreateConsumers()[0]; !strings.HasPrefix(consumers[0].Group, queue.TopicCache+"_") || consumers[0].Group == other.Group {
				t.Errorf("groups = %q, %q: each instance needs its own group", consumers[0].Group, other.Group)
			}
			hook.Reset()

			err := consumers[0].Action(tt.value)

			if (err != nil) != tt.wantFailure || (err != nil && !broker.IsPermanent(err)) {
				t.Errorf("Action() error = %v, want a permanent error: %v", err, tt.wantFailure)
			}

			if got := s.RevokedTokens.IsRevoked("revoked-token"); got != tt.wantRevoked {
				t.Errorf("IsRevoked() = %v, want %v", got, tt.wantRevoked)
			}
			warnings := 0
			for _, entry := range hook.AllEntries() {
				if entry.Level == slog.LevelWarn {
					warnings++
				}
			}
			if warnings != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d", warnings, tt.wantWarnings)
			}
		})
	}
}

// countWarnings returns the number of warnings logged while calling fn.
func countWarnings(fn func()) int {
	hook, restore := logtest.Replace(&log.Main, slog.LevelInfo)
	defer restore()

	fn()

	count := 0
	for _, entry := range hook.AllEntries() {
		if entry.Level == slog.LevelWarn {
			count++
		}
	}

	return count
}

func TestServer_configureProxies(t *testing.T) {
	tests := []struct {
		name    string
		proxies []string
		remote  string
		header  string
		want    string
	}{
		{"default: no proxy trusted", []string{}, "10.1.2.3:1234", "1.2.3.4", "10.1.2.3"},
		{"trusted proxy", []string{"10.0.0.0/8"}, "10.1.2.3:1234", "1.2.3.4", "1.2.3.4"},
		{"proxy not trusted", []string{"192.168.0.0/16"}, "10.1.2.3:1234", "1.2.3.4", "10.1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()
			s.Config.Server.TrustedProxies = tt.proxies

			engine := gin.New()
			if err := s.configureProxies(engine); err != nil {
				t.Fatal(err)
			}

			var got string
			engine.GET("/", func(c *gin.Context) { got = c.ClientIP() })

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remote
			req.Header.Set("X-Forwarded-For", tt.header)
			engine.ServeHTTP(httptest.NewRecorder(), req)

			if got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServer_configureProxies_Invalid(t *testing.T) {
	s := NewServer()
	s.Config.Server.TrustedProxies = []string{"not an address"}

	if err := s.configureProxies(gin.New()); err == nil {
		t.Error("configureProxies() must return an error")
	}
}

func TestServer_checkProviders(t *testing.T) {
	tests := []struct {
		name  string
		setup func(s *Server)
		want  string
	}{
		{"no provider", func(s *Server) {}, ""},
		{"facebook without secret", func(s *Server) {
			s.Config.APIFeatures = []string{FeatureFacebook}
			s.Config.Facebook.AppID = "id"
		}, "facebook.app_secret"},
		{"facebook", func(s *Server) {
			s.Config.APIFeatures = []string{FeatureFacebook}
			s.Config.Facebook.AppID, s.Config.Facebook.AppSecret = "id", "secret"
		}, ""},
		{"google without client ID", func(s *Server) { s.Config.APIFeatures = []string{FeatureGoogle} }, "google.client_id"},
		{"google", func(s *Server) {
			s.Config.APIFeatures = []string{FeatureGoogle}
			s.Config.Google.ClientID = "id"
		}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()
			tt.setup(s)

			err := s.checkProviders()
			if (tt.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), tt.want)) {
				t.Errorf("checkProviders() = %v, want an error naming %q", err, tt.want)
			}
		})
	}
}

func TestServer_Run_SetUpErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(s *Server)
		want  string
	}{
		{"invalid trusted proxies", func(s *Server) { s.Config.Server.TrustedProxies = []string{"not an address"} }, "trusted_proxies"},
		{"invalid templates", func(s *Server) { s.Config.TemplatesPath = t.TempDir() }, "templates"},
		{"unknown feature", func(s *Server) { s.Config.APIFeatures = []string{"nope"} }, "nope"},
		{"missing secret", func(s *Server) { s.Config.APIFeatures = []string{FeatureTerms} }, "jwt.secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer()
			s.Terms = noopTerms{}
			tt.setup(s)

			if err := s.Run(t.Context()); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Run() = %v, want an error about %q", err, tt.want)
			}
		})
	}
}

func TestServer_warnWeakSecret(t *testing.T) {
	s := NewServer()

	s.Config.JWT.Secret = "short"
	if got := countWarnings(s.warnWeakSecret); got != 1 {
		t.Errorf("short secret: warnings = %d, want 1", got)
	}

	s.Config.JWT.Secret = strings.Repeat("x", 32)
	if got := countWarnings(s.warnWeakSecret); got != 0 {
		t.Errorf("long secret: warnings = %d, want 0", got)
	}
}
