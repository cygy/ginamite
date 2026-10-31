package model

import (
	"context"
	"slices"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store"
)

// HasAbility reports whether the user has the ability.
func (user *User) HasAbility(ability string) bool {
	return slices.Contains(user.Abilities, ability)
}

// SetAbilities sets some abilities up to a user and saves the changes, returning an error if needed. The abilities
// which are not in allAbilities are ignored.
func (user *User) SetAbilities(ctx context.Context, abilities, allAbilities []string, db store.Database) error {
	filteredAbilities := knownAbilities(allAbilities, abilities)

	return user.Update(ctx, bson.M{"$set": bson.M{"abilities": filteredAbilities}}, db)
}

// RemoveAbilities removes some abilities from a user and saves the changes, returning an error if needed.
func (user *User) RemoveAbilities(ctx context.Context, abilities []string, db store.Database) error {
	return user.Update(ctx, bson.M{"$pullAll": bson.M{"abilities": abilities}}, db)
}

// AddAbilities adds some abilities to a user and saves the changes, returning an error if needed. The abilities
// which are not in allAbilities are ignored.
func (user *User) AddAbilities(ctx context.Context, abilities, allAbilities []string, db store.Database) error {
	filteredAbilities := knownAbilities(allAbilities, abilities)

	return user.Update(ctx, bson.M{"$addToSet": bson.M{"abilities": bson.M{"$each": filteredAbilities}}}, db)
}

// GetUsersWithAbility returns the users with an ability.
func GetUsersWithAbility(ctx context.Context, ability string, db store.Database) ([]User, error) {
	users := []User{}
	err := document.GetDocumentsBySelector(ctx, &users, UserCollection, bson.M{"abilities": ability}, db)

	return users, err
}

// knownAbilities returns the abilities listed in allAbilities, in the order given.
func knownAbilities(allAbilities, abilities []string) []string {
	known := []string{}
	for _, ability := range abilities {
		if slices.Contains(allAbilities, ability) {
			known = append(known, ability)
		}
	}

	return known
}
