// Package plugintest provides helpers shared by the tests of the MongoDB plugin: a gin context, and optionally a
// signed-in user.
package plugintest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/ctxkey"
	"github.com/cygy/ginamite/common/notifications"

	"github.com/cygy/ginamite/common/timezone"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// Secret is the secret used to sign the tokens of the tests.
const Secret = "secret"

func init() {
	gin.SetMode(gin.TestMode)
	log.Initialize("test", false)
	log.Main.Logger.SetOutput(io.Discard)
}

// NewContext returns a gin context bound to a request and to the recorder of the response.
// The locale of the context is "en", see InitLocalization. The configuration of the context is the default one, and
// can be changed with ctxkey.Config.Get, and has no auth token hooks and no timezones.
func NewContext(method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequestWithContext(context.Background(), method, target, body)
	ctxkey.Config.Set(c, config.NewConfiguration())
	ctxAuthToken.SetHooks(c, authentication.Hooks{})
	ctxkey.Queue.Set(c, senderInUse)
	ctxkey.Timezones.Set(c, timezone.NewConfiguration())
	ctxkey.Notifications.Set(c, notifications.NewRegistry())
	if Localization != nil {
		ctxkey.Localization.Set(c, Localization)
	}
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}

// SignedInContext returns a gin context like NewContext, with the user signed in: the auth token is parsed and the
// user is built from it, as the authentication middleware does.
func SignedInContext(t *testing.T, user model.User) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	c, recorder := NewContext(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+signToken(t, user.ID.Hex(), "token-1"))

	buildUser := func(*jwt.Token) (any, string, error) { return user, user.Settings.Locale, nil }
	ctxAuthToken.Parse(Secret, buildUser, true)(c)
	if c.IsAborted() {
		t.Fatalf("the user is not signed in: %s", recorder.Body.String())
	}
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}

func signToken(t *testing.T, userID, tokenID string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"id":  tokenID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	signed, err := token.SignedString([]byte(Secret))
	if err != nil {
		t.Fatalf("unable to sign the token: %v", err)
	}

	return signed
}

// Localization is the translation engine of the contexts created by NewContext, set by InitLocalization.
var Localization *localization.Engine

// InitLocalization sets the translation engine, see Localization, with the "en" messages for the duration of the test.
func InitLocalization(t *testing.T, messages map[string]string) {
	t.Helper()

	content, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("unable to encode the messages: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "en.all.json"), content, 0o600); err != nil {
		t.Fatalf("unable to write the messages: %v", err)
	}

	old := Localization
	engine, err := localization.Load(dir, []string{"en"})
	if err != nil {
		t.Fatal(err)
	}
	Localization = engine
	t.Cleanup(func() { Localization = old })
}

// Queue captures the messages sent to the broker.
type Queue struct {
	// Sender is the queue sending its messages to the capture.
	Sender *queue.Queue

	t      *testing.T
	broker *broker.Memory
	next   int
}

// senderInUse is the queue of the contexts created by NewContext. It is the one of the capture, if any.
var senderInUse = queue.New(broker.NewMemory())

// CaptureQueue returns a Queue capturing the messages sent by its Sender, which is the queue of the contexts created
// by NewContext for the duration of the test.
func CaptureQueue(t *testing.T) *Queue {
	t.Helper()

	memory := broker.NewMemory()
	sender := queue.New(memory)
	old := senderInUse
	senderInUse = sender
	t.Cleanup(func() { senderInUse = old })

	return &Queue{Sender: sender, t: t, broker: memory}
}

// Next returns the next message sent to the broker, and ok is false if there is none.
func (q *Queue) Next() (message queue.Message, ok bool) {
	q.t.Helper()

	sent := q.broker.Sent()
	if q.next >= len(sent) {
		return message, false
	}

	if err := json.Unmarshal([]byte(sent[q.next].Message), &message); err != nil {
		q.t.Fatal(err)
	}
	q.next++

	return message, true
}
