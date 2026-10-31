package mail

import (
	"context"
	"fmt"
	"html/template"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
)

// content describes an email built from the translations "email.<key>.*":
// subject, title, html.message, html.footer, text.message and text.footer, plus
// html.button.confirm and html.button.cancel if it has a confirmation or a cancellation link.
type content struct {
	key        string
	template   string
	recipient  string
	username   string
	locale     string
	vars       localization.H // The variables of the messages, besides the username and the links.
	confirmURL string
	cancelURL  string
}

// message returns the email described by the content.
func (m *Mailer) message(c content) Message {
	t := m.translate(c.locale)

	messageVars := localization.H{"Username": c.username}
	for name, value := range c.vars {
		messageVars[name] = value
	}

	htmlVars := localization.H{
		"Title":   t("email." + c.key + ".title"),
		"Message": template.HTML(t("email."+c.key+".html.message", messageVars)),
		"Footer":  t("email." + c.key + ".html.footer"),
	}
	textVars := localization.H{
		"Title":  t("email." + c.key + ".title"),
		"Footer": t("email." + c.key + ".text.footer"),
	}

	textMessageVars := localization.H{}
	for name, value := range messageVars {
		textMessageVars[name] = value
	}

	if c.confirmURL != "" {
		htmlVars["ConfirmURL"] = c.confirmURL
		htmlVars["ConfirmButton"] = t("email." + c.key + ".html.button.confirm")
		textMessageVars["ConfirmURL"] = c.confirmURL
	}
	if c.cancelURL != "" {
		htmlVars["CancelURL"] = c.cancelURL
		htmlVars["CancelButton"] = t("email." + c.key + ".html.button.cancel")
		textMessageVars["CancelURL"] = c.cancelURL
	}

	textVars["Message"] = t("email."+c.key+".text.message", textMessageVars)

	return Message{
		Subject:   t("email."+c.key+".subject", localization.H{"Username": c.username}),
		Recipient: c.recipient,
		Template:  c.template,
		HTMLVars:  htmlVars,
		TextVars:  textVars,
		Locale:    c.locale,
	}
}

// processURL returns the URL of a route of a process, with the ID and the code of the process.
func (m *Mailer) processURL(route, locale, id, code string) string {
	return fmt.Sprintf("%s?id=%s&key=%s", m.routeURL(route, locale), id, code)
}

// alwaysValid is the check of the emails which are sent whatever the state of the email address.
func alwaysValid(context.Context, string) (bool, error) {
	return true, nil
}

// sendNotification decodes a notification payload and sends the email "email.<key>.*" to its recipient,
// if the email address is valid.
func (m *Mailer) sendNotification(ctx context.Context, payload []byte, key string, isEmailAddressValid IsEmailAddressValidFunc) error {
	p := queue.MailNotification{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	if !isEmailAddressValid.check(ctx, p.EmailAddress) {
		return nil
	}

	return m.deliver(m.message(content{
		key:       key,
		template:  TemplateNotification,
		recipient: p.EmailAddress,
		username:  p.Username,
		locale:    p.Locale,
	}))
}

// process describes the email asking a user to confirm, and possibly to cancel, a process.
type process struct {
	key     string
	confirm string // The route to confirm the process.
	cancel  string // The route to cancel the process, if the process can be cancelled.
}

// sendProcess decodes a process payload and sends the email asking the user to confirm the process,
// if the email address is valid. The data of the payload is available as NewEmailAddress.
func (m *Mailer) sendProcess(ctx context.Context, payload []byte, pr process, isEmailAddressValid IsEmailAddressValidFunc) error {
	p := queue.MailProcessConfirmation{}
	if err := queue.UnmarshalPayload(payload, &p); err != nil {
		return broker.Permanent(err)
	}

	if !isEmailAddressValid.check(ctx, p.EmailAddress) {
		return nil
	}

	c := content{
		key:        pr.key,
		template:   TemplateConfirmation,
		recipient:  p.EmailAddress,
		username:   p.Username,
		locale:     p.Locale,
		vars:       localization.H{"NewEmailAddress": p.Data},
		confirmURL: m.processURL(pr.confirm, p.Locale, p.ProcessID, p.ProcessCode),
	}
	if pr.cancel != "" {
		c.template = TemplateConfirmationAndCancel
		c.cancelURL = m.processURL(pr.cancel, p.Locale, p.ProcessID, p.ProcessCode)
	}

	return m.deliver(m.message(c))
}
