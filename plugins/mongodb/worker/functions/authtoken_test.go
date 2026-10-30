package functions

import (
	"context"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDeleteExpiredTokens(t *testing.T) {
	db := withMainSession(t)
	db.Coll(model.AuthTokenCollection).DeleteResult = &mongo.DeleteResult{DeletedCount: 4}

	if got := DeleteExpiredTokens(context.Background()); got != 4 {
		t.Errorf("count = %d", got)
	}

	// A failure is only logged.
	db.Coll(model.AuthTokenCollection).Err = errBoom
	if got := DeleteExpiredTokens(context.Background()); got != 0 {
		t.Errorf("count = %d", got)
	}
}
