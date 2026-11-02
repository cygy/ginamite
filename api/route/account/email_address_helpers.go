package account

import (
	"github.com/cygy/ginamite/common/ctxkey"

	"github.com/cygy/ginamite/common/queue"

	"github.com/gin-gonic/gin"
)

func startProcessValidateNewEmailAddress(c *gin.Context, userID, emailAddress, username, locale string) (string, string, error) {
	if StorageOf(c).Processes == nil {
		return "", "", errNoProcesses
	}

	processID, processKey, err := StorageOf(c).Processes.CreateNewVerifyEmailAddressProcess(c, userID)
	if err != nil {
		return "", "", err
	}

	ctxkey.Queue.Get(c).SendMail(queue.MessageMailValidateNewEmailAddress, queue.MailValidateNewEmailAddress{
		EmailAddress: emailAddress,
		Username:     username,
		Locale:       locale,
		ProcessID:    processID,
		ProcessCode:  processKey,
	})

	return processID, processKey, nil
}
