package model

import (
	"github.com/cygy/ginamite/common/mongo/document"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// HasAbility : returns true if the user has the ability.
func (user *User) HasAbility(ability string) bool {
	for _, a := range user.Abilities {
		if a == ability {
			return true
		}
	}

	return false
}

// SetAbilities : sets some abilities up to a user and saves the changes, returns an error if needed.
func (user *User) SetAbilities(abilities []string, db *mgo.Database) error {
	filteredAbilities := []string{}

	for _, ability := range abilities {
		found := false

		for _, userAbility := range AllUserAbilities {
			if ability == userAbility {
				found = true
				break
			}
		}

		if found {
			filteredAbilities = append(filteredAbilities, ability)
		}
	}

	return user.Update(bson.M{"$set": bson.M{"abilities": filteredAbilities}}, db)
}

// RemoveAbilities : removes some abilities from a user and saves the changes, returns an error if needed.
func (user *User) RemoveAbilities(abilities []string, db *mgo.Database) error {
	return user.Update(bson.M{"$pullAll": bson.M{"abilities": abilities}}, db)
}

// AddAbilities : adds some abilities to a user and saves the changes, returns an error if needed.
func (user *User) AddAbilities(abilities []string, db *mgo.Database) error {
	filteredAbilities := []string{}

	for _, ability := range abilities {
		found := false

		for _, userAbility := range AllUserAbilities {
			if ability == userAbility {
				found = true
				break
			}
		}

		if found {
			filteredAbilities = append(filteredAbilities, ability)
		}
	}

	return user.Update(bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": filteredAbilities}}}, db)
}

// GetUsersWithAbility : return the users with an ability.
func GetUsersWithAbility(ability string, db *mgo.Database) ([]User, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(&users, UserCollection, bson.M{"abilities": ability}, db)

	return users, err
}
