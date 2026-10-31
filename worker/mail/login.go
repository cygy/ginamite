package mail

import (
	"context"
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// NewLogin sends an email to notify a user that a login occurs on their account.
// The email has a link to cancel the login and a link to unsubscribe from these emails.
func (m *Mailer) NewLogin(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) error {
	p := queue.MailNewLogin{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	if !isEmailAddressValid.check(ctx, p.EmailAddress) {
		return nil
	}

	const key = "email.login.new"

	t := m.translate(p.Locale)
	unsubscribeURL := fmt.Sprintf("%s?key=%s", m.routeURL(weburl.AccountUnsubscribe, p.Locale), p.UnsubscribeKey)
	footerVars := localization.H{"UnsubscribeURL": unsubscribeURL}

	msg := m.message(content{
		key:       "login.new",
		template:  TemplateAlert,
		recipient: p.EmailAddress,
		username:  p.Username,
		locale:    p.Locale,
		vars: localization.H{
			"IP":     p.IPAddress,
			"Source": p.Source,
			"Device": p.Device,
		},
		cancelURL: m.processURL(weburl.AccountLoginCancel, p.Locale, p.TokenID, p.TokenKey),
	})

	msg.HTMLVars["Footer"] = template.HTML(t(key+".html.footer", footerVars))
	msg.TextVars["Footer"] = t(key+".text.footer", footerVars)

	return m.deliver(msg)
}
