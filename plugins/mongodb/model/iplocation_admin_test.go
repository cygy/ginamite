package model

import "testing"

func TestAdminIPLocation(t *testing.T) {
	got := NewAdminIPLocation(IPLocation{
		ID: mustObjectID(), IP: "1.2.3.4", Region: "IDF", RegionCode: "FR-IDF", City: "Paris",
		Country: IPLocationCountry{Code: "FR", Name: "France"}, Source: "icloud-private-relay",
	})
	if got.IP != "1.2.3.4" || got.Region != "IDF" || got.City != "Paris" || got.Country.Code != "FR" ||
		got.RegionCode != "FR-IDF" || got.Source != "icloud-private-relay" {
		t.Errorf("got %+v", got)
	}
	if got.IsEmpty() {
		t.Error("the location is not empty")
	}

	for name, location := range map[string]IPLocation{
		"nothing":      {},
		"only a city":  {City: "Paris"},
		"country code": {Country: IPLocationCountry{Code: "FR"}},
	} {
		if !NewAdminIPLocation(location).IsEmpty() {
			t.Errorf("%s: the location is empty for the admin", name)
		}
	}
}
