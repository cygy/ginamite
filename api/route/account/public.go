package account

import (
	"regexp"
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	registrationError "github.com/cygy/ginamite/common/errors/registration"
	"github.com/cygy/ginamite/common/localization"

	res "github.com/cygy/ginamite/common/response"
	"github.com/gin-gonic/gin"
)

// CheckUsernameParameter : checks the username parameter.
func CheckUsernameParameter(c *gin.Context, username, parameterName string) bool {
	locale := context.GetLocale(c)

	usernameLength := len(username)

	// Username is mandatory.
	if usernameLength == 0 {
		message := localization.Translate(locale)("error.username.not_found.message")
		reason := localization.Translate(locale)("error.username.not_found.reason")
		recovery := localization.Translate(locale)("error.username.not_found.recovery")

		response.NotFoundParameterValue(c, parameterName, message, reason, recovery)
		return false
	}

	// Username must contain at least X characters.
	if usernameLength < authentication.MinimumCharactersForUsername {
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
	// Do not accept some patterns because of spam.
	if doesUsernameContainForbiddenStrings(username) || !isUsernamePatternAccepted(username) {
		// Spoof the error, if it happens it's because of a spam bot.
		t := localization.Translate(locale)
		response.PreconditionFailedWithError(c, res.Error{
			Domain:   registrationError.Domain,
			Code:     registrationError.AccountAlreadyExists,
			Message:  t("error.registration.message"),
			Reason:   t("error.registration.account.already_exists.reason"),
			Recovery: t("error.registration.account.already_exists.recovery"),
		})
		return false
	}

	return true
}

func doesUsernameContainForbiddenStrings(username string) bool {
	loweredUsername := strings.ToLower(username)
	for _, invalidString := range config.Main.Account.InvalidStringsForUsername {
		if strings.Contains(loweredUsername, strings.ToLower(invalidString)) {
			return true
		}
	}
	return false
}

func isUsernamePatternAccepted(username string) bool {
	usernameLength := len(username)
	if usernameLength < 8 {
		return true
	}

	// sequence of 4 or more consonants
	rejectedRe := regexp.MustCompile(`[bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXYZ]{4,}`)
	if rejectedRe.MatchString(username) {
		return false
	}

	// sequencee of uppeer and lower case characters with not enough vowels
	rejectedRe = regexp.MustCompile(`^([a-z]+[A-Z]+|[A-Z]+[a-z]+){1,}([a-z]*|[A-Z]*)?$`)
	if rejectedRe.MatchString(username) {
		acceptedRe1 := regexp.MustCompile(`^([A-Z][a-z]+){1,2}([a-z]*|[A-Z]*)?$`)
		acceptedRe2 := regexp.MustCompile(`^[a-z]+([A-Z][a-z]+)*$`)
		if acceptedRe1.MatchString(username) || acceptedRe2.MatchString(username) {
			loweredUsername := strings.ToLower(username)
			vowels := len(regexp.MustCompile(`[aeiouy]`).FindAllString(loweredUsername, -1))
			return vowels > usernameLength/6
		}
		return false
	}

	return true
}

/*
package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	accepted := []string{
		"AvianAnalyst",
		"MartinEdush",
		"JacquesVettE",
		"lllolllo",
		"Sasakura",
		"byakkun",
		"theoLord",
		"ScottLat",
		"JesterJunk",
		"cloudcolors",
		"Adelicya",
		"Cyril",
		"FredLand",
		"Kikine",
		"keitaro",
	}

	rejected := []string{
		"tytHRBCWGz",
		"CbIVUrts",
		"bTrAPIQFaIyQv",
		"aceXlhyYuhB",
		"zhZmjyKf",
		"rNophjAAYb",
		"SArFyugvbqr",
		"XnTfybBSz",
		"jIvomLJgupPz",
		"XbJGmVzrThVXHb",
		"siJpvkKLglGS",
		"jcmEFaOJEhe",
		"YHXhzwBi",
		"PtZKbPNxNemWG",
		"wCOdjEca",
		"PBlLTMRdUbyB",
		"uCQsKZmariK",
		"YHXhzwBi",
		"CzNifldgDoAVfr",
		"yoWhCoGRrvN",
		"yvJtLwhuCbjTCF",
		"VfPtBKFGDbsnn",
		"VtRrVKSRJAqzBW",
		"iOTQAbUvI",
		"OMelGMbXbIVG",
		"gBwdmOQsKMt",
		"nTdJaEfM",
		"VTBVrxtWkNyN",
		"oVBcmYYhDu",
		"XwJwFJEghxaIqG",
		"XYgIhkUrEK",
		"jDiZxjGJotF",
		"qgEycFijJWyhr",
		"wgDMraTOiDhr",
		"iijoyUJaUYe",
		"zTIPaOuazVAWAq",
		"AWJLtbLPixl",
		"eCUeyEtIe",
		"qnRlmccBb",
		"GaLyQvHhIrYbnKk",
		"FhWbPCRwNYC",
	}

	fmt.Println("Accepted strings (validation):")
	for _, str := range accepted {
		if IsAccepted(str) {
			fmt.Println(str, "=> Accepted")
		} else {
			fmt.Println(str, "=> Rejected (should be accepted)")
		}
	}

	fmt.Println("\nRejected strings (validation):")
	for _, str := range rejected {
		if !IsAccepted(str) {
			fmt.Println(str, "=> Rejected")
		} else {
			fmt.Println(str, "=> Accepted (should be rejected)")
		}
	}
}
*/
