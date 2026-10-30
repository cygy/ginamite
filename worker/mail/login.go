package mail

import (
	"context"
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// NewLogin sends an email to notify a user that a login occurs on their account.
// The email has a link to cancel the login and a link to unsubscribe from these emails.
func NewLogin(ctx context.Context, payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailNewLogin{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return
	}

	if !isEmailAddressValid.check(ctx, p.EmailAddress) {
		return
	}

	const key = "email.login.new"

	t := translate(p.Locale)
	unsubscribeURL := fmt.Sprintf("%s?key=%s", routeURL(weburl.AccountUnsubscribe, p.Locale), p.UnsubscribeKey)
	footerVars := localization.H{"UnsubscribeURL": unsubscribeURL}

	msg := content{
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
		cancelURL: processURL(weburl.AccountLoginCancel, p.Locale, p.TokenID, p.TokenKey),
	}.message()

	msg.HTMLVars["Footer"] = template.HTML(t(key+".html.footer", footerVars))
	msg.TextVars["Footer"] = t(key+".text.footer", footerVars)

	deliver(msg)
}
