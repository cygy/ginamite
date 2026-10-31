package session

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/store"
	"github.com/cygy/ginamite/common/mongo/store/mock"

	"github.com/gin-gonic/gin"
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
			defer func() {
				if recover() == nil {
					t.Error("Initialize() must panic after logging the error")
				}
			}()

			// The timeout is in seconds.
			NewSession().Initialize("127.0.0.1:1", "app", tt.username, "secret", 1)
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

func TestSession_InjectAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := offline(t)
	router := gin.New()
	router.GET("/with", s.Inject(), func(c *gin.Context) {
		if db := Get(c); db == nil || db.Name() != "app" {
			t.Errorf("Get() = %v", db)
		}
		c.Status(http.StatusNoContent)
	})
	router.GET("/without", func(c *gin.Context) {
		if db := Get(c); db != nil {
			t.Errorf("Get() = %v, want nil", db)
		}
		c.Status(http.StatusNoContent)
	})

	for _, path := range []string{"/with", "/without"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusNoContent {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestSession_Store(t *testing.T) {
	stub := mock.NewDatabase("stub")

	s := NewSession()
	s.Store = stub
	if s.Database() != store.Database(stub) {
		t.Error("Database() must return the store when there is one")
	}

	// Without client, the session can still be injected and closed.
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	s.Inject()(c)
	if Get(c) != store.Database(stub) {
		t.Errorf("Get() = %v", Get(c))
	}
	s.Close()
}

func TestSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	stub := mock.NewDatabase("stub")
	Set(c, stub)

	if Get(c) != store.Database(stub) {
		t.Errorf("Get() = %v", Get(c))
	}
}
