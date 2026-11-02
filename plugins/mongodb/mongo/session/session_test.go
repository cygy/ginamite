package session

import (
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMain(m *testing.M) {
	log.Initialize("test", false)
	os.Exit(m.Run())
}

// offline returns a session whose client is never connected: the driver connects lazily.
func offline(t *testing.T) *Session {
	t.Helper()

	client, err := mongo.Connect(options.Client().SetHosts([]string{"127.0.0.1:1"}).SetServerSelectionTimeout(50e6))
	if err != nil {
		t.Fatal(err)
	}

	return &Session{Client: client, DatabaseName: "app"}
}

func TestNewSession(t *testing.T) {
	if s := NewSession(); s == nil || s.Client != nil || s.DatabaseName != "" {
		t.Errorf("NewSession() = %+v", s)
	}
}

func TestSession_Initialize_Unreachable(t *testing.T) {
	tests := []struct {
		name     string
		username string
	}{
		{"without user", ""},
		{"with user", "john"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The timeout is in seconds.
			if err := NewSession().Initialize("127.0.0.1:1", "app", tt.username, "secret", 1); err == nil {
				t.Error("Initialize() must return an error")
			}
		})
	}
}

func TestSession_Close(t *testing.T) {
	NewSession().Close()

	s := offline(t)
	s.Close()
}

func TestSession_Database(t *testing.T) {
	if got := offline(t).Database().Name(); got != "app" {
		t.Errorf("Database().Name() = %q", got)
	}
}

func TestSession_Store(t *testing.T) {
	stub := mock.NewDatabase("stub")

	s := NewSession()
	s.Store = stub
	if s.Database() != store.Database(stub) {
		t.Error("Database() must return the store when there is one")
	}

	// Without client, the session can still be closed.
	s.Close()
}
