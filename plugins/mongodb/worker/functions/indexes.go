package functions

import (
	"fmt"
	"time"

	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/session"
	"github.com/cygy/ginamite/plugins/mongodb/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

// CreateIndexes : creates the indexes of the collections.
func CreateIndexes(indexes []IndexToCreate) {
	session, db := session.Copy()
	defer session.Close()

	for _, index := range indexes {
		err := db.C(index.Collection).EnsureIndex(index.Index)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": index.Collection,
				"index":      index.Index.Name,
			}).Error("create mongoDB index")
		}
	}
}

// RemoveIndexes : removes indexes from the collections.
func RemoveIndexes(indexes []IndexToDelete) {
	session, db := session.Copy()
	defer session.Close()

	for _, index := range indexes {
		err := db.C(index.Collection).DropIndexName(index.IndexName)
		if err != nil && err.Error() != fmt.Sprintf("index not found with name [%s]", index.IndexName) {
			log.WithFields(logrus.Fields{
				"error":      err.Error(),
				"collection": index.Collection,
				"index":      index.IndexName,
			}).Error("delete mongoDB index")
		}
	}
}

// CreateAllIndexes : checks the indexes of the collections.
func CreateAllIndexes() {
	indexes := []IndexToCreate{

		// Indexes of the collection "authTokens"
		IndexToCreate{
			Collection: model.AuthTokenCollection,
			Index: mgo.Index{
				Key:        []string{"userId", "-createdAt"},
				Background: true,
				Name:       "userId_1_createdAt_-1_1",
			},
		},
		IndexToCreate{
			Collection: model.AuthTokenCollection,
			Index: mgo.Index{
				Key:        []string{"expiresAt"},
				Background: true,
				Name:       "expiresAt_1_1",
			},
		},

		// Indexes of the collection "deletedUsers"
		IndexToCreate{
			Collection: model.DeletedUserCollection,
			Index: mgo.Index{
				Key:        []string{"userId"},
				Background: true,
				Unique:     true,
				Name:       "userId_1_1",
			},
		},

		// Indexes of the collection "disabledUsers"
		IndexToCreate{
			Collection: model.DisabledUserCollection,
			Index: mgo.Index{
				Key:        []string{"userId"},
				Background: true,
				Unique:     true,
				Name:       "userId_1_1",
			},
		},
		IndexToCreate{
			Collection: model.DisabledUserCollection,
			Index: mgo.Index{
				Key:        []string{"username"},
				Background: true,
				Unique:     true,
				Name:       "username_1_1",
			},
		},
		IndexToCreate{
			Collection: model.DisabledUserCollection,
			Index: mgo.Index{
				Key:        []string{"privateInfos.email", "privateInfos.isEmailValid"},
				Background: true,
				Unique:     true,
				Name:       "privateInfos.email_1_privateInfos.isEmailValid_1_1",
			},
		},
		IndexToCreate{
			Collection: model.DisabledUserCollection,
			Index: mgo.Index{
				Key:        []string{"identificationMethods.facebook.userId"},
				Background: true,
				Unique:     true,
				PartialFilter: bson.M{
					"identificationMethods.facebook.userId": bson.M{
						"$exists": true,
					},
				},
				Name: "identificationMethods.facebook.userId_1_1",
			},
		},
		IndexToCreate{
			Collection: model.DisabledUserCollection,
			Index: mgo.Index{
				Key:        []string{"identificationMethods.google.userId"},
				Background: true,
				Unique:     true,
				PartialFilter: bson.M{
					"identificationMethods.facebook.userId": bson.M{
						"$exists": true,
					},
				},
				Name: "identificationMethods.google.userId_1_1",
			},
		},

		// Indexes of the collection "forgotPassword"
		IndexToCreate{
			Collection: model.ForgotPasswordProcessCollection,
			Index: mgo.Index{
				Key:         []string{"userId"},
				Background:  true,
				Unique:      true,
				ExpireAfter: time.Duration(7200) * time.Second,
				Name:        "userId_1_1",
			},
		},

		// Indexes of the collection "IPLocation"
		IndexToCreate{
			Collection: model.IPLocationCollection,
			Index: mgo.Index{
				Key:         []string{"IP"},
				Background:  true,
				Unique:      true,
				ExpireAfter: time.Duration(1296000) * time.Second,
				Name:        "IP_1_1",
			},
		},

		// Indexes of the collection "updateProperty"
		IndexToCreate{
			Collection: model.UpdatePropertyProcessCollection,
			Index: mgo.Index{
				Key:        []string{"userId", "type"},
				Background: true,
				Unique:     true,
				Name:       "userId_1_type_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UpdatePropertyProcessCollection,
			Index: mgo.Index{
				Key:         []string{"userIdtype"},
				Background:  true,
				Unique:      true,
				ExpireAfter: time.Duration(7200) * time.Second,
				Name:        "userIdtype_1_1",
			},
		},

		// Indexes of the collection "users"
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"username"},
				Background: true,
				Unique:     true,
				Name:       "username_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"privateInfos.email", "privateInfos.isEmailValid"},
				Background: true,
				Unique:     true,
				Name:       "privateInfos.email_1_privateInfos.isEmailValid_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"isAnonymous"},
				Background: true,
				Name:       "isAnonymous_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"identificationMethods.facebook.userId"},
				Background: true,
				Unique:     true,
				PartialFilter: bson.M{
					"identificationMethods.facebook.userId": bson.M{
						"$exists": true,
					},
				},
				Name: "identificationMethods.facebook.userId_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"identificationMethods.google.userId"},
				Background: true,
				Unique:     true,
				PartialFilter: bson.M{
					"identificationMethods.google.userId": bson.M{
						"$exists": true,
					},
				},
				Name: "identificationMethods.google.userId_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"_id", "notifications"},
				Background: true,
				Name:       "_id_1_notifications_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"abilities"},
				Background: true,
				Name:       "abilities_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"lastLogin"},
				Background: true,
				Sparse:     true,
				Name:       "lastLogin_1_1",
			},
		},
		IndexToCreate{
			Collection: model.UserCollection,
			Index: mgo.Index{
				Key:        []string{"registrationInfos.date", "lastLogin"},
				Background: true,
				Name:       "registrationInfos.date_1_lastLogin_1_1",
			},
		},

		// Indexes of the collection "contact"
		IndexToCreate{
			Collection: model.ContactCollection,
			Index: mgo.Index{
				Key:        []string{"createdAt"},
				Background: true,
				Name:       "createdAt_1_1",
			},
		},

		// Indexes of the collection "notifications"
		IndexToCreate{
			Collection: model.NotificationCollection,
			Index: mgo.Index{
				Key:        []string{"userId", "-createdAt"},
				Background: true,
				Name:       "userId_1_createdAt_-1_1",
			},
		},
		IndexToCreate{
			Collection: model.NotificationCollection,
			Index: mgo.Index{
				Key:        []string{"userId", "read"},
				Background: true,
				Name:       "userId_1_read_1_1",
			},
		},
	}

	CreateIndexes(indexes)
}

// RemoveUnusedIndexes : removes some indexes from the collections.
func RemoveUnusedIndexes() {
	indexes := []IndexToDelete{

		// Indexes of the collection "authTokens"
		IndexToDelete{
			Collection: model.AuthTokenCollection,
			IndexName:  "userId_1_createdAt_-1",
		},
		IndexToDelete{
			Collection: model.AuthTokenCollection,
			IndexName:  "expiresAt_1",
		},

		// Indexes of the collection "deletedUsers"
		IndexToDelete{
			Collection: model.DeletedUserCollection,
			IndexName:  "userId_1",
		},

		// Indexes of the collection "disabledUsers"
		IndexToDelete{
			Collection: model.DisabledUserCollection,
			IndexName:  "privateInfos.email_1_1",
		},

		// Indexes of the collection "forgotPassword"
		IndexToDelete{
			Collection: model.ForgotPasswordProcessCollection,
			IndexName:  "userId_1",
		},

		// Indexes of the collection "IPLocation"
		IndexToDelete{
			Collection: model.IPLocationCollection,
			IndexName:  "IP_1",
		},

		// Indexes of the collection "notifications"
		IndexToDelete{
			Collection: model.NotificationCollection,
			IndexName:  "userId_1_unread_1_1",
		},

		// Indexes of the collection "updateProperty"
		IndexToDelete{
			Collection: model.UpdatePropertyProcessCollection,
			IndexName:  "userId_1_type_1",
		},
		IndexToDelete{
			Collection: model.UpdatePropertyProcessCollection,
			IndexName:  "userIdtype_1",
		},

		// Indexes of the collection "users"
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "username_1_enabled_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_enabled_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "isAnonymous_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "identificationMethods.facebook.userId_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "identificationMethods.google.userId_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "notifications_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "abilities_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "username_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "_id_1_notifications_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "username_1_enabled_1_1",
		},
		IndexToDelete{
			Collection: model.UserCollection,
			IndexName:  "privateInfos.email_1_enabled_1_1",
		},
	}

	RemoveIndexes(indexes)
}
