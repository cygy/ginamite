package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// DisableAccountConfirmation : sends a mail to confirm the account disable.
func DisableAccountConfirmation(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDisableAccountConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountDisableConfirm, p.Locale), p.ProcessID, p.ProcessCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountDisableCancel, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.disable_account.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.disable_account.confirm.title"),
			"Message": template.HTML(t("email.disable_account.confirm.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.disable_account.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.disable_account.confirm.html.button.cancel"),
			"Footer":        t("email.disable_account.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.disable_account.confirm.title"),
			"Message": t("email.disable_account.confirm.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
				"CancelURL":  cancelURL,
			}),
			"Footer": t("email.disable_account.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// DisableAccountConfirmed : sends a mail to notify the account disable is confirmed.
func DisableAccountConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDisableAccountConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.disable_account.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.disable_account.confirmed.title"),
			"Message": template.HTML(t("email.disable_account.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.disable_account.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.disable_account.confirmed.title"),
			"Message": t("email.disable_account.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.disable_account.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// DisableAccountCancelled : sends a mail to notify the account disable is cancelled.
func DisableAccountCancelled(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDisableAccountCancelled{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.disable_account.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.disable_account.cancelled.title"),
			"Message": template.HTML(t("email.disable_account.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.disable_account.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.disable_account.cancelled.title"),
			"Message": t("email.disable_account.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.disable_account.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// EnableAccountConfirmed : sends a mail to notify the account enable is confirmed.
func EnableAccountConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailEnableAccountConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.enable_account.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.enable_account.confirmed.title"),
			"Message": template.HTML(t("email.enable_account.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.enable_account.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.enable_account.confirmed.title"),
			"Message": t("email.enable_account.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.enable_account.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
