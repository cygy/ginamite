package account

import (
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"

	"github.com/gin-gonic/gin"
)

// CheckUsernameParameter : checks the username parameter.
func CheckUsernameParameter(c *gin.Context, username, parameterName string) bool {
	locale := context.GetLocale(c)

	// Username is mandatory.
	if len(username) == 0 {
		message := localization.Translate(locale)("error.username.not_found.message")
		reason := localization.Translate(locale)("error.username.not_found.reason")
		recovery := localization.Translate(locale)("error.username.not_found.recovery")

		response.NotFoundParameterValue(c, parameterName, message, reason, recovery)
		return false
	}

	// Username must contain at least X characters.
	if len(username) < authentication.MinimumCharactersForUsername {
		message := localization.Translate(locale)("error.username.invalid.message")
		reason := localization.Translate(locale)("error.username.minimum_characters.reason", localization.H{"Count": authentication.MinimumCharactersForUsername})
		recovery := localization.Translate(locale)("error.username.minimum_characters.recovery", localization.H{"Count": authentication.MinimumCharactersForUsername})

		response.InvalidParameterValue(c, parameterName, message, reason, recovery)
		return false
	}

	// Username must not contain invalid characters.
	if strings.ContainsAny(username, authentication.InvalidCharactersSetForUsername) {
		message := localization.Translate(locale)("error.username.invalid.message")
		reason := localization.Translate(locale)("error.username.invalid_characters.reason")
		recovery := localization.Translate(locale)("error.username.invalid_characters.recovery", localization.H{"Set": authentication.InvalidCharactersSetForUsername})

		response.InvalidParameterValue(c, parameterName, message, reason, recovery)
		return false
	}

	// Username must not contain invalid strings.
	for _, invalidString := range config.Main.Account.InvalidStringsForUsername {
		if strings.Contains(strings.ToLower(username), strings.ToLower(invalidString)) {
			// Spoof the error, if it happens it's because of a spam bot.
			response.InternalServerError(c)
			return false
		}
	}

	return true
}
