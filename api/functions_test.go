package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/cygy/ginamite/api/internal/apitest"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/route"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func TestNewServer(t *testing.T) {
	s := NewServer()

	if s == nil || s.customRoutes == nil || len(s.customRoutes) != 0 {
		t.Errorf("NewServer() = %+v, want a server without custom routes", s)
	}
}

func TestServer_routeFunctions(t *testing.T) {
	s := NewServer()

	// Every function of the server is defined.
	values := reflect.ValueOf(&s.Functions).Elem()
	for i := range values.NumField() {
		field := values.Field(i)
		if field.Kind() == reflect.Func {
			field.Set(reflect.MakeFunc(field.Type(), func([]reflect.Value) []reflect.Value {
				out := make([]reflect.Value, field.Type().NumOut())
				for j := range out {
					out[j] = reflect.Zero(field.Type().Out(j))
				}

				return out
			}))
		}
	}

	account, user, contact, terms := s.routeFunctions()

	// Every function of the routes, except the ones which have a default, is the function of the server.
	for name, functions := range map[string]any{"account": &account, "user": &user, "contact": &contact, "terms": &terms} {
		fields := reflect.ValueOf(functions).Elem()
		for i := range fields.NumField() {
			fieldName := fields.Type().Field(i).Name
			if fieldName == "NewTokenValidator" {
				continue
			}

			if fields.Field(i).IsNil() {
				t.Errorf("%s.%s is not set from the functions of the server", name, fieldName)
			}
		}
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

	s.Functions.IsValidID = func(id string) bool { return id == "ok" }
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

func TestServer_CreateKafkaConsumers(t *testing.T) {
	expiration := time.Now().Add(time.Hour)

	tests := []struct {
		name         string
		handler      func(string) bool
		value        []byte
		wantRevoked  bool
		wantWarnings int
	}{
		{
			name:        "invalidated auth token",
			value:       cacheMessage(t, queue.MessageCacheInvalidAuthToken, queue.CacheAuthToken{Tokens: map[string]time.Time{"revoked-token": expiration}}),
			wantRevoked: true,
		},
		{name: "undecodable message", value: []byte("nope")},
		{name: "unknown message without handler", value: cacheMessage(t, "unknown", nil), wantWarnings: 1},
		{name: "unknown message ignored by the handler", handler: func(string) bool { return false }, value: cacheMessage(t, "unknown", nil), wantWarnings: 1},
		{name: "unknown message handled", handler: func(string) bool { return true }, value: cacheMessage(t, "unknown", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := logtest.NewLocal(log.Main.Logger)
			s := NewServer()
			s.Functions.NewMessageFromQueueCache = tt.handler

			consumers := s.CreateKafkaConsumers()
			if len(consumers) != 1 || consumers[0].Topic != queue.TopicCache {
				t.Fatalf("consumers = %+v, want one consumer of the topic %q", consumers, queue.TopicCache)
			}
			if other := s.CreateKafkaConsumers()[0]; !strings.HasPrefix(consumers[0].Group, queue.TopicCache+"_") || consumers[0].Group == other.Group {
				t.Errorf("groups = %q, %q: each instance needs its own group", consumers[0].Group, other.Group)
			}
			hook.Reset()

			consumers[0].Action(tt.value)

			if got := s.RevokedTokens.IsRevoked("revoked-token"); got != tt.wantRevoked {
				t.Errorf("IsRevoked() = %v, want %v", got, tt.wantRevoked)
			}
			warnings := 0
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel {
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
	hook := logtest.NewLocal(log.Main.Logger)
	defer hook.Reset()

	fn()

	count := 0
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel {
			count++
		}
	}

	return count
}

func TestServer_warnUndefinedFunctions(t *testing.T) {
	handlers := reflect.TypeFor[Handlers]().NumField()

	t.Run("every function undefined", func(t *testing.T) {
		s := NewServer()

		if got := countWarnings(s.warnUndefinedFunctions); got != handlers {
			t.Errorf("warnings = %d, want one per function (%d)", got, handlers)
		}
	})

	t.Run("every function defined", func(t *testing.T) {
		s := NewServer()
		values := reflect.ValueOf(&s.Functions).Elem()
		for i := range values.NumField() {
			field := values.Field(i)
			field.Set(reflect.MakeFunc(field.Type(), func([]reflect.Value) []reflect.Value {
				out := make([]reflect.Value, field.Type().NumOut())
				for j := range out {
					out[j] = reflect.Zero(field.Type().Out(j))
				}

				return out
			}))
		}

		if got := countWarnings(s.warnUndefinedFunctions); got != 0 {
			t.Errorf("warnings = %d, want 0", got)
		}
	})
}
