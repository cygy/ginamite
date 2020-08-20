package mail

import (
	"fmt"
	"net/smtp"

	"github.com/domodwyer/mailyak"
	"github.com/sirupsen/logrus"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
)

// Initialize initializes the mail session.
func Initialize(identity, username, password, host string, port int, HTMLVars, textVars map[string]map[string]interface{}) {
	mailAuth = smtp.PlainAuth(identity, username, password, host)
	smtpHost = fmt.Sprintf("%s:%d", host, port)
	commonHTMLVars = HTMLVars
	commonTextVars = textVars
}

// Send sends a HTML mail to a recipient.
func (m Message) Send() {
	// Merge the HTML vars with the common vars.
	fullHTMLVars := make(map[string]interface{})
	for k, v := range commonHTMLVars {
		fullHTMLVars[k] = v[m.Locale]
	}
	for k, v := range m.HTMLVars {
		fullHTMLVars[k] = v
	}

	// Merge the text vars with the common vars.
	fullTextVars := make(map[string]interface{})
	for k, v := range commonTextVars {
		fullTextVars[k] = v[m.Locale]
	}
	for k, v := range m.TextVars {
		fullTextVars[k] = v
	}

	// Create the mail.
	message := mailyak.New(smtpHost, mailAuth)
	message.Subject(m.Subject)
	message.From(config.Main.SenderEmailAddress)
	message.FromName(config.Main.SenderName)
	message.To(m.Recipient)
	message.ReplyTo(config.Main.NoReplyEmailAddress)

	// Parse the HTML template.
	if err := html.Main.ExecuteTemplate(message.HTML(), m.Template+".html", fullHTMLVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": m.Template,
			"vars":     fullHTMLVars,
			"error":    err.Error(),
		}).Error("unable to execute the template")
	}

	// Parse the text template.
	if err := text.Main.ExecuteTemplate(message.Plain(), m.Template+".txt", fullTextVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": m.Template,
			"vars":     fullTextVars,
			"error":    err.Error(),
		}).Error("unable to execute the template")
	}

	// Send the mail.
	if err := message.Send(); err != nil {
		log.WithFields(logrus.Fields{
			"SMTP":  smtpHost,
			"error": err.Error(),
		}).Error("unable to send mail")
	}
}
