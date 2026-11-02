package model

import (
	"slices"
	"testing"

	"github.com/cygy/ginamite/api/rights"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()

	if s.UserMaxValidateRegistrationAttemps != 5 || s.UserMaxDescriptionLength != 300 || s.UserMaxImageSize != 2000000 ||
		s.UserImagePath != "/images" || s.UserImageWidth != 800 || s.UserImageHeight != 800 || !s.UserImageSmallVersion || !s.UserImageThumbVersion {
		t.Errorf("settings = %+v", s)
	}
	if !slices.Equal(s.AllUserAbilities, []string{rights.AbilityToManageUsers, rights.AbilityToManageContacts}) || s.DefaultUserAbilities == nil || len(s.DefaultUserAbilities) != 0 {
		t.Errorf("abilities = %v, %v", s.AllUserAbilities, s.DefaultUserAbilities)
	}

	// Each call returns independent lists.
	s.AllUserAbilities[0] = "changed"
	if DefaultSettings().AllUserAbilities[0] == "changed" {
		t.Error("the settings share their abilities")
	}
}
