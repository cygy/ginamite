package functions

import (
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/gin-gonic/gin"
)

// GetContactsForAdmin : returns the contacts for the admin panel.
func GetContactsForAdmin(c *gin.Context, offset, limit int) (interface{}, error) {
	mongoSession := session.Get(c)
	contacts, err := model.GetContactsForAdmin(offset, limit, mongoSession)

	return contacts, err
}

// GetContactForAdmin : gets a contact.
func GetContactForAdmin(c *gin.Context, contactID string) (interface{}, error) {
	mongoSession := session.Get(c)

	contact := &model.Contact{}
	if err := contact.GetByID(contactID, mongoSession); err != nil {
		return nil, err
	}

	return model.NewAdminContact(contact), nil
}

// UpdateContact : updates a contact.
func UpdateContact(c *gin.Context, contactID string, done bool) error {
	mongoSession := session.Get(c)

	contact := &model.Contact{}
	if err := contact.GetByID(contactID, mongoSession); err != nil {
		return err
	}

	return contact.UpdateDone(done, mongoSession)
}

// CreateContact : creates a contact.
func CreateContact(c *gin.Context, emailAddress, subject, text string) error {
	mongoSession := session.Get(c)

	contact := model.NewContact()
	contact.EmailAddress = emailAddress
	contact.Subject = subject
	contact.Text = text

	_, err := contact.Save(mongoSession)
	return err
}
