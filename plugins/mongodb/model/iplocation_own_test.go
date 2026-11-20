package model

import "testing"

func TestOwnIPLocation(t *testing.T) {
	got := NewOwnIPLocation(IPLocation{
		ID: mustObjectID(), IP: "1.2.3.4", Region: "IDF", RegionCode: "FR-IDF", City: "Paris",
		Country: IPLocationCountry{Code: "FR", Name: "France"}, Source: "icloud-private-relay",
	})
	if got.IP != "1.2.3.4" || got.Region != "IDF" || got.City != "Paris" || got.Country.Name != "France" ||
		got.RegionCode != "FR-IDF" || got.Source != "icloud-private-relay" {
		t.Errorf("got %+v", got)
	}
	if got.IsEmpty() {
		t.Error("the location is not empty")
	}

	if !NewOwnIPLocation(IPLocation{City: "Paris"}).IsEmpty() {
		t.Error("a location with only a city is empty")
	}
}
