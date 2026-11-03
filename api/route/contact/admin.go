package contact

import (
	"errors"

	"github.com/cygy/ginamite/common/ctxkey"

	ctxLocale "github.com/cygy/ginamite/api/context/locale"
	"github.com/cygy/ginamite/api/request"
	"github.com/cygy/ginamite/api/response"
	"github.com/gin-gonic/gin"
)

// errNoStorage is logged when a route of the administration is called without storage of the contacts.
var errNoStorage = errors.New("undefined storage of the contacts")

// GetAdminContacts returns the contact documents.
func GetAdminContacts(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	offset := request.GetOffset(c)
	limit := request.GetLimit(c)

	contacts, err := storage.GetContacts(c, offset, limit)
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
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	contactID := getContactID(c)

	contact, err := storage.GetContact(c, contactID)
	if err != nil {
		response.SendError(c, err)
		return
	}

	response.Ok(c, response.H{
		"contact": contact,
	})
}

// Update updates a contact document.
func Update(c *gin.Context) {
	storage := StorageOf(c)
	if storage == nil {
		c.Error(errNoStorage)
		response.InternalServerError(c)
		return
	}

	var jsonBody struct {
		Done bool `json:"done"`
	}
	if !request.DecodeBody(c, &jsonBody) {
		return
	}

	contactID := getContactID(c)

	err := storage.UpdateContact(c, contactID, jsonBody.Done)
	if err != nil {
		response.SendError(c, err)
		return
	}

	locale := ctxLocale.Get(c)
	t := ctxkey.Localization.Get(c).Translate(locale)

	response.OkWithStatus(c, t("status.admin.contact.updated"))
}
