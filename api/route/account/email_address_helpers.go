package account

import (
	"errors"

	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

func startProcessValidateNewEmailAddress(c *gin.Context, userID, emailAddress, username, locale string) (string, string, error) {
	if CreateNewVerifyEmailAddressProcess == nil {
		return "", "", errors.New("undefined function 'CreateNewVerifyEmailAddressProcess'")
	}

	processID, processKey, err := CreateNewVerifyEmailAddressProcess(c, userID)
	if err != nil {
		return "", "", err
	}

	queue.SendMail(queue.MessageMailValidateNewEmailAddress, queue.MailValidateNewEmailAddress{
		EmailAddress: emailAddress,
		Username:     username,
		Locale:       locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})

	return processID, processKey, nil
}
