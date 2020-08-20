package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// UpdatePasswordConfirmation : sends a mail to confirm the password update.
func UpdatePasswordConfirmation(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdatePasswordConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountUpdatePasswordConfirm, p.Locale), p.ProcessID, p.ProcessCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountUpdatePasswordCancel, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.update_password.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.update_password.confirm.title"),
			"Message": template.HTML(t("email.update_password.confirm.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.update_password.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.update_password.confirm.html.button.cancel"),
			"Footer":        t("email.update_password.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_password.confirm.title"),
			"Message": t("email.update_password.confirm.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
				"CancelURL":  cancelURL,
			}),
			"Footer": t("email.update_password.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// UpdatePasswordConfirmed : sends a mail to notify the password update is confirmed.
func UpdatePasswordConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdatePasswordConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.update_password.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.update_password.confirmed.title"),
			"Message": template.HTML(t("email.update_password.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.update_password.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_password.confirmed.title"),
			"Message": t("email.update_password.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.update_password.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// UpdatePasswordCancelled : sends a mail to notify the password update is cancelled.
func UpdatePasswordCancelled(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdatePasswordCancelled{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.update_password.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.update_password.cancelled.title"),
			"Message": template.HTML(t("email.update_password.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.update_password.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_password.cancelled.title"),
			"Message": t("email.update_password.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.update_password.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
