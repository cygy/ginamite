// Package mail builds and sends the emails of the application.
package mail

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/template/html"
	"github.com/cygy/ginamite/common/template/text"
	"github.com/cygy/ginamite/common/weburl"

	"github.com/domodwyer/mailyak/v3"
	"github.com/sirupsen/logrus"
)

// The templates of the emails, each one having a HTML (.html) and a text (.txt) version.
const (
	// TemplateAlert is the template for important notifications with a cancel button.
	TemplateAlert = "alert"

	// TemplateConfirmation is the template for notifications with a confirmation button.
	TemplateConfirmation = "confirmation"

	// TemplateConfirmationAndCancel is the template for notifications with a confirmation and a cancel buttons.
	TemplateConfirmationAndCancel = "confirmation_and_cancel"

	// TemplateNotification is the template for simple notifications.
	TemplateNotification = "notification"
)

// Message is the details of an email.
type Message struct {
	Subject   string
	Recipient string
	Template  string
	HTMLVars  map[string]any
	TextVars  map[string]any
	Locale    string
}

// IsEmailAddressValidFunc reports whether an email can be sent to an email address.
type IsEmailAddressValidFunc func(ctx context.Context, emailAddress string) (bool, error)

// check reports whether an email can be sent to the email address.
// It is false, and logged, if the function is undefined or fails.
func (f IsEmailAddressValidFunc) check(ctx context.Context, emailAddress string) bool {
	if f == nil {
		log.Error("unable to check the email address: the function is undefined")
		return false
	}

	isValid, err := f(ctx, emailAddress)
	if err != nil {
		log.WithField("error", err.Error()).Error("unable to check the email address")
	}

	return isValid && err == nil
}

var (
	mailAuth       smtp.Auth
	smtpHost       string
	addresses      Sender
	commonHTMLVars map[string]map[string]any // i.e. commonHTMLVars["StaticHost"]["fr-fr"] => "http://..."
	commonTextVars map[string]map[string]any // i.e. commonTextVars["StaticHost"]["fr-fr"] => "http://..."
)

// Replaced by the tests.
var (
	translate = localization.Translate
	routeURL  = weburl.GetURL
	deliver   = Message.Send
)

// Sender holds the addresses of the emails sent.
type Sender struct {
	EmailAddress        string
	Name                string
	NoReplyEmailAddress string
}

// Initialize initializes the mail session. The variables of the templates are indexed by name, then by locale.
func Initialize(identity, username, password, host string, port int, sender Sender, HTMLVars, textVars map[string]map[string]any) {
	addresses = sender
	mailAuth = smtp.PlainAuth(identity, username, password, host)
	smtpHost = fmt.Sprintf("%s:%d", host, port)
	commonHTMLVars = HTMLVars
	commonTextVars = textVars
}

// mergeVars returns the common variables of the locale completed, and overridden, by the variables of the message.
func mergeVars(common map[string]map[string]any, locale string, vars map[string]any) map[string]any {
	merged := make(map[string]any, len(common)+len(vars))
	for name, byLocale := range common {
		merged[name] = byLocale[locale]
	}
	for name, value := range vars {
		merged[name] = value
	}

	return merged
}

// Send sends the HTML and text versions of the email to the recipient.
// Nothing is sent if a template can not be executed. The errors are logged.
func (m Message) Send() {
	fullHTMLVars := mergeVars(commonHTMLVars, m.Locale, m.HTMLVars)
	fullTextVars := mergeVars(commonTextVars, m.Locale, m.TextVars)

	message := mailyak.New(smtpHost, mailAuth)
	message.Subject(m.Subject)
	message.From(addresses.EmailAddress)
	message.FromName(addresses.Name)
	message.To(m.Recipient)
	message.ReplyTo(addresses.NoReplyEmailAddress)

	// The variables are not logged: they contain the keys of the confirmation links.
	if err := html.Main.ExecuteTemplate(message.HTML(), m.Template+".html", fullHTMLVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": m.Template,
			"error":    err.Error(),
		}).Error("unable to execute the template")
		return
	}

	if err := text.Main.ExecuteTemplate(message.Plain(), m.Template+".txt", fullTextVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": m.Template,
			"error":    err.Error(),
		}).Error("unable to execute the template")
		return
	}

	if err := message.Send(); err != nil {
		log.WithFields(logrus.Fields{
			"SMTP":  smtpHost,
			"error": err.Error(),
		}).Error("unable to send mail")
	}
}
