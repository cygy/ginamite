package mail

import (
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/weburl"
)

// UpdateEmailAddressConfirmation : sends a mail to confirm the email address update.
func UpdateEmailAddressConfirmation(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdateEmailAddressConfirmation{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountUpdateEmailAddressConfirm, p.Locale), p.ProcessID, p.ProcessCode)
	cancelURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountUpdateEmailAddressCancel, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.update_email_address.confirm.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmationAndCancel,
		HTMLVars: localization.H{
			"Title": t("email.update_email_address.confirm.title"),
			"Message": template.HTML(t("email.update_email_address.confirm.html.message", localization.H{
				"Username":        p.Username,
				"NewEmailAddress": p.Data,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.update_email_address.confirm.html.button.confirm"),
			"CancelURL":     cancelURL,
			"CancelButton":  t("email.update_email_address.confirm.html.button.cancel"),
			"Footer":        t("email.update_email_address.confirm.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_email_address.confirm.title"),
			"Message": t("email.update_email_address.confirm.text.message", localization.H{
				"Username":        p.Username,
				"NewEmailAddress": p.Data,
				"ConfirmURL":      confirmURL,
				"CancelURL":       cancelURL,
			}),
			"Footer": t("email.update_email_address.confirm.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// UpdateEmailAddressConfirmed : sends a mail to notify the email address update is confirmed.
func UpdateEmailAddressConfirmed(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdateEmailAddressConfirmed{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.update_email_address.confirmed.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.update_email_address.confirmed.title"),
			"Message": template.HTML(t("email.update_email_address.confirmed.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.update_email_address.confirmed.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_email_address.confirmed.title"),
			"Message": t("email.update_email_address.confirmed.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.update_email_address.confirmed.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// UpdateEmailAddressCancelled : sends a mail to notify the email address update is cancelled.
func UpdateEmailAddressCancelled(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailUpdateEmailAddressCancelled{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	msg := Message{
		Subject:   t("email.update_email_address.cancelled.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateNotification,
		HTMLVars: localization.H{
			"Title": t("email.update_email_address.cancelled.title"),
			"Message": template.HTML(t("email.update_email_address.cancelled.html.message", localization.H{
				"Username": p.Username,
			})),
			"Footer": t("email.update_email_address.cancelled.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.update_email_address.cancelled.title"),
			"Message": t("email.update_email_address.cancelled.text.message", localization.H{
				"Username": p.Username,
			}),
			"Footer": t("email.update_email_address.cancelled.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}

// ValidateNewEmailAddress : sends a mail to validate the new email address.
func ValidateNewEmailAddress(payload []byte, isEmailAddressValid IsEmailAddressValidFunc) {
	p := queue.MailValidateNewEmailAddress{}
	queue.UnmarshalPayload(payload, &p)

	if !isEmailAddressValid(p.EmailAddress) {
		return
	}

	t := localization.Translate(p.Locale)

	confirmURL := fmt.Sprintf("%s?id=%s&key=%s", weburl.GetURL(weburl.AccountValidateNewEmailAddress, p.Locale), p.ProcessID, p.ProcessCode)

	msg := Message{
		Subject:   t("email.validate_new_email_address.subject", localization.H{"Username": p.Username}),
		Recipient: p.EmailAddress,
		Template:  TemplateConfirmation,
		HTMLVars: localization.H{
			"Title": t("email.validate_new_email_address.title"),
			"Message": template.HTML(t("email.validate_new_email_address.html.message", localization.H{
				"Username": p.Username,
			})),
			"ConfirmURL":    confirmURL,
			"ConfirmButton": t("email.validate_new_email_address.html.button.confirm"),
			"Footer":        t("email.validate_new_email_address.html.footer"),
		},
		TextVars: localization.H{
			"Title": t("email.validate_new_email_address.title"),
			"Message": t("email.validate_new_email_address.text.message", localization.H{
				"Username":   p.Username,
				"ConfirmURL": confirmURL,
			}),
			"Footer": t("email.validate_new_email_address.text.footer"),
		},
		Locale: p.Locale,
	}
	msg.Send()
}
