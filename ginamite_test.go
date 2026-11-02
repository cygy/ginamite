package ginamite

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

func init() {
	gin.SetMode(gin.TestMode)

	logger, _ := logtest.NewNullLogger()
	logger.ExitFunc = func(int) {}
	log.Main = logger.WithField("server", "test")
}

type testPlugin struct {
	configured []*Server
	err        error // Returned by Configure.
}

func (p *testPlugin) Configure(s *Server) error {
	p.configured = append(p.configured, s)
	return p.err
}

// testConnector is a Plugin which records its connection into events.
type testConnector struct {
	testPlugin
	name       string
	events     *[]string
	connectErr error // Returned by Connect.
}

func (c *testConnector) Connect(configuration *config.Configuration) error {
	*c.events = append(*c.events, "connect "+c.name+" "+configuration.Database.Service)
	return c.connectErr
}
func (c *testConnector) Close() { *c.events = append(*c.events, "close "+c.name) }

func TestNewServer(t *testing.T) {
	s := NewServer()

	if s.API == nil || s.Worker == nil || s.Web == nil || s.plugins == nil || s.ApplicationConfiguration.FileRequired {
		t.Fatalf("server = %+v", s)
	}

	f := s.API.Hooks
	if f.NewMessageFromQueueCache == nil || f.BuildTokenFromID == nil || f.ExtraPropertiesForTokenWithID == nil ||
		f.ExtendTokenExpirationDateFromID == nil || f.MiddlewareIsValidAuthToken == nil || f.MiddlewareGetUserAbilities == nil {
		t.Fatal("a placeholder function is missing")
	}
}

func TestNewServer_SharedServices(t *testing.T) {
	s := NewServer()

	if s.Config == nil || s.Deduplicator == nil || s.Queue == nil || s.Localization == nil || s.Timezones == nil {
		t.Fatalf("server = %+v", s)
	}
	if s.API.Queue != s.Queue || s.Worker.Queue != s.Queue {
		t.Error("the queue must be shared with the API and the worker")
	}
}

func TestNewServer_Placeholders(t *testing.T) {
	f := NewServer().API.Hooks
	c, _ := gin.CreateTestContext(nil)

	if !f.NewMessageFromQueueCache("message") {
		t.Error("NewMessageFromQueueCache() = false")
	}
	if token, err := f.BuildTokenFromID(context.Background(), "id"); err != nil || token == nil {
		t.Errorf("BuildTokenFromID() = %v, %v", token, err)
	}
	if props, err := f.ExtraPropertiesForTokenWithID("id"); err != nil || props == nil || len(props) != 0 {
		t.Errorf("ExtraPropertiesForTokenWithID() = %v, %v", props, err)
	}
	if err := f.ExtendTokenExpirationDateFromID(context.Background(), "id", time.Hour); err != nil {
		t.Errorf("ExtendTokenExpirationDateFromID() = %v", err)
	}
	if valid, err := f.MiddlewareIsValidAuthToken(c, "id"); valid || err != nil {
		t.Errorf("MiddlewareIsValidAuthToken() = %v, %v, want false, nil", valid, err)
	}
	if ok, abilities, err := f.MiddlewareGetUserAbilities(c, "id"); ok || abilities == nil || len(abilities) != 0 || err != nil {
		t.Errorf("MiddlewareGetUserAbilities() = %v, %v, %v", ok, abilities, err)
	}
}

func TestServer_AddPlugin(t *testing.T) {
	s := NewServer()
	first, second := &testPlugin{}, &testPlugin{}

	s.AddPlugin(first)
	s.AddPlugin(second)

	if len(s.plugins) != 2 || s.plugins[0] != first || s.plugins[1] != second {
		t.Errorf("plugins = %v", s.plugins)
	}
}

func TestServer_UseBroker(t *testing.T) {
	s := NewServer()
	memory := broker.NewMemory()

	s.UseBroker(memory)

	if s.Broker != memory || s.API.Queue != s.Queue || s.Worker.Queue != s.Queue {
		t.Fatal("the broker and the queues are not shared")
	}

	s.Queue.SendMail("welcome", map[string]string{"email": "a@b.c"})
	if sent := memory.Sent(); len(sent) != 1 || sent[0].Topic != queue.TopicMail {
		t.Errorf("sent = %+v", sent)
	}
}
