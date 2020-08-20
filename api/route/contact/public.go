package contact

import (
	"strings"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common"
	"github.com/cygy/ginamite/common/localization"
	"github.com/gin-gonic/gin"
)

// Post : post a new contact document.
func Post(c *gin.Context) {
	var jsonBody struct {
		EmailAddress string `json:"email_address"`
		Subject      string `json:"subject"`
		Text         string `json:"text"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	message := t("error.contact.message")

	// Email address is mandatory.
	emailAddress := jsonBody.EmailAddress
	if len(emailAddress) == 0 {
		response.NotFoundParameterValue(c, "email_address",
			message,
			t("error.contact.email_address.not_found.reason"),
			t("error.contact.email_address.not_found.recovery"),
		)
		return
	}

	// Subject is mandatory.
	subject := jsonBody.Subject
	if len(subject) == 0 {
		response.NotFoundParameterValue(c, "subject",
			message,
			t("error.contact.subject.not_found.reason"),
			t("error.contact.subject.not_found.recovery"),
		)
		return
	}

	// Text is mandatory.
	text := jsonBody.Text
	if len(text) == 0 {
		response.NotFoundParameterValue(c, "text",
			message,
			t("error.contact.text.not_found.reason"),
			t("error.contact.text.not_found.recovery"),
		)
		return
	}

	// Subject must not exceed a max length.
	if len(subject) > common.ContactSubjectMaxLength {
		response.InvalidParameterValue(c, "subject",
			message,
			t("error.contact.subject.max_length_exceeded.reason"),
			t("error.contact.subject.max_length_exceeded.recovery", localization.H{"MaxLength": common.ContactSubjectMaxLength}),
		)
		return
	}

	// Text must not exceed a max length.
	if len(text) > common.ContactTextMaxLength {
		response.InvalidParameterValue(c, "text",
			message,
			t("error.contact.text.max_length_exceeded.reason"),
			t("error.contact.text.max_length_exceeded.recovery", localization.H{"MaxLength": common.ContactTextMaxLength}),
		)
		return
	}

	// Email address must be valid.
	if !strings.Contains(emailAddress, "@") {
		response.InvalidParameterValue(c, "email_address",
			message,
			t("error.contact.email_address.invalid.reason"),
			t("error.contact.email_address.invalid.recovery"),
		)
		return
	}

	if CreateContact != nil {
		if err := CreateContact(c, emailAddress, subject, text); err != nil {
			response.InternalServerError(c)
			return
		}
	}

	response.CreatedWithStatus(c, t("status.contact.created"))
}
