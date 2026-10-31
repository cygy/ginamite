package functions

import (
	"context"
	"time"

	"github.com/cygy/ginamite/common/broker"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// dedupTimeout is the longest wait for the database when a message is checked or remembered.
const dedupTimeout = 5 * time.Second

// deduplicator is a broker.Deduplicator remembering the handled messages in the database, shared by the workers: the
// messages are still remembered after a restart. The documents expire, see model.ProcessedMessageTTL.
type deduplicator struct {
	functions *Functions
}

// Deduplicator returns the broker.Deduplicator working on the database. A failure of the database is logged and means
// the message is not known as handled: it is handled again rather than lost.
func (f *Functions) Deduplicator() broker.Deduplicator {
	return deduplicator{functions: f}
}

func (d deduplicator) key(group, id string) string { return group + "/" + id }

// Handled reports whether the group already handled the message.
func (d deduplicator) Handled(group, id string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dedupTimeout)
	defer cancel()

	count, err := d.functions.database().Collection(model.ProcessedMessageCollection).CountDocuments(ctx, bson.M{"_id": d.key(group, id)})
	if err != nil {
		log.WithFields(logrus.Fields{"group": group, "message_id": id, "error": err.Error()}).Error("unable to check whether the message was handled")

		return false
	}

	return count > 0
}

// MarkHandled remembers that the group handled the message.
func (d deduplicator) MarkHandled(group, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), dedupTimeout)
	defer cancel()

	document := bson.M{"_id": d.key(group, id), "createdAt": time.Now()}
	_, err := d.functions.database().Collection(model.ProcessedMessageCollection).InsertOne(ctx, document)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		log.WithFields(logrus.Fields{"group": group, "message_id": id, "error": err.Error()}).Error("unable to remember the handled message")
	}
}
