package account

import (
	"errors"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	ctxUser "github.com/cygy/ginamite/api/context/user"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/ctxconfig"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

// DisableAccount starts the disable processus and sends a mail to confirm.
func DisableAccount(c *gin.Context) {
	if FunctionsOf(c).GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		response.InternalServerError(c)
		return
	}

	if FunctionsOf(c).CreateNewDisableAccountProcess == nil {
		c.Error(errors.New("undefined function 'CreateNewDisableAccountProcess'"))
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
	t := localization.Translate(locale)

	userID := ctxUser.GetID(c)
	user, err := FunctionsOf(c).GetUserDetailsByID(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	processID, processKey, err := FunctionsOf(c).CreateNewDisableAccountProcess(c, userID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	// The process must be validated by the user.
	// In Env test, we return the code and the user ID.
	if ctxconfig.Get(c).IsTestEnvironment() {
		response.Ok(c, response.H{
			"status": t("status.disable_account.email_sent"),
			"id":     processID,
			"key":    processKey,
		})
		return
	}

	queue.SendMail(queue.MessageMailDisableAccountConfirmation, queue.MailDisableAccountConfirmation{
		EmailAddress: user.EmailAddress,
		Username:     user.Username,
		Locale:       user.Locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})
	response.OkWithStatus(c, t("status.disable_account.email_sent"))
}

// ValidateDisableAccount confirms the disable process.
func ValidateDisableAccount(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessDisableAccount, true)
}

// CancelDisableAccount cancels the disable process.
func CancelDisableAccount(c *gin.Context) {
	validateOrCancelUpdateProperty(c, authentication.ProcessDisableAccount, false)
}
