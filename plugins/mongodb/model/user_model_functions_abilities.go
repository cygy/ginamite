package model

import (
	"context"

	"github.com/cygy/ginamite/common/mongo/document"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
func (user *User) SetAbilities(ctx context.Context, abilities []string, db *mongo.Database) error {
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

	return user.Update(ctx, bson.M{"$set": bson.M{"abilities": filteredAbilities}}, db)
}

// RemoveAbilities : removes some abilities from a user and saves the changes, returns an error if needed.
func (user *User) RemoveAbilities(ctx context.Context, abilities []string, db *mongo.Database) error {
	return user.Update(ctx, bson.M{"$pullAll": bson.M{"abilities": abilities}}, db)
}

// AddAbilities : adds some abilities to a user and saves the changes, returns an error if needed.
func (user *User) AddAbilities(ctx context.Context, abilities []string, db *mongo.Database) error {
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

	return user.Update(ctx, bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": filteredAbilities}}}, db)
}

// GetUsersWithAbility : return the users with an ability.
func GetUsersWithAbility(ctx context.Context, ability string, db *mongo.Database) ([]User, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection, bson.M{"abilities": ability}, db)

	return users, err
}
