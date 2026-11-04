package mongodb

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cygy/ginamite"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

func loadConfiguration(t *testing.T, content string) *config.Configuration {
	t.Helper()

	path := filepath.Join(t.TempDir(), "app.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	c := config.NewConfiguration()
	if err := c.Initialize(path, "test", "final"); err != nil {
		t.Fatal(err)
	}

	return c
}

func TestPlugin_Connect_Settings(t *testing.T) {
	t.Run("missing host", func(t *testing.T) {
		err := NewPlugin().Connect(loadConfiguration(t, "api_rate_limiter:\n  limits: {}\n"))
		if err == nil || !strings.Contains(err.Error(), "api_rate_limiter.mongodb") {
			t.Errorf("Connect() = %v, want the missing settings", err)
		}
	})

	t.Run("invalid settings", func(t *testing.T) {
		err := NewPlugin().Connect(loadConfiguration(t, "api_rate_limiter:\n  mongodb:\n    port: abc\n"))
		if err == nil {
			t.Error("Connect() must refuse a port which is not a number")
		}
	})

	t.Run("unreachable server", func(t *testing.T) {
		err := NewPlugin().Connect(loadConfiguration(t, "api_rate_limiter:\n  mongodb:\n    host: 127.0.0.1\n    port: 1\n    database: x\n    timeout: 1\n"))
		if err == nil {
			t.Error("Connect() must fail when the server is not reachable")
		}
	})
}

func TestPlugin_Close_NotConnected(t *testing.T) {
	NewPlugin().Close()
}

func TestPlugin_Configure_NotConnected(t *testing.T) {
	s := ginamite.NewServer()
	if err := NewPlugin().Configure(s); err != nil {
		t.Fatal(err)
	}
	if s.API.Hooks.NewRateLimiter != nil {
		t.Error("a plugin which is not connected must not set the limiters")
	}
}

func TestPlugin_Configure(t *testing.T) {
	db := mock.NewDatabase("limits")
	p := NewPlugin()
	p.database = db

	s := ginamite.NewServer()
	if err := p.Configure(s); err != nil {
		t.Fatal(err)
	}

	if s.API.Hooks.NewRateLimiter == nil || s.API.Hooks.RateLimiterStore != Store {
		t.Errorf("hooks = %+v", s.API.Hooks)
	}
	if created := db.Created(); len(created) != 1 || created[0] != DefaultCollection {
		t.Errorf("created = %v", created)
	}
}

func TestPlugin_Configure_Twice(t *testing.T) {
	p := NewPlugin()
	p.database = mock.NewDatabase("limits")

	s := ginamite.NewServer()
	if err := p.Configure(s); err != nil {
		t.Fatal(err)
	}
	if err := p.Configure(s); err == nil {
		t.Error("a second rate limiting plugin must be refused")
	}
}

func TestPlugin_Configure_Error(t *testing.T) {
	db := mock.NewDatabase("limits")
	db.CreateCollectionErr = errors.New("boom")
	p := NewPlugin()
	p.database = db

	if err := p.Configure(ginamite.NewServer()); err == nil {
		t.Error("Configure() must return the error of the database")
	}
}

func TestPrepare(t *testing.T) {
	t.Run("collection which exists", func(t *testing.T) {
		db := mock.NewDatabase("limits")
		db.CreateCollectionErr = mongo.CommandError{Code: errorCodeNamespaceExists, Name: "NamespaceExists"}

		if err := prepare(t.Context(), db, "counters"); err != nil {
			t.Fatal(err)
		}

		view := db.Coll("counters").Indexes().(*mock.IndexView)
		if len(view.Created) != 1 {
			t.Fatalf("indexes = %+v", view.Created)
		}

		var index options.IndexOptions
		for _, set := range view.Created[0].Options.Opts {
			if err := set(&index); err != nil {
				t.Fatal(err)
			}
		}
		if index.ExpireAfterSeconds == nil || *index.ExpireAfterSeconds != 0 || index.Name == nil || *index.Name != "expireAt_1_1" {
			t.Errorf("index = %+v, want the expiration of the counters", index)
		}
	})

	t.Run("index error", func(t *testing.T) {
		db := mock.NewDatabase("limits")
		db.Coll("counters").Indexes().(*mock.IndexView).Err = errors.New("boom")

		if err := prepare(t.Context(), db, "counters"); err == nil {
			t.Error("prepare() must return the error of the index")
		}
	})
}
