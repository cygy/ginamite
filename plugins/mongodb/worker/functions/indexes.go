package functions

import (
	"context"
	"fmt"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateIndexes creates the indexes of the collections.
func (f *Functions) CreateIndexes(ctx context.Context, indexes []IndexToCreate) {
	db := f.database()

	for _, index := range indexes {
		_, err := db.Collection(index.Collection).Indexes().CreateOne(ctx, index.Index)
		if err != nil {
			log.Error("create mongoDB index",
				"error", err.Error(),
				"collection", index.Collection,
			)
		}
	}
}

// RemoveIndexes removes indexes from the collections.
func (f *Functions) RemoveIndexes(ctx context.Context, indexes []IndexToDelete) {
	db := f.database()

	for _, index := range indexes {
		err := db.Collection(index.Collection).Indexes().DropOne(ctx, index.IndexName)
		if err != nil && !IsNotFound(err, errorCodeIndexNotFound, fmt.Sprintf("index not found with name [%s]", index.IndexName)) {
			log.Error("delete mongoDB index",
				"error", err.Error(),
				"collection", index.Collection,
				"index", index.IndexName,
			)
		}
	}
}

// CreateAllIndexes checks the indexes of the collections.
func (f *Functions) CreateAllIndexes(ctx context.Context) {
	indexes := []IndexToCreate{
		// Indexes of the collection "authTokens".
		{
			Collection: model.AuthTokenCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
					{
						Key:   "createdAt",
						Value: -1,
					},
				},
				Options: options.Index().
					SetName("userId_1_createdAt_-1_1"),
			},
		},
		{
			Collection: model.AuthTokenCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "expiresAt",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("expiresAt_1_1"),
			},
		},

		// Indexes of the collection "deletedUsers".
		{
			Collection: model.DeletedUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userId_1_1").
					SetUnique(true),
			},
		},

		// Indexes of the collection "disabledUsers".
		{
			Collection: model.DisabledUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userId_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.DisabledUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "username",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("username_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.DisabledUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "privateInfos.email",
						Value: 1,
					},
					{
						Key:   "privateInfos.isEmailValid",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("privateInfos.email_1_privateInfos.isEmailValid_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.DisabledUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "identificationMethods.facebook.userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("identificationMethods.facebook.userId_1_1").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{
						"identificationMethods.facebook.userId": bson.M{"$exists": true},
					}),
			},
		},
		{
			Collection: model.DisabledUserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "identificationMethods.google.userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("identificationMethods.google.userId_1_1").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{
						"identificationMethods.google.userId": bson.M{"$exists": true},
					}),
			},
		},

		// Indexes of the collection "forgotPassword".
		{
			Collection: model.ForgotPasswordProcessCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userId_1_1").
					SetUnique(true).
					SetExpireAfterSeconds(7200),
			},
		},

		// Indexes of the collection "IPLocation".
		{
			Collection: model.IPLocationCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "IP",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("IP_1_1").
					SetUnique(true).
					SetExpireAfterSeconds(1296000),
			},
		},

		// Index of the collection "processedMessages": the handled messages are forgotten after some days.
		{
			Collection: model.ProcessedMessageCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "createdAt",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("createdAt_1_1").
					SetExpireAfterSeconds(model.ProcessedMessageTTL),
			},
		},

		// Index of the collection "rateLimits": the counters are removed when their window ends.
		{
			Collection: model.RateLimitCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "expireAt",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("expireAt_1_1").
					SetExpireAfterSeconds(0),
			},
		},

		// Indexes of the collection "updateProperty".
		{
			Collection: model.UpdatePropertyProcessCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
					{
						Key:   "type",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userId_1_type_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.UpdatePropertyProcessCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userIdtype",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userIdtype_1_1").
					SetUnique(true).
					SetExpireAfterSeconds(7200),
			},
		},

		// Indexes of the collection "users".
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "username",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("username_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "privateInfos.email",
						Value: 1,
					},
					{
						Key:   "privateInfos.isEmailValid",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("privateInfos.email_1_privateInfos.isEmailValid_1_1").
					SetUnique(true),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "isAnonymous",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("isAnonymous_1_1"),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "identificationMethods.facebook.userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("identificationMethods.facebook.userId_1_1").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{
						"identificationMethods.facebook.userId": bson.M{"$exists": true},
					}),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "identificationMethods.google.userId",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("identificationMethods.google.userId_1_1").
					SetUnique(true).
					SetPartialFilterExpression(bson.M{
						"identificationMethods.google.userId": bson.M{"$exists": true},
					}),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "_id",
						Value: 1,
					},
					{
						Key:   "notifications",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("_id_1_notifications_1_1"),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "abilities",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("abilities_1_1"),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "lastLogin",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("lastLogin_1_1").
					SetSparse(true),
			},
		},
		{
			Collection: model.UserCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "registrationInfos.date",
						Value: 1,
					},
					{
						Key:   "lastLogin",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("registrationInfos.date_1_lastLogin_1_1"),
			},
		},

		// Indexes of the collection "contact".
		{
			Collection: model.ContactCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "createdAt",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("createdAt_1_1"),
			},
		},

		// Indexes of the collection "notifications".
		{
			Collection: model.NotificationCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
					{
						Key:   "createdAt",
						Value: -1,
					},
				},
				Options: options.Index().
					SetName("userId_1_createdAt_-1_1"),
			},
		},
		{
			Collection: model.NotificationCollection,
			Index: mongo.IndexModel{
				Keys: bson.D{
					{
						Key:   "userId",
						Value: 1,
					},
					{
						Key:   "read",
						Value: 1,
					},
				},
				Options: options.Index().
					SetName("userId_1_read_1_1"),
			},
		},
	}

	f.CreateIndexes(ctx, indexes)
}

// RemoveUnusedIndexes removes some indexes from the collections.
func (f *Functions) RemoveUnusedIndexes(ctx context.Context) {
	indexes := []IndexToDelete{
		// Indexes of the collection "authTokens".
		{
			Collection: model.AuthTokenCollection,
			IndexName:  "userId_1_createdAt_-1",
		},
		{
			Collection: model.AuthTokenCollection,
			IndexName:  "expiresAt_1",
		},

		// Indexes of the collection "deletedUsers".
		{
			Collection: model.DeletedUserCollection,
			IndexName:  "userId_1",
		},

		// Indexes of the collection "disabledUsers".
		{
			Collection: model.DisabledUserCollection,
			IndexName:  "privateInfos.email_1_1",
		},

		// Indexes of the collection "forgotPassword".
		{
			Collection: model.ForgotPasswordProcessCollection,
			IndexName:  "userId_1",
		},

		// Indexes of the collection "IPLocation".
		{
			Collection: model.IPLocationCollection,
			IndexName:  "IP_1",
		},

		// Indexes of the collection "notifications".
		{
			Collection: model.NotificationCollection,
			IndexName:  "userId_1_unread_1_1",
		},

		// Indexes of the collection "updateProperty".
		{
			Collection: model.UpdatePropertyProcessCollection,
			IndexName:  "userId_1_type_1",
		},
		{
			Collection: model.UpdatePropertyProcessCollection,
			IndexName:  "userIdtype_1",
		},

		// Indexes of the collection "users".
		{
			Collection: model.UserCollection,
			IndexName:  "username_1_enabled_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_enabled_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "isAnonymous_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "identificationMethods.facebook.userId_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "identificationMethods.google.userId_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "notifications_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "abilities_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "username_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "_id_1_notifications_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "username_1_enabled_1_1",
		},
		{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_enabled_1_1",
		},
	}

	f.RemoveIndexes(ctx, indexes)
}
