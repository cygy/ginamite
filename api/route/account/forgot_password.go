package account

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/config"
	commonErrors "github.com/cygy/ginamite/common/errors"
	authError "github.com/cygy/ginamite/common/errors/auth"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	"github.com/cygy/ginamite/common/random"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// ForgotPassword : sends a mail to reset the password.
func ForgotPassword(c *gin.Context) {
	if GetUserDetailsByIdentifier == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByIdentifier'"))
		response.InternalServerError(c)
		return
	}

	if NewForgotPasswordProcessForUser == nil {
		c.Error(errors.New("undefined function 'NewForgotPasswordProcessForUser'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Identifier string `json:"identifier"` // must be the username or the email address
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	// Identifier is mandatory.
	identifier := jsonBody.Identifier
	if len(identifier) == 0 {
		message := t("error.forgot_password.identifier.not_found.message")
		reason := t("error.forgot_password.identifier.not_found.reason")
		recovery := t("error.forgot_password.identifier.not_found.recovery")

		response.NotFoundParameterValue(c, "identifier", message, reason, recovery)
		return
	}

	user, err := GetUserDetailsByIdentifier(c, identifier)
	if err != nil {
		if commonErrors.IsNotFound(err) {
			message := t("error.forgot_password.account.not_found.message")
			reason := t("error.forgot_password.account.not_found.reason")
			recovery := t("error.forgot_password.account.not_found.recovery")

			response.NotFoundWithError(c, res.Error{
				Domain:   authError.Domain,
				Code:     authError.AccountNotFound,
				Message:  message,
				Reason:   reason,
				Recovery: recovery,
			})
		} else {
			response.InternalServerError(c)
		}
		return
	}

	processPrivateKey := random.String(64)
	processID, err := NewForgotPasswordProcessForUser(c, user.ID, processPrivateKey)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The email address must be validated by the user.
	// In Env test, we return the code and the user ID.
	if config.Main.IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.forgot_password.email_sent"),
			"id":     processID,
			"key":    processPrivateKey,
		})
		return
	}

	queue.SendMail(queue.MessageMailForgotPasswordConfirmation, queue.MailForgotPasswordConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processPrivateKey,
	})
	response.OkWithStatus(c, t("status.forgot_password.email_sent"))
}

// ValidateForgotPassword : validates the reset password.
func ValidateForgotPassword(c *gin.Context) {
	validateOrCancelForgotPassword(c, true)
}

// CancelForgotPassword : cancels the reset password.
func CancelForgotPassword(c *gin.Context) {
	validateOrCancelForgotPassword(c, false)
}
