package mail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.com/cygy/ginamite/common/localization"
)

const (
	address = "john@example.com"
	newAddr = "new@example.com"
)

// valid, invalid and failing are checks of the email addresses.
func valid(context.Context, string) (bool, error)   { return true, nil }
func invalid(context.Context, string) (bool, error) { return false, nil }
func failing(context.Context, string) (bool, error) { return true, errors.New("boom") }

// sender is a function sending a mail from the payload of a message.
type sender func(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc)

// stub replaces the translations, the URLs and the delivery of the mails, for the duration of the test,
// and returns the delivered messages. A translation is its ID followed by its variables.
func stub(t *testing.T) *[]Message {
	t.Helper()

	oldTranslate, oldRouteURL, oldDeliver := translate, routeURL, deliver
	t.Cleanup(func() { translate, routeURL, deliver = oldTranslate, oldRouteURL, oldDeliver })

	translate = func(string) localization.TranslateFunc {
		return func(id string, args ...any) string {
			if len(args) == 0 {
				return id
			}

			return fmt.Sprintf("%s %v", id, args[0])
		}
	}
	routeURL = func(route, locale string) string { return "https://web/" + locale + "/" + route }

	delivered := &[]Message{}
	deliver = func(m Message) { *delivered = append(*delivered, m) }

	return delivered
}

// encode returns the JSON of a payload.
func encode(t *testing.T, v any) []byte {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

// notificationPayload returns the payload of a notification.
func notificationPayload(t *testing.T) []byte {
	t.Helper()

	return encode(t, map[string]string{"email": address, "username": "john", "locale": "en"})
}

// processPayload returns the payload of a process confirmation.
func processPayload(t *testing.T) []byte {
	t.Helper()

	return encode(t, map[string]string{
		"email": address, "username": "john", "locale": "en", "data": newAddr, "process_id": "pid", "process_code": "pcode",
	})
}

// contains fails the test if the HTML or the text s does not contain all the parts.
func contains(t *testing.T, name string, s any, parts ...string) {
	t.Helper()

	got := fmt.Sprint(s)
	for _, part := range parts {
		if !strings.Contains(got, part) {
			t.Errorf("%s = %q, want it to contain %q", name, got, part)
		}
	}
}

// checkCommon checks the fields of a message built from the translations "email.<key>.*".
func checkCommon(t *testing.T, m Message, key, tmpl string) {
	t.Helper()

	prefix := "email." + key
	if !strings.HasPrefix(m.Subject, prefix+".subject") || !strings.Contains(m.Subject, "Username:john") {
		t.Errorf("Subject = %q", m.Subject)
	}
	if m.Recipient != address || m.Template != tmpl || m.Locale != "en" {
		t.Errorf("message = %+v", m)
	}

	if m.HTMLVars["Title"] != prefix+".title" || m.TextVars["Title"] != prefix+".title" {
		t.Errorf("titles = %v, %v", m.HTMLVars["Title"], m.TextVars["Title"])
	}
	contains(t, "HTML footer", m.HTMLVars["Footer"], prefix+".html.footer")
	contains(t, "text footer", m.TextVars["Footer"], prefix+".text.footer")

	htmlMessage, ok := m.HTMLVars["Message"].(template.HTML)
	if !ok {
		t.Fatalf("the HTML message is a %T, want template.HTML", m.HTMLVars["Message"])
	}
	contains(t, "HTML message", htmlMessage, prefix+".html.message", "Username:john")
	contains(t, "text message", m.TextVars["Message"], prefix+".text.message", "Username:john")
}

// checkNotification checks that send delivers the simple notification "email.<key>.*".
// Nothing must be delivered for an invalid payload or an invalid email address, unless checked is false.
func checkNotification(t *testing.T, send sender, key string, checked bool) {
	t.Helper()

	ctx := context.Background()

	t.Run("sent", func(t *testing.T) {
		delivered := stub(t)
		send(ctx, notificationPayload(t), valid)

		if len(*delivered) != 1 {
			t.Fatalf("delivered %d messages, want 1", len(*delivered))
		}
		checkCommon(t, (*delivered)[0], key, TemplateNotification)
	})

	t.Run("invalid payload", func(t *testing.T) {
		delivered := stub(t)
		send(ctx, []byte("{"), valid)

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})

	if !checked {
		return
	}

	for name, check := range map[string]IsEmailAddressValidFunc{"invalid address": invalid, "failing check": failing, "no check": nil} {
		t.Run(name, func(t *testing.T) {
			delivered := stub(t)
			send(ctx, notificationPayload(t), check)

			if len(*delivered) != 0 {
				t.Errorf("delivered %d messages, want none", len(*delivered))
			}
		})
	}
}

// checkProcess checks that send delivers the email "email.<key>.*" asking to confirm, and possibly to cancel,
// a process. Nothing must be delivered for an invalid payload or an invalid email address.
func checkProcess(t *testing.T, send sender, key, confirmRoute, cancelRoute string) {
	t.Helper()

	ctx := context.Background()

	t.Run("sent", func(t *testing.T) {
		delivered := stub(t)
		send(ctx, processPayload(t), valid)

		if len(*delivered) != 1 {
			t.Fatalf("delivered %d messages, want 1", len(*delivered))
		}
		m := (*delivered)[0]

		tmpl := TemplateConfirmation
		if cancelRoute != "" {
			tmpl = TemplateConfirmationAndCancel
		}
		checkCommon(t, m, key, tmpl)

		prefix := "email." + key
		confirmURL := "https://web/en/" + confirmRoute + "?id=pid&key=pcode"
		if m.HTMLVars["ConfirmURL"] != confirmURL || m.HTMLVars["ConfirmButton"] != prefix+".html.button.confirm" {
			t.Errorf("confirmation = %v, %v", m.HTMLVars["ConfirmURL"], m.HTMLVars["ConfirmButton"])
		}
		contains(t, "HTML message", m.HTMLVars["Message"], "NewEmailAddress:"+newAddr)
		contains(t, "text message", m.TextVars["Message"], "ConfirmURL:"+confirmURL, "NewEmailAddress:"+newAddr)

		if cancelRoute == "" {
			if _, ok := m.HTMLVars["CancelURL"]; ok {
				t.Error("the email must not have a cancel link")
			}

			return
		}

		cancelURL := "https://web/en/" + cancelRoute + "?id=pid&key=pcode"
		if m.HTMLVars["CancelURL"] != cancelURL || m.HTMLVars["CancelButton"] != prefix+".html.button.cancel" {
			t.Errorf("cancellation = %v, %v", m.HTMLVars["CancelURL"], m.HTMLVars["CancelButton"])
		}
		contains(t, "text message", m.TextVars["Message"], "CancelURL:"+cancelURL)
	})

	t.Run("invalid payload", func(t *testing.T) {
		delivered := stub(t)
		send(ctx, []byte("{"), valid)

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})

	t.Run("invalid address", func(t *testing.T) {
		delivered := stub(t)
		send(ctx, processPayload(t), invalid)

		if len(*delivered) != 0 {
			t.Errorf("delivered %d messages, want none", len(*delivered))
		}
	})
}
