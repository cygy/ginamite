// Package mongostore saves the payments of the plugin stripe in a collection of MongoDB.
package mongostore

import (
	"context"
	"fmt"
	"time"

	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"
	"github.com/cygy/ginamite/plugins/stripe"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DefaultCollection is the collection of the payments when none is given.
const DefaultCollection = "stripePayments"

// Payment is a payment as it is saved in the collection.
type Payment struct {
	ID                bson.ObjectID     `bson:"_id,omitempty"`
	Reference         string            `bson:"ref"`
	PaymentIntentID   string            `bson:"pId,omitempty"`
	CheckoutSessionID string            `bson:"sId,omitempty"`
	UserID            string            `bson:"uId,omitempty"`
	CustomerID        string            `bson:"cId,omitempty"`
	Amount            int64             `bson:"amount"`
	Currency          string            `bson:"currency"`
	LiveMode          bool              `bson:"live"`
	Status            stripe.Status     `bson:"status"`
	Metadata          map[string]string `bson:"md,omitempty"`
	Data              any               `bson:"data,omitempty"`
	CreatedAt         time.Time         `bson:"createdAt"`
	UpdatedAt         time.Time         `bson:"updatedAt"`
}

// Store saves the payments in a collection, see stripe.Store.
type Store struct {
	database   func() store.Database
	collection string
	now        func() time.Time
}

var _ stripe.Store = (*Store)(nil)

// New returns a Store saving the payments in the collection of the database, DefaultCollection if it is empty. The
// database is a function because it is opened after the plugins are created, when the server connects them.
func New(database func() store.Database, collection string) *Store {
	if collection == "" {
		collection = DefaultCollection
	}

	return &Store{database: database, collection: collection, now: time.Now}
}

// EnsureIndexes creates the unique index of the references, and the index of the payments of a user.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "ref", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "uId", Value: 1}, {Key: "createdAt", Value: -1}}},
	}

	for _, index := range indexes {
		if _, err := s.coll().Indexes().CreateOne(ctx, index); err != nil {
			return fmt.Errorf("unable to create the index of the payments: %w", err)
		}
	}

	return nil
}

// Save inserts the payment.
func (s *Store) Save(ctx context.Context, payment *stripe.Payment) error {
	_, err := s.coll().InsertOne(ctx, Payment{
		Reference:         payment.Reference,
		PaymentIntentID:   payment.PaymentIntentID,
		CheckoutSessionID: payment.CheckoutSessionID,
		UserID:            payment.UserID,
		CustomerID:        payment.CustomerID,
		Amount:            payment.Amount,
		Currency:          payment.Currency,
		LiveMode:          payment.LiveMode,
		Status:            payment.Status,
		Metadata:          payment.Metadata,
		Data:              payment.Data,
		CreatedAt:         payment.CreatedAt,
		UpdatedAt:         payment.UpdatedAt,
	})

	return err
}

// Claim marks the payment as done if it was not yet, in a single update, and reports whether it did.
func (s *Store) Claim(ctx context.Context, reference string) (bool, error) {
	result, err := s.coll().UpdateOne(ctx,
		bson.M{"ref": reference, "status": bson.M{"$ne": stripe.StatusDone}},
		bson.M{"$set": bson.M{"status": stripe.StatusDone, "updatedAt": s.now()}})
	if err != nil {
		return false, err
	}

	return result.ModifiedCount > 0, nil
}

// Release marks the payment as not done.
func (s *Store) Release(ctx context.Context, reference string) error {
	_, err := s.coll().UpdateOne(ctx, bson.M{"ref": reference},
		bson.M{"$set": bson.M{"status": stripe.StatusTodo, "updatedAt": s.now()}})

	return err
}

// Complete saves the IDs of the payment intent and of the customer.
func (s *Store) Complete(ctx context.Context, reference, paymentIntentID, customerID string) error {
	_, err := s.coll().UpdateOne(ctx, bson.M{"ref": reference},
		bson.M{"$set": bson.M{"pId": paymentIntentID, "cId": customerID, "updatedAt": s.now()}})

	return err
}

// Get returns the payment with the reference. The error matches mongo.ErrNoDocuments if there is none.
func (s *Store) Get(ctx context.Context, reference string) (Payment, error) {
	var payment Payment
	err := s.coll().FindOne(ctx, bson.M{"ref": reference}).Decode(&payment)

	return payment, err
}

// coll returns the collection of the payments.
func (s *Store) coll() store.Collection {
	return s.database().Collection(s.collection)
}
