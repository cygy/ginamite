package model

import (
	"reflect"
	"testing"
)

func TestNewOwnUser(t *testing.T) {
	user := &User{
		ID: mustObjectID(), Username: "john", Abilities: []string{"a"}, Verified: true, VersionOfTermsAccepted: "v2", PrivateKey: "secret",
		PublicInfos: UserPublicInfos{
			Description: map[string]string{"fr": "bonjour"}, Facebook: "fb", Twitter: "tw", Instagram: "ig",
			Image: testImage(),
		},
		PrivateInfos: UserPrivateInfos{FirstName: "John", LastName: "Doe", Email: "j@example.com", IsEmailValid: true},
		Settings:     UserSettings{Locale: "fr", Timezone: "Europe/Paris"},
	}

	got := NewOwnUser(user)
	if got.ID != user.ID || got.Username != "john" || !got.Verified || got.VersionOfTermsAccepted != "v2" || !reflect.DeepEqual(got.Abilities, []string{"a"}) {
		t.Errorf("got %+v", got)
	}
	if got.PublicInfos.Facebook != "fb" || got.PublicInfos.Twitter != "tw" || got.PublicInfos.Instagram != "ig" || got.PublicInfos.Image.Full.Size1x != "full1" {
		t.Errorf("public infos = %+v", got.PublicInfos)
	}
	if got.PrivateInfos != (OwnUserPrivateInfos{FirstName: "John", LastName: "Doe", Email: "j@example.com", IsEmailValid: true}) {
		t.Errorf("private infos = %+v", got.PrivateInfos)
	}
	if got.Settings != (OwnUserSettings{Locale: "fr", Timezone: "Europe/Paris"}) {
		t.Errorf("settings = %+v", got.Settings)
	}

	if !got.HasAbility("a") || got.HasAbility("b") {
		t.Error("HasAbility() does not match the abilities of the user")
	}
	if got.PublicInfos.LocalizedDescription("fr") != "bonjour" || got.PublicInfos.LocalizedDescription("en") != "" {
		t.Error("LocalizedDescription() does not match the descriptions")
	}
}
