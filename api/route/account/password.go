package account

import (
	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// UpdatePassword updates the password and sends a mail to confirm.
func UpdatePassword(c *gin.Context) {
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
		Password string `json:"password"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

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

	userID := ctxUser.GetID(c)
	user, err := StorageOf(c).Accounts.GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	processID, processKey, err := StorageOf(c).Processes.CreateNewUpdatePasswordProcess(c, userID, encryptedPassword)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The password must be validated by the user.
	// In Env test, we return the code and the user ID.
	if ctxkey.Config.Get(c).IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.update_password.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	ctxkey.Queue.Get(c).SendMail(queue.MessageMailUpdatePasswordConfirmation, queue.MailUpdatePasswordConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})
	response.OkWithStatus(c, t("status.update_password.email_sent"))
}

// ValidateUpdatePassword confirms the password update.
func ValidateUpdatePassword(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdatePassword, true)
}

// CancelUpdatePassword cancels the password update.
func CancelUpdatePassword(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessUpdatePassword, false)
}
