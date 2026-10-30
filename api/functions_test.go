package api

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	tokencache "github.com/cygy/ginamite/api/cache/authtoken"
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

func TestServer_setUpIDValidation(t *testing.T) {
	old := request.IsValidID
	t.Cleanup(func() { request.IsValidID = old })

	s := NewServer()
	s.setUpIDValidation()
	if !request.IsValidID("anything") {
		t.Error("without a function, every ID must stay valid")
	}

	s.Functions.IsValidID = func(id string) bool { return id == "ok" }
	s.setUpIDValidation()
	if !request.IsValidID("ok") || request.IsValidID("anything") {
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
			hook.Reset()

			consumers[0].Action(tt.value)

			if got := tokencache.IsRevoked("revoked-token"); got != tt.wantRevoked {
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
			tokencache.Remove("revoked-token")
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
