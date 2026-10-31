// Package plugintest provides helpers shared by the tests of the MongoDB plugin: a gin context bound to a stub
// database, and optionally to a signed-in user.
package plugintest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/ctxqueue"

	ctxAuthToken "github.com/cygy/ginamite/api/context/authtoken"
	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/common/kafka"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/IBM/sarama/mocks"
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

// NewContext returns a gin context bound to a request, to the stub database and to the recorder of the response.
// The locale of the context is "en", see InitLocalization. The configuration of the context is the default one, and
// can be changed with ctxconfig.Get, and has no auth token hooks.
func NewContext(db *mock.Database, method, target string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, body)
	session.Set(c, db)
	ctxconfig.Set(c, config.NewConfiguration())
	ctxAuthToken.SetHooks(c, authentication.Hooks{})
	ctxqueue.Set(c, senderInUse)
	ctxLocale.Set(c, "en", "en")

	return c, recorder
}

// SignedInContext returns a gin context like NewContext, with the user signed in: the auth token is parsed and the
// user is built from it, as the authentication middleware does.
func SignedInContext(t *testing.T, db *mock.Database, user model.User) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	c, recorder := NewContext(db, http.MethodGet, "/", nil)
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

// InitLocalization initializes the translation engine with the "en" messages.
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

	localization.Initialize(dir, []string{"en"})
}

// Queue captures the messages sent to Kafka.
type Queue struct {
	// Sender is the queue sending its messages to the capture.
	Sender *queue.Queue

	t        *testing.T
	producer *mocks.AsyncProducer
}

// senderInUse is the queue of the contexts created by NewContext. It is the one of the capture, if any.
var senderInUse = queue.New(kafka.NewClient())

// CaptureQueue returns a Queue capturing the messages sent by its Sender, which is the queue of the contexts created
// by NewContext for the duration of the test.
func CaptureQueue(t *testing.T) *Queue {
	t.Helper()

	config := mocks.NewTestConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewAsyncProducer(t, config)
	for range maxMessages {
		producer.ExpectInputAndSucceed()
	}

	client := kafka.NewClient()
	client.HasProducer = true
	client.Producer = producer

	sender := queue.New(client)
	old := senderInUse
	senderInUse = sender
	t.Cleanup(func() { senderInUse = old })

	return &Queue{Sender: sender, t: t, producer: producer}
}

// maxMessages is the number of messages a Queue can capture.
const maxMessages = 10

// Next returns the next message sent to Kafka, and ok is false if there is none after a short delay.
func (q *Queue) Next() (message queue.Message, ok bool) {
	q.t.Helper()

	select {
	case sent := <-q.producer.Successes():
		value, err := sent.Value.Encode()
		if err != nil {
			q.t.Fatal(err)
		}
		if err := json.Unmarshal(value, &message); err != nil {
			q.t.Fatal(err)
		}

		return message, true
	case <-time.After(time.Second):
		return message, false
	}
}
