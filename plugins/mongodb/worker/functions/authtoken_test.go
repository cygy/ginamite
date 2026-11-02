package functions

import (
	"context"
	"errors"
	"testing"

	"github.com/cygy/ginamite/plugins/mongodb/model"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestDeleteExpired(t *testing.T) {
	fns, db := withMainSession(t)
	db.Coll(model.AuthTokenCollection).DeleteResult = &mongo.DeleteResult{DeletedCount: 4}

	if got, err := fns.DeleteExpired(context.Background()); err != nil || got != 4 {
		t.Errorf("DeleteExpired() = %d, %v", got, err)
	}

	db.Coll(model.AuthTokenCollection).Err = errBoom
	if _, err := fns.DeleteExpired(context.Background()); !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want the error of the database", err)
	}
}
