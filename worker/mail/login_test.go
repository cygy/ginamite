package mail

import (
	"context"
	"html/template"
	"testing"

	"github.com/cygy/ginamite/common/weburl"
)

func TestNewLogin(t *testing.T) {
	payload := encode(t, map[string]string{
		"email": address, "username": "john", "locale": "en", "unsubscribe_key": "ukey",
		"ip": "1.2.3.4", "source": "web", "device": "phone", "token_id": "tid", "token_key": "tkey",
	})

	t.Run("sent", func(t *testing.T) {
		delivered := stub(t)
		NewLogin(context.Background(), payload, valid)

		if len(*delivered) != 1 {
			t.Fatalf("delivered %d messages, want 1", len(*delivered))
		}
		m := (*delivered)[0]
		checkCommon(t, m, "login.new", TemplateAlert)

		cancelURL := "https://web/en/" + weburl.AccountLoginCancel + "?id=tid&key=tkey"
		unsubscribeURL := "https://web/en/" + weburl.AccountUnsubscribe + "?key=ukey"
		if m.HTMLVars["CancelURL"] != cancelURL {
			t.Errorf("CancelURL = %v", m.HTMLVars["CancelURL"])
		}
		contains(t, "HTML message", m.HTMLVars["Message"], "IP:1.2.3.4", "Source:web", "Device:phone")
		contains(t, "text message", m.TextVars["Message"], "IP:1.2.3.4", "CancelURL:"+cancelURL)

		footer, ok := m.HTMLVars["Footer"].(template.HTML)
		if !ok {
			t.Fatalf("the HTML footer is a %T, want template.HTML", m.HTMLVars["Footer"])
		}
		contains(t, "HTML footer", footer, "email.login.new.html.footer", "UnsubscribeURL:"+unsubscribeURL)
		contains(t, "text footer", m.TextVars["Footer"], "email.login.new.text.footer", "UnsubscribeURL:"+unsubscribeURL)
	})

	t.Run("invalid payload", func(t *testing.T) {
		delivered := stub(t)
		NewLogin(context.Background(), []byte("{"), valid)

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})

	t.Run("invalid address", func(t *testing.T) {
		delivered := stub(t)
		NewLogin(context.Background(), payload, invalid)

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})
}
