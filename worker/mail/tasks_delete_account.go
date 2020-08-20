package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// DeleteAccountConfirmation : sends a mail to confirm the account deletion.
func DeleteAccountConfirmation(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDeleteAccountConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountDeleteConfirm, p.Locale), p.ProcessID, p.ProcessCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountDeleteCancel, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.delete_account.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.delete_account.confirm.title"),
			"Message": template.HTML(t("email.delete_account.confirm.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.delete_account.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.delete_account.confirm.html.button.cancel"),
			"Footer":        t("email.delete_account.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.delete_account.confirm.title"),
			"Message": t("email.delete_account.confirm.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
				"CancelURL":  cancelURL,
			}),
			"Footer": t("email.delete_account.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// DeleteAccountConfirmed : sends a mail to notify the account deletion is confirmed.
func DeleteAccountConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDeleteAccountConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.delete_account.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.delete_account.confirmed.title"),
			"Message": template.HTML(t("email.delete_account.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.delete_account.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.delete_account.confirmed.title"),
			"Message": t("email.delete_account.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.delete_account.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// DeleteAccountCancelled : sends a mail to notify the account deletion is cancelled.
func DeleteAccountCancelled(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailDeleteAccountCancelled{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.delete_account.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.delete_account.cancelled.title"),
			"Message": template.HTML(t("email.delete_account.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.delete_account.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.delete_account.cancelled.title"),
			"Message": t("email.delete_account.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.delete_account.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
