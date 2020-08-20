package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// NewLogin : sends a mail to notify a user that a login occurs on its account.
func NewLogin(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailNewLogin{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	cancelLoginURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountLoginCancel, p.Locale), p.TokenID, p.TokenKey)
	unsubscribeURL := fmt.Sprintf("%s?key=%s", weburl.GetURL(weburl.AccountUnsubscribe, p.Locale), p.UnsubscribeKey)

	msg := Message{
		Subject:   t("email.login.new.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateAlert,
		HTMLVars: localization.H{
			"Title": t("email.login.new.title"),
			"Message": template.HTML(t("email.login.new.html.message", localization.H{
				"IP":     p.IPAddress,
				"Source": p.Source,
				"Device": p.Device,
			})),
			"CancelURL":    cancelLoginURL,
			"CancelButton": t("email.login.new.html.button.cancel"),
			"Footer": template.HTML(t("email.login.new.html.footer", localization.H{
				"UnsubscribeURL": unsubscribeURL,
			})),
		},
		TextVars: localization.H{
			"Title": t("email.login.new.title"),
			"Message": t("email.login.new.text.message", localization.H{
				"IP":        p.IPAddress,
				"Source":    p.Source,
				"Device":    p.Device,
				"CancelURL": cancelLoginURL,
			}),
			"Footer": t("email.login.new.text.footer", localization.H{
				"UnsubscribeURL": unsubscribeURL,
			}),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
