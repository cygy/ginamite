package account

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	commonErrors "github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/random"

	"github.com/gin-gonic/gin"
)

// errNoAccounts is logged when a route needing the accounts is called without storage of the accounts.
var errNoAccounts = errors.New("undefined storage of the accounts")

// errNoProcesses is logged when a route needing the processes is called without storage of the processes.
var errNoProcesses = errors.New("undefined storage of the processes")

// ForgotPassword sends a mail to reset the password.
func ForgotPassword(c *gin.Context) {
	if StorageOf(c).Accounts == nil {
		c.Error(errNoAccounts)
		response.InternalServerError(c)
		return
	}

	if StorageOf(c).Processes == nil {
		c.Error(errNoProcesses)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Identifier string `json:"identifier"` // must be the username or the email address
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	// Identifier is mandatory.
	identifier := jsonBody.Identifier
	if len(identifier) == 0 {
		message := t("error.forgot_password.identifier.not_found.message")
		reason := t("error.forgot_password.identifier.not_found.reason")
		recovery := t("error.forgot_password.identifier.not_found.recovery")

		response.NotFoundParameterValue(c, "identifier", message, reason, recovery)
		return
	}

	user, err := StorageOf(c).Accounts.GetUserDetailsByIdentifier(c, identifier)
	if err != nil {
		if commonErrors.IsNotFound(err) {
			// Do not reveal that the account does not exist: the response is the one of an account found, so that
			// this route can not be used to guess the identifiers of the users.
			response.OkWithStatus(c, t("status.forgot_password.email_sent"))
		} else {
			response.InternalServerError(c)
		}
		return
	}

	processPrivateKey := random.String(64)
	processID, err := StorageOf(c).Processes.NewForgotPasswordProcessForUser(c, user.ID, processPrivateKey)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The email address must be validated by the user.
	// In Env test, we return the code and the user ID.
	if ctxkey.Config.Get(c).IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.forgot_password.email_sent"),
			"id":     processID,
			"key":    processPrivateKey,
		})
		return
	}

	ctxkey.Queue.Get(c).SendMail(queue.MessageMailForgotPasswordConfirmation, queue.MailForgotPasswordConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processPrivateKey,
	})
	response.OkWithStatus(c, t("status.forgot_password.email_sent"))
}

// ValidateForgotPassword validates the reset password.
func ValidateForgotPassword(c *gin.Context) {
	validateOrCancelForgotPassword(c, true)
}

// CancelForgotPassword cancels the reset password.
func CancelForgotPassword(c *gin.Context) {
	validateOrCancelForgotPassword(c, false)
}
