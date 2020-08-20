package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// RegistrationWelcome : sends a mail to welcome the user after the registration.
func RegistrationWelcome(payload []byte) {
	p := queue.MailRegistrationWelcome{}
	queue.UnmarshalPayload(payload, &p)

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.registration.welcome.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.registration.welcome.title"),
			"Message": template.HTML(t("email.registration.welcome.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.registration.welcome.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.registration.welcome.title"),
			"Message": t("email.registration.welcome.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.registration.welcome.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// RegistrationConfirmation : sends a mail to confirm the registration.
func RegistrationConfirmation(payload []byte) {
	p := queue.MailRegistrationConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountRegistrationConfirm, p.Locale), p.UserID, p.RegistrationCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountRegistrationCancel, p.Locale), p.UserID, p.RegistrationCode)

	msg := Message{
		Subject:   t("email.registration.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.registration.confirm.title"),
			"Message": template.HTML(t("email.registration.confirm.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.registration.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.registration.confirm.html.button.cancel"),
			"Footer":        t("email.registration.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.registration.confirm.title"),
			"Message": t("email.registration.confirm.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
				"CancelURL":  cancelURL,
			}),
			"Footer": t("email.registration.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// RegistrationConfirmed : sends a mail to notify the registration is confirmed.
func RegistrationConfirmed(payload []byte) {
	p := queue.MailRegistrationConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.registration.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.registration.confirmed.title"),
			"Message": template.HTML(t("email.registration.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.registration.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.registration.confirmed.title"),
			"Message": t("email.registration.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.registration.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// RegistrationCancelled : sends a mail to notify the registration is cancelled.
func RegistrationCancelled(payload []byte) {
	p := queue.MailRegistrationCancelled{}
	queue.UnmarshalPayload(payload, &p)

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.registration.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.registration.cancelled.title"),
			"Message": template.HTML(t("email.registration.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.registration.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.registration.cancelled.title"),
			"Message": t("email.registration.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.registration.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
