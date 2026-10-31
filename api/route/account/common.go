package account

import (
	"errors"
	"strings"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// newTokenValidator returns the validator of the tokens issued by a third-party authentication method: the one of the
// functions of the context if any, the validators of Facebook and Google else. It returns nil for another method.
func newTokenValidator(c *gin.Context, method string) validator.TokenValidator {
	if newValidator := FunctionsOf(c).NewTokenValidator; newValidator != nil {
		return newValidator(c, method)
	}

	switch method {
	case authentication.MethodFacebook:
		return validator.NewFacebookTokenValidator(ctxconfig.Get(c).Facebook.AppID, ctxconfig.Get(c).Facebook.APIVersion, ctxconfig.Get(c).Facebook.TimeOut, ctxconfig.Get(c).Facebook.RetryCount, ctxconfig.Get(c).IsDebugModeEnabled())
	case authentication.MethodGoogle:
		return validator.NewGoogleTokenValidator(ctxconfig.Get(c).Google.ClientID, ctxconfig.Get(c).Google.APIVersion, ctxconfig.Get(c).Google.TimeOut, ctxconfig.Get(c).Google.RetryCount, ctxconfig.Get(c).IsDebugModeEnabled())
	}

	return nil
}

func validateOrCancelUpdateProperty(c *gin.Context, property string, validate bool) {
	// Check the type of the process.
	isEmailAddressProcess := property == authentication.ProcessUpdateEmailAddress
	isPasswordProcess := property == authentication.ProcessUpdatePassword
	isValidateEmailAddressProcess := property == authentication.ProcessVerifyEmailAddress
	isDeleteAccountProcess := property == authentication.ProcessDeleteAccount
	isDisableAccountProcess := property == authentication.ProcessDisableAccount

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	// Here is a local function that returns an error. This route does only return a single error.
	sendError := func(err error) {
		if err != nil {
			c.Error(err)
		}

		var message, reason, recovery string
		if validate {
			if isEmailAddressProcess {
				message = t("error.update_email_address.validate.message")
			} else if isPasswordProcess {
				message = t("error.update_password.validate.message")
			} else if isValidateEmailAddressProcess {
				message = t("error.validate_email_address.message")
			} else if isDeleteAccountProcess {
				message = t("error.delete_account.validate.message")
			} else if isDisableAccountProcess {
				message = t("error.disable_account.validate.message")
			}
		} else {
			if isEmailAddressProcess {
				message = t("error.update_email_address.cancel.message")
			} else if isPasswordProcess {
				message = t("error.update_password.cancel.message")
			} else if isDeleteAccountProcess {
				message = t("error.delete_account.cancel.message")
			} else if isDisableAccountProcess {
				message = t("error.disable_account.cancel.message")
			}
		}

		if isEmailAddressProcess {
			reason = t("error.update_email_address.validate_or_cancel.reason")
			recovery = t("error.update_email_address.validate_or_cancel.recovery")
		} else if isPasswordProcess {
			reason = t("error.update_password.validate_or_cancel.reason")
			recovery = t("error.update_password.validate_or_cancel.recovery")
		} else if isValidateEmailAddressProcess {
			reason = t("error.validate_email_address.reason")
			recovery = t("error.validate_email_address.recovery")
		} else if isDeleteAccountProcess {
			reason = t("error.delete_account.validate_or_cancel.reason")
			recovery = t("error.delete_account.validate_or_cancel.recovery")
		} else if isDisableAccountProcess {
			reason = t("error.disable_account.validate_or_cancel.reason")
			recovery = t("error.disable_account.validate_or_cancel.recovery")
		}

		response.InvalidRequestParameterWithDetails(c, message, reason, recovery)
	}

	// Do not continue if the process if unknown!
	if !isEmailAddressProcess && !isPasswordProcess && !isValidateEmailAddressProcess && !isDeleteAccountProcess && !isDisableAccountProcess {
		sendError(errors.New("wrong process '" + property + "'"))
		return
	}

	// ID and key are mandatory.
	processID := c.Param("id")
	key := c.Param("key")
	if len(processID) == 0 || len(key) == 0 {
		sendError(errors.New("the process ID or the key is missing"))
		return
	}

	// Some functions must be verified.
	undefined := FunctionsOf(c).GetUserDetailsByID == nil || FunctionsOf(c).VerifyProcess == nil || FunctionsOf(c).DeleteProcessByID == nil

	if validate {
		if isEmailAddressProcess {
			undefined = undefined || FunctionsOf(c).UpdateUserEmailAddress == nil
		} else if isPasswordProcess {
			undefined = undefined || FunctionsOf(c).UpdateUserPassword == nil
		} else if isValidateEmailAddressProcess {
			undefined = undefined || FunctionsOf(c).UserEmailAddressIsVerified == nil
		} else if isDeleteAccountProcess {
			undefined = undefined || FunctionsOf(c).DeleteUserByID == nil
		} else if isDisableAccountProcess {
			undefined = undefined || FunctionsOf(c).DisableUserByID == nil
		}
	}

	if undefined {
		sendError(errors.New("a function in the package 'ginamite/api/route/account' must be defined"))
		return
	}

	// Verify the process.
	userID, value, err := FunctionsOf(c).VerifyProcess(c, processID, key)
	if err != nil {
		sendError(err)
		return
	}

	user, err := FunctionsOf(c).GetUserDetailsByID(c, userID)
	if err != nil {
		sendError(err)
		return
	}

	// Now update the password/email address.
	if validate {
		if isEmailAddressProcess {
			if err := FunctionsOf(c).UpdateUserEmailAddress(c, userID, value); err != nil {
				sendError(err)
				return
			}
		} else if isPasswordProcess {
			if err := FunctionsOf(c).UpdateUserPassword(c, userID, value, true); err != nil {
				sendError(err)
				return
			}
		} else if isValidateEmailAddressProcess {
			if err := FunctionsOf(c).UserEmailAddressIsVerified(c, userID); err != nil {
				sendError(err)
				return
			}
		} else if isDeleteAccountProcess {
			if err := FunctionsOf(c).DeleteUserByID(c, userID); err != nil {
				sendError(err)
				return
			}

			queue.DeleteUser(queue.TaskDeleteUser{
				UserID: userID,
			})
		} else if isDisableAccountProcess {
			if _, err := FunctionsOf(c).DisableUserByID(c, userID); err != nil {
				sendError(err)
				return
			}

			queue.DisableUser(queue.TaskDisableUser{
				UserID: userID,
			})
		}
	}

	// Now delete the process.
	err = FunctionsOf(c).DeleteProcessByID(c, processID)
	if err != nil && !validate {
		sendError(err)
		return
	}

	// Send mail
	if validate {
		if isEmailAddressProcess {
			// The confirmation mail is sent to the previous email address.
			queue.SendMail(queue.MessageMailUpdateEmailAddressConfirmed, queue.MailUpdateEmailAddressConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})

			// Another mail is sent to the new email address to verify it.
			if _, _, err := startProcessValidateNewEmailAddress(c, user.ID, value, user.Username, user.Locale); err != nil {
				c.Error(err)
			}

			response.OkWithStatus(c, t("status.update_email_address.confirmed"))
		} else if isPasswordProcess {
			queue.SendMail(queue.MessageMailUpdatePasswordConfirmed, queue.MailUpdatePasswordConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_password.confirmed"))
		} else if isValidateEmailAddressProcess {
			response.OkWithStatus(c, t("status.validate_email_address.confirmed"))
		} else if isDeleteAccountProcess {
			queue.SendMail(queue.MessageMailDeleteAccountConfirmed, queue.MailDeleteAccountConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.delete_account.confirmed"))
		} else if isDisableAccountProcess {
			queue.SendMail(queue.MessageMailDisableAccountConfirmed, queue.MailDisableAccountConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.disable_account.confirmed"))
		}
	} else {
		if isEmailAddressProcess {
			queue.SendMail(queue.MessageMailUpdateEmailAddressCancelled, queue.MailUpdateEmailAddressCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_email_address.cancelled"))
		} else if isPasswordProcess {
			queue.SendMail(queue.MessageMailUpdatePasswordCancelled, queue.MailUpdatePasswordCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_password.cancelled"))
		} else if isDeleteAccountProcess {
			queue.SendMail(queue.MessageMailDeleteAccountCancelled, queue.MailDeleteAccountCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.delete_account.cancelled"))
		} else if isDisableAccountProcess {
			queue.SendMail(queue.MessageMailDisableAccountCancelled, queue.MailDisableAccountCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.disable_account.cancelled"))
		}
	}
}

func validatePassword(c *gin.Context, password, messageKey string) bool {
	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	message := t(messageKey)

	// Password is mandatory.
	if len(password) == 0 {
		response.NotFoundParameterValue(c, "password",
			message,
			t("error.login.password.not_found.reason"),
			t("error.login.password.not_found.recovery"),
		)
		return false
	}

	var reason, recovery string

	// Password must contain at least X characters.
	if len(password) < authentication.MinimumCharactersForPassword {
		reason = t("error.password.minimum_characters.reason", localization.H{"Count": authentication.MinimumCharactersForPassword})
		recovery = t("error.password.minimum_characters.recovery", localization.H{"Count": authentication.MinimumCharactersForPassword})
	}

	// Password must contain at least 1 lowercase character.
	if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		reason = t("error.password.missing_lowercase_character.reason")
		recovery = t("error.password.missing_lowercase_character.recovery")
	}

	// Password must contain at least 1 uppercase character.
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		reason = t("error.password.missing_uppercase_character.reason")
		recovery = t("error.password.missing_uppercase_character.recovery")
	}

	// Password must contain at least 1 numeric character.
	if !strings.ContainsAny(password, "0123456789") {
		reason = t("error.password.missing_numeric_character.reason")
		recovery = t("error.password.missing_numeric_character.recovery")
	}

	if len(reason) > 0 {
		response.InvalidParameterValue(c, "password", message, reason, recovery)
		return false
	}

	return true
}

func validateEmailAddress(c *gin.Context, emailAddress, messageKey string) bool {
	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	message := t(messageKey)

	// Email address is mandatory.
	if len(emailAddress) == 0 {
		response.NotFoundParameterValue(c, "email_address",
			message,
			t("error.registration.email.not_found.reason"),
			t("error.registration.email.not_found.recovery"),
		)
		return false
	}

	// Email address must be valid.
	if !strings.Contains(emailAddress, "@") {
		response.InvalidParameterValue(c, "email_address",
			message,
			t("error.registration.email.invalid.reason"),
			t("error.registration.email.invalid.recovery"),
		)
		return false
	}

	return true
}
