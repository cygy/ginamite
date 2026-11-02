package mail

import (
	"context"
	"html/template"
	"testing"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/weburl"
)

func TestNewLogin(t *testing.T) {
	payload := encode(t, map[string]string{
		"email": address, "username": "john", "locale": "en", "unsubscribe_key": "ukey",
		"ip": "1.2.3.4", "source": "web", "device": "phone", "token_id": "tid", "token_key": "tkey",
	})

	t.Run("sent", func(t *testing.T) {
		mailer, delivered := stub(t)
		if err := mailer.NewLogin(context.Background(), payload, valid); err != nil {
			t.Fatalf("error = %v", err)
		}

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

	t.Run("the values given by the user are escaped in HTML", func(t *testing.T) {
		mailer, delivered := stub(t)
		injected := encode(t, map[string]string{
			"email": address, "username": "<b>john</b>", "locale": "en", "unsubscribe_key": "ukey",
			"ip": "1.2.3.4", "source": "web", "device": `<a href="https://evil">phone</a>`, "token_id": "tid", "token_key": "tkey",
		})
		if err := mailer.NewLogin(context.Background(), injected, valid); err != nil {
			t.Fatalf("error = %v", err)
		}

		m := (*delivered)[0]
		contains(t, "HTML message", m.HTMLVars["Message"], "Username:&lt;b&gt;john&lt;/b&gt;", "Device:&lt;a href=&#34;https://evil&#34;&gt;phone&lt;/a&gt;")
		contains(t, "HTML footer", m.HTMLVars["Footer"], "?key=ukey")
		contains(t, "text message", m.TextVars["Message"], `Device:<a href="https://evil">phone</a>`)
	})

	t.Run("invalid payload", func(t *testing.T) {
		mailer, delivered := stub(t)
		if err := mailer.NewLogin(context.Background(), []byte("{"), valid); !broker.IsPermanent(err) {
			t.Errorf("error = %v, want a permanent error", err)
		}

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})

	t.Run("invalid address", func(t *testing.T) {
		mailer, delivered := stub(t)
		if err := mailer.NewLogin(context.Background(), payload, invalid); err != nil {
			t.Errorf("error = %v, an invalid address is not a failure", err)
		}

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})
}
