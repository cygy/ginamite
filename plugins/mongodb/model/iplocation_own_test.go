package model

import "testing"

func TestOwnIPLocation(t *testing.T) {
	got := NewOwnIPLocation(IPLocation{ID: mustObjectID(), IP: "1.2.3.4", Region: "IDF", City: "Paris", Country: IPLocationCountry{Code: "FR", Name: "France"}})
	if got.IP != "1.2.3.4" || got.Region != "IDF" || got.City != "Paris" || got.Country.Name != "France" {
		t.Errorf("got %+v", got)
	}
	if got.IsEmpty() {
		t.Error("the location is not empty")
	}

	if !NewOwnIPLocation(IPLocation{City: "Paris"}).IsEmpty() {
		t.Error("a location with only a city is empty")
	}
}
