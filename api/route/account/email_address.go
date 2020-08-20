package account

import (
	"errors"
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	accountErrors "github.com/cygy/ginamite/common/errors/account"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"
	res "github.com/cygy/ginamite/common/response"

	"github.com/gin-gonic/gin"
)

// UpdateEmailAddress : updates the email address and sends a mail to confirm.
func UpdateEmailAddress(c *gin.Context) {
	if GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if CreateNewUpdateEmailAddressProcess == nil {
		c.Error(errors.New("undefined function 'CreateNewUpdateEmailAddressProcess'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		EmailAddress string `json:"email_address"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	emailAddress := jsonBody.EmailAddress

	// Email address is mandatory.
	if !validateEmailAddress(c, emailAddress, "error.update_email_address.message") {
		return
	}

	userID := context.GetUserID(c)
	user, err := GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The same email address can not be used.
	if strings.ToLower(emailAddress) == strings.ToLower(user.EmailAddress) {
		message := t("error.update_email_address.message")
		reason := t("error.update_email_address.same_email_address.reason")
		recovery := t("error.update_email_address.same_email_address.recovery")
		response.InvalidParameterValue(c, "email_address", message, reason, recovery)
		return
	}

	processID, processKey, err := CreateNewUpdateEmailAddressProcess(c, userID, emailAddress)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The email address must be validated by the user.
	// In Env test, we return the code and the user ID.
	if config.Main.IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.update_email_address.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	queue.SendMail(queue.MessageMailUpdateEmailAddressConfirmation, queue.MailUpdateEmailAddressConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		Data:         emailAddress,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})
	response.OkWithStatus(c, t("status.update_email_address.email_sent"))
}

// ValidateUpdateEmailAddress : confirms the email address update.
func ValidateUpdateEmailAddress(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdateEmailAddress, true)
}

// CancelUpdateEmailAddress : cancels the email address update.
func CancelUpdateEmailAddress(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdateEmailAddress, false)
}

// ValidateNewEmailAddress : valiadtes the new email address.
func ValidateNewEmailAddress(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessVerifyEmailAddress, true)
}

// ConfirmEmailAddress : validates the new email address.
func ConfirmEmailAddress(c *gin.Context) {
	if GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	userID := context.GetUserID(c)
	user, err := GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// Do not confirm an email address already confirmed.
	if user.IsEmailValid {
		response.PreconditionFailedWithError(c, res.Error{
			Domain:   accountErrors.Domain,
			Code:     accountErrors.EmailAddressAlreadyConfirmed,
			Message:  t("error.confirm_email_address.already_confirmed.message"),
			Reason:   t("error.confirm_email_address.already_confirmed.reason"),
			Recovery: t("error.confirm_email_address.already_confirmed.recovery"),
		})
		return
	}

	if processID, processKey, err := startProcessValidateNewEmailAddress(c, user.ID, user.EmailAddress, user.Username, user.Locale); err != nil {
		response.InternalServerError(c)
		return

		// The email address must be validated by the user.
		// In Env test, we return the code and the user ID.
	} else if config.Main.IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.update_email_address.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	response.OkWithStatus(c, t("status.update_email_address.email_sent"))
}
