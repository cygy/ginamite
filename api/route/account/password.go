package account

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// UpdatePassword : updates the password and sends a mail to confirm.
func UpdatePassword(c *gin.Context) {
	if GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if CreateNewUpdatePasswordProcess == nil {
		c.Error(errors.New("undefined function 'CreateNewUpdatePasswordProcess'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Password string `json:"password"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	// Password is mandatory.
	password := jsonBody.Password
	if !validatePassword(c, password, "error.update_password.password.not_found.message") {
		return
	}

	encryptedPassword, err := authentication.EncryptPassword(password)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	userID := context.GetUserID(c)
	user, err := GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	processID, processKey, err := CreateNewUpdatePasswordProcess(c, userID, encryptedPassword)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The password must be validated by the user.
	// In Env test, we return the code and the user ID.
	if config.Main.IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.update_password.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	queue.SendMail(queue.MessageMailUpdatePasswordConfirmation, queue.MailUpdatePasswordConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})
	response.OkWithStatus(c, t("status.update_password.email_sent"))
}

// ValidateUpdatePassword : confirms the password update.
func ValidateUpdatePassword(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdatePassword, true)
}

// CancelUpdatePassword : cancels the password update.
func CancelUpdatePassword(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdatePassword, false)
}
