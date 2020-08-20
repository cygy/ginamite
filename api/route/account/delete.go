package account

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// DeleteAccount : starts the delete processus and sends a mail to confirm.
func DeleteAccount(c *gin.Context) {
	if GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if CreateNewDeleteAccountProcess == nil {
		c.Error(errors.New("undefined function 'CreateNewDeleteAccountProcess'"))
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

	processID, processKey, err := CreateNewDeleteAccountProcess(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The process must be validated by the user.
	// In Env test, we return the code and the user ID.
	if config.Main.IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.delete_account.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	queue.SendMail(queue.MessageMailDeleteAccountConfirmation, queue.MailDeleteAccountConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})
	response.OkWithStatus(c, t("status.delete_account.email_sent"))
}

// ValidateDeleteAccount : confirms the delete process.
func ValidateDeleteAccount(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessDeleteAccount, true)
}

// CancelDeleteAccount : cancels the delete process.
func CancelDeleteAccount(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessDeleteAccount, false)
}
