package mail

import (
	"context"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/localization"
)

func TestMailer_message(t *testing.T) {
	mailer, _ := stub(t)

	t.Run("without links", func(t *testing.T) {
		m := mailer.message(content{key: "k", template: TemplateNotification, recipient: address, username: "john", locale: "en"})

		checkCommon(t, m, "k", TemplateNotification)
		for _, name := range []string{"ConfirmURL", "ConfirmButton", "CancelURL", "CancelButton"} {
			if _, ok := m.HTMLVars[name]; ok {
				t.Errorf("%s must not be set", name)
			}
		}
	})

	t.Run("with links and variables", func(t *testing.T) {
		m := mailer.message(content{
			key: "k", template: TemplateConfirmationAndCancel, recipient: address, username: "john", locale: "en",
			vars: localization.H{"Extra": "x"}, confirmURL: "https://c", cancelURL: "https://x",
		})

		checkCommon(t, m, "k", TemplateConfirmationAndCancel)
		contains(t, "HTML message", m.HTMLVars["Message"], "Extra:x")
		contains(t, "text message", m.TextVars["Message"], "Extra:x", "ConfirmURL:https://c", "CancelURL:https://x")
		if m.HTMLVars["ConfirmURL"] != "https://c" || m.HTMLVars["CancelURL"] != "https://x" {
			t.Errorf("HTMLVars = %v", m.HTMLVars)
		}
		if m.Subject == "" || strings.Contains(m.Subject, "Extra") {
			t.Errorf("the subject only has the username: %q", m.Subject)
		}
	})
}

func TestMailer_processURL(t *testing.T) {
	mailer, _ := stub(t)

	if got, want := mailer.processURL("route", "fr", "id", "code"), "https://web/fr/route?id=id&key=code"; got != want {
		t.Errorf("processURL = %q, want %q", got, want)
	}
}

func TestAlwaysValid(t *testing.T) {
	if ok, err := alwaysValid(context.Background(), ""); !ok || err != nil {
		t.Errorf("alwaysValid = %v, %v", ok, err)
	}
}
