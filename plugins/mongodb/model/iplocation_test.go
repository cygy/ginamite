package model

import (
	"context"
	"testing"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/document"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store/mock"
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

func TestIPLocationCountry_Flag(t *testing.T) {
	tests := map[string]string{"FR": "🇫🇷", "us": "🇺🇸", "": "", "F": "", "FRA": "", "F1": "", "É1": ""}
	for code, want := range tests {
		if got := (IPLocationCountry{Code: code}).Flag(); got != want {
			t.Errorf("Flag(%q) = %q, want %q", code, got, want)
		}
	}
}

func TestIPLocation_IsFresh(t *testing.T) {
	now := time.Now()
	const day = 24 * time.Hour
	tests := map[string]struct {
		updatedAt time.Time
		want      bool
	}{
		"never updated":        {time.Time{}, false},
		"updated an hour ago":  {now.Add(-time.Hour), true},
		"updated one day ago":  {now.Add(-day), false},
		"updated two days ago": {now.Add(-48 * time.Hour), false},
	}
	for name, tt := range tests {
		location := &IPLocation{UpdatedAt: tt.updatedAt}
		if got := location.IsFresh(now, day); got != tt.want {
			t.Errorf("%s: IsFresh = %v, want %v", name, got, tt.want)
		}
	}
}
