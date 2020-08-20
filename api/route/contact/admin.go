package contact

import (
	"errors"

	"github.com/cygy/ginamite/api/context"
	"github.com/cygy/ginamite/api/middleware"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/cygy/ginamite/common/localization"
	"github.com/gin-gonic/gin"
)

// GetAdminContacts : returns the contact documents.
func GetAdminContacts(c *gin.Context) {
	if GetContacts == nil {
		c.Error(errors.New("undefined function 'GetContacts'"))
		response.InternalServerError(c)
		return
	}

	offset := middleware.GetOffsetFromRequest(c)
	limit := middleware.GetLimitFromRequest(c)

	contacts, err := GetContacts(c, offset, limit)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"contacts": contacts,
	})
}

// GetAdminContact : returns a contact document.
func GetAdminContact(c *gin.Context) {
	if GetContact == nil {
		c.Error(errors.New("undefined function 'GetContact'"))
		response.InternalServerError(c)
		return
	}

	errorMessageKey := "error.admin.contact.get.message"

	contactID, ok := checkContactID(c, errorMessageKey)
	if !ok {
		return
	}

	contact, err := GetContact(c, contactID)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Ok(c, response.H{
		"contact": contact,
	})
}

// Update : updates a contact document.
func Update(c *gin.Context) {
	errorMessageKey := "error.admin.contact.update.message"

	contactID, ok := checkContactID(c, errorMessageKey)
	if !ok {
		return
	}

	var jsonBody struct {
		Done bool `json:"done"`
	}
	request.DecodeBody(c, &jsonBody)

	locale := context.GetLocale(c)
	t := localization.Translate(locale)

	var err error
	if UpdateContact == nil {
		err = c.Error(errors.New("undefined function 'UpdateContact'"))
	} else {
		err = UpdateContact(c, contactID, jsonBody.Done)
	}

	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.OkWithStatus(c, t("status.admin.contact.updated"))
}
