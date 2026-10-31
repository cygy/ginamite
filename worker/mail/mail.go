// Package mail builds and sends the emails of the application.
package mail

import (
	"context"
	"fmt"
	htmltemplate "html/template"
	"net/smtp"
	texttemplate "text/template"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/log"
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

// Sender holds the addresses of the emails sent.
type Sender struct {
	EmailAddress        string
	Name                string
	NoReplyEmailAddress string
}

// Settings are the properties of a Mailer.
type Settings struct {
	Identity string
	Username string
	Password string
	Host     string // The host of the SMTP server.
	Port     int    // The port of the SMTP server.
	Sender   Sender

	// HTMLVars and TextVars are the variables available in all the emails, indexed by name, then by locale.
	HTMLVars map[string]map[string]any
	TextVars map[string]map[string]any

	// Localization translates the emails.
	Localization *localization.Engine

	// Routes gives the URLs of the pages of the website linked by the emails.
	Routes *weburl.Configuration

	// HTMLTemplates and TextTemplates are the templates of the emails.
	HTMLTemplates *htmltemplate.Template
	TextTemplates *texttemplate.Template
}

// Mailer builds and sends the emails. Create it with New.
type Mailer struct {
	auth     smtp.Auth
	smtpHost string
	settings Settings

	// translate, routeURL and deliver are replaced by the tests.
	translate func(locale string) localization.TranslateFunc // Returns nil if the locale is unknown.
	routeURL  func(route, locale string) string
	deliver   func(Message) error
}

// New returns a Mailer sending the emails with the settings.
func New(settings Settings) *Mailer {
	m := &Mailer{
		auth:     smtp.PlainAuth(settings.Identity, settings.Username, settings.Password, settings.Host),
		smtpHost: fmt.Sprintf("%s:%d", settings.Host, settings.Port),
		settings: settings,
	}
	m.translate = settings.Localization.Translate
	m.routeURL = settings.Routes.GetURL
	m.deliver = m.Send

	return m
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

// Send sends the HTML and text versions of the message to its recipient.
// Nothing is sent if a template can not be executed. The errors are logged and returned: the one of a template is
// marked as permanent, see broker.Permanent, since sending the message again can not fix it.
func (m *Mailer) Send(msg Message) error {
	fullHTMLVars := mergeVars(m.settings.HTMLVars, msg.Locale, msg.HTMLVars)
	fullTextVars := mergeVars(m.settings.TextVars, msg.Locale, msg.TextVars)

	message := mailyak.New(m.smtpHost, m.auth)
	message.Subject(msg.Subject)
	message.From(m.settings.Sender.EmailAddress)
	message.FromName(m.settings.Sender.Name)
	message.To(msg.Recipient)
	message.ReplyTo(m.settings.Sender.NoReplyEmailAddress)

	// The variables are not logged: they contain the keys of the confirmation links.
	if err := m.settings.HTMLTemplates.ExecuteTemplate(message.HTML(), msg.Template+".html", fullHTMLVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": msg.Template,
			"error":    err.Error(),
		}).Error("unable to execute the template")

		return broker.Permanent(fmt.Errorf("unable to execute the template %s: %w", msg.Template, err))
	}

	if err := m.settings.TextTemplates.ExecuteTemplate(message.Plain(), msg.Template+".txt", fullTextVars); err != nil {
		log.WithFields(logrus.Fields{
			"template": msg.Template,
			"error":    err.Error(),
		}).Error("unable to execute the template")

		return broker.Permanent(fmt.Errorf("unable to execute the template %s: %w", msg.Template, err))
	}

	if err := message.Send(); err != nil {
		log.WithFields(logrus.Fields{
			"SMTP":  m.smtpHost,
			"error": err.Error(),
		}).Error("unable to send mail")

		return fmt.Errorf("unable to send mail: %w", err)
	}

	return nil
}
