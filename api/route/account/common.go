package account

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/api/validator"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// newTokenValidator returns the validator of the tokens issued by a third-party authentication method: the one of the
// functions of the context if any, the validators of Facebook and Google else. It returns nil for another method.
func newTokenValidator(c *gin.Context, method string) validator.TokenValidator {
	if newValidator := StorageOf(c).NewTokenValidator; newValidator != nil {
		return newValidator(c, method)
	}

	switch method {
	case authentication.MethodFacebook:
		return validator.NewFacebookTokenValidator(ctxkey.Config.Get(c).Facebook.AppID, ctxkey.Config.Get(c).Facebook.APIVersion, ctxkey.Config.Get(c).Facebook.TimeOut, ctxkey.Config.Get(c).Facebook.RetryCount, ctxkey.Config.Get(c).IsDebugModeEnabled())
	case authentication.MethodGoogle:
		return validator.NewGoogleTokenValidator(ctxkey.Config.Get(c).Google.ClientID, ctxkey.Config.Get(c).Google.APIVersion, ctxkey.Config.Get(c).Google.TimeOut, ctxkey.Config.Get(c).Google.RetryCount, ctxkey.Config.Get(c).IsDebugModeEnabled())
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
	t := ctxkey.Localization.Get(c).Translate(locale)

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

	if StorageOf(c).Accounts == nil || StorageOf(c).Processes == nil {
		sendError(errors.New("the storages of the accounts and of the processes must be defined"))
		return
	}

	// Verify the process.
	userID, value, err := StorageOf(c).Processes.VerifyProcess(c, processID, key)
	if err != nil {
		sendError(err)
		return
	}

	user, err := StorageOf(c).Accounts.GetUserDetailsByID(c, userID)
	if err != nil {
		sendError(err)
		return
	}

	// Now update the password/email address.
	if validate {
		if isEmailAddressProcess {
			if err := StorageOf(c).Accounts.UpdateUserEmailAddress(c, userID, value); err != nil {
				sendError(err)
				return
			}
		} else if isPasswordProcess {
			if err := StorageOf(c).Accounts.UpdateUserPassword(c, userID, value, true); err != nil {
				sendError(err)
				return
			}
		} else if isValidateEmailAddressProcess {
			if err := StorageOf(c).Accounts.UserEmailAddressIsVerified(c, userID); err != nil {
				sendError(err)
				return
			}
		} else if isDeleteAccountProcess {
			if err := StorageOf(c).Accounts.DeleteUserByID(c, userID); err != nil {
				sendError(err)
				return
			}

			ctxkey.Queue.Get(c).DeleteUser(queue.TaskDeleteUser{
				UserID: userID,
			})
		} else if isDisableAccountProcess {
			if _, err := StorageOf(c).Accounts.DisableUserByID(c, userID); err != nil {
				sendError(err)
				return
			}

			ctxkey.Queue.Get(c).DisableUser(queue.TaskDisableUser{
				UserID: userID,
			})
		}
	}

	// Now delete the process.
	err = StorageOf(c).Processes.DeleteProcessByID(c, processID)
	if err != nil && !validate {
		sendError(err)
		return
	}

	// Send mail
	if validate {
		if isEmailAddressProcess {
			// The confirmation mail is sent to the previous email address.
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailUpdateEmailAddressConfirmed, queue.MailUpdateEmailAddressConfirmed{
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
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailUpdatePasswordConfirmed, queue.MailUpdatePasswordConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_password.confirmed"))
		} else if isValidateEmailAddressProcess {
			response.OkWithStatus(c, t("status.validate_email_address.confirmed"))
		} else if isDeleteAccountProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailDeleteAccountConfirmed, queue.MailDeleteAccountConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.delete_account.confirmed"))
		} else if isDisableAccountProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailDisableAccountConfirmed, queue.MailDisableAccountConfirmed{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.disable_account.confirmed"))
		}
	} else {
		if isEmailAddressProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailUpdateEmailAddressCancelled, queue.MailUpdateEmailAddressCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_email_address.cancelled"))
		} else if isPasswordProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailUpdatePasswordCancelled, queue.MailUpdatePasswordCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.update_password.cancelled"))
		} else if isDeleteAccountProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailDeleteAccountCancelled, queue.MailDeleteAccountCancelled{
				EmailAddress: user.EmailAddress,
				Username:     user.Username,
				Locale:       user.Locale,
			})
			response.OkWithStatus(c, t("status.delete_account.cancelled"))
		} else if isDisableAccountProcess {
			ctxkey.Queue.Get(c).SendMail(queue.MessageMailDisableAccountCancelled, queue.MailDisableAccountCancelled{
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
	t := ctxkey.Localization.Get(c).Translate(locale)

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
	t := ctxkey.Localization.Get(c).Translate(locale)

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
