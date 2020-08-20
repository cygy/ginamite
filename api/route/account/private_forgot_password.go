package account

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/authentication"
	"github.com/cygy/ginamite/common/localization"
	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

func validateOrCancelForgotPassword(c *gin.Context, validate bool) {
	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	// Here is a local function that returns an error. This route does only return a single error.
	sendError := func() {
		var message string
		if validate {
			message = t("error.forgot_password.validate.message")
		} else {
			message = t("error.forgot_password.cancel.message")
		}

		response.InvalidRequestParameterWithDetails(c,
			message,
			t("error.forgot_password.validate_or_cancel.reason"),
			t("error.forgot_password.validate_or_cancel.recovery"),
		)
	}

	if GetUserDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetUserDetailsByID'"))
		sendError()
		return
	}

	if GetForgotPasswordProcessDetailsByID == nil {
		c.Error(errors.New("undefined function 'GetForgotPasswordProcessDetailsByID'"))
		sendError()
		return
	}

	if UpdateUserPassword == nil {
		c.Error(errors.New("undefined function 'UpdateUserPassword'"))
		sendError()
		return
	}

	if DeleteForgotPasswordProcess == nil {
		c.Error(errors.New("undefined function 'DeleteForgotPasswordProcess'"))
		sendError()
		return
	}

	// ID and key are mandatory.
	processID := c.Param("id")
	key := c.Param("key")
	if len(processID) == 0 || len(key) == 0 {
		sendError()
		return
	}

	// New password mmust be provided if the processed is validated.
	var password string
	if validate {
		var jsonBody struct {
			Password string `json:"password"`
		}
		request.DecodeBody(c, &jsonBody)

		password = jsonBody.Password
		if !validatePassword(c, password, "error.forgot_password.validate.message") {
			return
		}
	}

	userID, processPrivateKey, err := GetForgotPasswordProcessDetailsByID(c, processID)
	if err != nil {
		sendError()
		return
	}

	if processPrivateKey != key {
		if WrongAccessToForgotPasswordProcess != nil {
			WrongAccessToForgotPasswordProcess(c, processID)
		}
		sendError()
		return
	}

	user, err := GetUserDetailsByID(c, userID)
	if err != nil {
		sendError()
		return
	}

	if validate {
		encryptedPassword, err := authentication.EncryptPassword(password)
		if err != nil {
			sendError()
			return
		}
		if err := UpdateUserPassword(c, userID, encryptedPassword, true); err != nil {
			sendError()
			return
		}
	}

	if err := DeleteForgotPasswordProcess(c, processID); err != nil && !validate {
		sendError()
		return
	}

	// Send mail
	if validate {
		queue.SendMail(queue.MessageMailForgotPasswordConfirmed, queue.MailForgotPasswordConfirmed{
			EmailAddress: user.EmailAddress,
			Username:     user.Username,
			Locale:       user.Locale,
		})
		response.OkWithStatus(c, t("status.forgot_password.confirmed"))
	} else {
		queue.SendMail(queue.MessageMailForgotPasswordCancelled, queue.MailForgotPasswordCancelled{
			EmailAddress: user.EmailAddress,
			Username:     user.Username,
			Locale:       user.Locale,
		})
		response.OkWithStatus(c, t("status.forgot_password.cancelled"))
	}
}
