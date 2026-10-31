package plugintest

import (
	"net/http"
	"testing"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/common/mongo/store/mock"
	"github.com/cygy/ginamite/plugins/mongodb/model"
)

func TestNewContext(t *testing.T) {
	db := mock.NewDatabase("app")

	c, recorder := NewContext(db, http.MethodPost, "/path", nil)
	if c.Request.Method != http.MethodPost || c.Request.URL.Path != "/path" || recorder == nil {
		t.Errorf("request = %+v", c.Request)
	}
	if session.Get(c) == nil || session.Get(c).Name() != "app" {
		t.Error("the stub database must be in the context")
	}
	if ctxLocale.Get(c) != "en" {
		t.Errorf("locale = %q", ctxLocale.Get(c))
	}
	if ctxUser.GetID(c) != "" {
		t.Error("nobody is signed in")
	}
}

func TestSignedInContext(t *testing.T) {
	id := bson.NewObjectID()

	c, _ := SignedInContext(t, mock.NewDatabase("app"), model.User{ID: id, Username: "john"})
	if got := ctxUser.GetID(c); got != id.Hex() {
		t.Errorf("user ID = %q", got)
	}
	if user, ok := ctxUser.Get(c).(model.User); !ok || user.Username != "john" {
		t.Errorf("user = %#v", ctxUser.Get(c))
	}
	if ctxLocale.Get(c) != "en" {
		t.Errorf("locale = %q", ctxLocale.Get(c))
	}
}

func TestInitLocalization(t *testing.T) {
	InitLocalization(t, map[string]string{"hello": "Hello"})

	translate := localization.Translate("en")
	if translate == nil || translate("hello") != "Hello" {
		t.Error("the messages must be loaded")
	}
}

func TestCaptureQueue(t *testing.T) {
	messages := CaptureQueue(t)

	if _, ok := messages.Next(); ok {
		t.Error("nothing was sent")
	}

	messages.Sender.CreateTask("task", map[string]string{"key": "value"})
	message, ok := messages.Next()
	if !ok || message.Type != "task" {
		t.Errorf("message = %+v, %v", message, ok)
	}
}
