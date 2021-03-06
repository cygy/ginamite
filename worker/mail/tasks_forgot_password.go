package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// ForgotPasswordConfirmation : sends a mail to confirm the forgot password.
func ForgotPasswordConfirmation(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailForgotPasswordConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountForgotPasswordConfirm, p.Locale), p.ProcessID, p.ProcessCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountForgotPasswordCancel, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.forgot_password.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.forgot_password.confirm.title"),
			"Message": template.HTML(t("email.forgot_password.confirm.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.forgot_password.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.forgot_password.confirm.html.button.cancel"),
			"Footer":        t("email.forgot_password.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.forgot_password.confirm.title"),
			"Message": t("email.forgot_password.confirm.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
				"CancelURL":  cancelURL,
			}),
			"Footer": t("email.forgot_password.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// ForgotPasswordConfirmed : sends a mail to notify the forgot password reset is confirmed.
func ForgotPasswordConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailForgotPasswordConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.forgot_password.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.forgot_password.confirmed.title"),
			"Message": template.HTML(t("email.forgot_password.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.forgot_password.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.forgot_password.confirmed.title"),
			"Message": t("email.forgot_password.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.forgot_password.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// ForgotPasswordCancelled : sends a mail to notify the forgot password reset is cancelled.
func ForgotPasswordCancelled(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailForgotPasswordCancelled{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.forgot_password.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.forgot_password.cancelled.title"),
			"Message": template.HTML(t("email.forgot_password.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.forgot_password.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.forgot_password.cancelled.title"),
			"Message": t("email.forgot_password.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.forgot_password.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
