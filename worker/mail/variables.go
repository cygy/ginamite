package mail

import "net/smtp"

var (
	mailAuth       smtp.Auth
	smtpHost       string
	commonHTMLVars map[string]map[string]any // i.e. commonHTMLVars["StaticHost"]["fr-fr"] => "http://..."
	commonTextVars map[string]map[string]any // i.e. commonTextVars["StaticHost"]["fr-fr"] => "http://..."
)
