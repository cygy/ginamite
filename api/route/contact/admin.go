package contact

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxlocalization"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/gin-gonic/gin"
)

// GetAdminContacts returns the contact documents.
func GetAdminContacts(c *gin.Context) {
	if FunctionsOf(c).GetContacts == nil {
		c.Error(errors.New("undefined function 'GetContacts'"))
		response.InternalServerError(c)
		return
	}

	offset := request.GetOffset(c)
	limit := request.GetLimit(c)

	contacts, err := FunctionsOf(c).GetContacts(c, offset, limit)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"contacts": contacts,
	})
}

// GetAdminContact returns a contact document.
func GetAdminContact(c *gin.Context) {
	if FunctionsOf(c).GetContact == nil {
		c.Error(errors.New("undefined function 'GetContact'"))
		response.InternalServerError(c)
		return
	}

	contactID := getContactID(c)

	contact, err := FunctionsOf(c).GetContact(c, contactID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"contact": contact,
	})
}

// Update updates a contact document.
func Update(c *gin.Context) {
	if FunctionsOf(c).UpdateContact == nil {
		c.Error(errors.New("undefined function 'UpdateContact'"))
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Done bool `json:"done"`
	}
	request.DecodeBody(c, &jsonBody)

	contactID := getContactID(c)

	err := FunctionsOf(c).UpdateContact(c, contactID, jsonBody.Done)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxlocalization.Get(c).Translate(locale)

	response.OkWithStatus(c, t("status.admin.contact.updated"))
}
