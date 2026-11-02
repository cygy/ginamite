package document

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSortKeys(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   bson.D
	}{
		{"none", nil, bson.D{}},
		{"ascending", []string{"name"}, bson.D{{Key: "name", Value: 1}}},
		{"descending", []string{"-created_at"}, bson.D{{Key: "created_at", Value: -1}}},
		{"ordered", []string{"name", "-age", "city"}, bson.D{{Key: "name", Value: 1}, {Key: "age", Value: -1}, {Key: "city", Value: 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortKeys(tt.fields...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
