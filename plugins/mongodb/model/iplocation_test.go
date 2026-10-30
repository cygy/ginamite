package model

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/common/mongo/bson"
	"github.com/cygy/ginamite/common/mongo/document"
	"github.com/cygy/ginamite/common/mongo/store/mock"
)

func TestIPLocation_Document(t *testing.T) {
	testDocument(t, IPLocationCollection, func() document.Document { return NewIPLocation() })
}

func TestIPLocation_GetByIPAddress(t *testing.T) {
	db := mock.NewDatabase("app")
	db.Coll(IPLocationCollection).One = mock.Result{Value: bson.M{"_id": mustObjectID(), "IP": "1.2.3.4", "city": "Paris"}}

	location := NewIPLocation()
	if err := location.GetByIPAddress(context.Background(), "1.2.3.4", db); err != nil || location.City != "Paris" {
		t.Errorf("got %+v, %v", location, err)
	}
	checkCall(t, db, IPLocationCollection, "FindOne", bson.M{"IP": "1.2.3.4"})
}

func TestIPLocation_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		location IPLocation
		want     bool
	}{
		{"empty", IPLocation{}, true},
		{"only an IP", IPLocation{IP: "1.2.3.4", City: "Paris"}, true},
		{"region", IPLocation{Region: "IDF"}, false},
		{"country name", IPLocation{Country: IPLocationCountry{Name: "France"}}, false},
		{"country code", IPLocation{Country: IPLocationCountry{Code: "FR"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.location.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v", got)
			}
		})
	}
}
