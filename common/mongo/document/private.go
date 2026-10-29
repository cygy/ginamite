package document

import (
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// sortKeys : converts mgo-style sort fields ("name", "-created_at") to an ordered bson.D.
func sortKeys(fields ...string) bson.D {
	keys := make(bson.D, 0, len(fields))
	for _, field := range fields {
		value := 1
		if strings.HasPrefix(field, "-") {
			field, value = field[1:], -1
		}
		keys = append(keys, bson.E{Key: field, Value: value})
	}
	return keys
}
