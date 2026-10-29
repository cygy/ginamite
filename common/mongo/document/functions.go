package document

import (
	"context"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/common/mongo/bson"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SaveDocument : inserts a document into a collection.
func SaveDocument(ctx context.Context, document Document, collection string, db *mongo.Database) (string, error) {
	objectID := document.GetID()
	// With mongo-driver, the Hex() of an empty ObjectID is 24 zeros, not "".
	if bson.IsNilObjectID(objectID) {
		objectID = bson.NewObjectID().Hex()
		if err := document.SetID(objectID); err != nil {
			return "", err
		}
	}

	_, err := db.Collection(collection).InsertOne(ctx, document)
	log.DatabaseError(err, collection)

	return objectID, err
}

// UpdateDocument : updates a document into a collection.
func UpdateDocument(ctx context.Context, documentID bson.ObjectID, collection string, update any, db *mongo.Database) error {
	result, err := db.Collection(collection).UpdateByID(ctx, documentID, update)
	if err == nil && result.MatchedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	log.DatabaseError(err, collection)
	return err
}

// UpdateDocumentsBySelector : updates documents into a collection.
func UpdateDocumentsBySelector(ctx context.Context, collection string, selector any, update any, db *mongo.Database) (int64, error) {
	result, err := db.Collection(collection).UpdateMany(ctx, selector, update)
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.MatchedCount, nil
}

// UpsertDocumentBySelector : updates documents into a collection.
func UpsertDocumentBySelector(ctx context.Context, collection string, selector any, update any, db *mongo.Database) (int64, error) {
	result, err := db.Collection(collection).UpdateOne(ctx, selector, update, options.UpdateOne().SetUpsert(true))
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.MatchedCount, nil
}

// DeleteDocument : deletes a document from a collection.
func DeleteDocument(ctx context.Context, documentID bson.ObjectID, collection string, db *mongo.Database) error {
	result, err := db.Collection(collection).DeleteOne(ctx, bson.D{{Key: "_id", Value: documentID}})
	if err == nil && result.DeletedCount == 0 {
		err = mongo.ErrNoDocuments
	}
	log.DatabaseError(err, collection)
	return err
}

// DeleteDocumentsBySelector : deletes documents into a collection.
func DeleteDocumentsBySelector(ctx context.Context, collection string, selector any, db *mongo.Database) (int64, error) {
	result, err := db.Collection(collection).DeleteMany(ctx, selector)
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetDocumentByID : returns a document from a collection.
func GetDocumentByID(ctx context.Context, document Document, collection string, id string, db *mongo.Database) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	err = db.Collection(collection).FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(document)
	log.DatabaseError(err, collection)

	if !document.IsSaved() {
		return errors.NotFound()
	}

	return err
}

// GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx context.Context, result any, collection string, selector any, fields bson.M, sort []string, offset, limit int64, db *mongo.Database) error {
	opts := options.Find()

	if len(fields) > 0 {
		opts.SetProjection(fields)
	}

	if len(sort) > 0 {
		opts.SetSort(sortKeys(sort...))
	}

	if offset > -1 {
		opts.SetSkip(int64(offset))
	}

	if limit > -1 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := db.Collection(collection).Find(ctx, selector, opts)
	if err == nil {
		// All() decodes every document into result (a pointer to a slice) and closes the cursor.
		err = cursor.All(ctx, result)
	}
	log.DatabaseError(err, collection)

	return err
}

// GetDocumentsBySelectorAndSortWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySelectorAndSortWithOffsetAndLimit(ctx context.Context, result any, collection string, selector any, sort []string, offset, limit int64, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, sort, offset, limit, db)
}

// GetDocumentsBySelectorAndSort : returns the documents from a collection
func GetDocumentsBySelectorAndSort(ctx context.Context, result any, collection string, selector any, sort []string, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, sort, -1, -1, db)
}

// GetDocumentsBySelector : returns the documents from a collection
func GetDocumentsBySelector(ctx context.Context, result any, collection string, selector any, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, []string{}, -1, -1, db)
}

// GetDocumentsBySelectorWithFields : returns the documents from a collection
func GetDocumentsBySelectorWithFields(ctx context.Context, result any, collection string, selector any, fields bson.M, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, fields, []string{}, -1, -1, db)
}

// GetDocumentsBySelectorWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySelectorWithOffsetAndLimit(ctx context.Context, result any, collection string, selector any, offset, limit int64, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, []string{}, offset, limit, db)
}

// GetDocumentsBySort : returns the documents from a collection
func GetDocumentsBySort(ctx context.Context, result any, collection string, sort []string, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, sort, -1, -1, db)
}

// GetDocumentsBySortWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySortWithOffsetAndLimit(ctx context.Context, result any, collection string, sort []string, offset, limit int64, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, sort, offset, limit, db)
}

// GetDocuments : returns the documents from a collection
func GetDocuments(ctx context.Context, result any, collection string, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, []string{}, -1, -1, db)
}

// GetDocumentsWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsWithOffsetAndLimit(ctx context.Context, result any, collection string, offset, limit int64, db *mongo.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, []string{}, offset, limit, db)
}

// GetDocumentBySelectorAndSort : returns a document from a collection
func GetDocumentBySelectorAndSort(ctx context.Context, document Document, collection string, selector any, sort []string, db *mongo.Database) error {
	opts := options.FindOne()

	if len(sort) > 0 {
		opts.SetSort(sortKeys(sort...))
	}

	err := db.Collection(collection).FindOne(ctx, selector, opts).Decode(document)
	log.DatabaseError(err, collection)

	return err
}

// GetDocumentBySelector : returns a document from a collection
func GetDocumentBySelector(ctx context.Context, document Document, collection string, selector any, db *mongo.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, selector, []string{}, db)
}

// GetDocumentBySort : returns a document from a collection
func GetDocumentBySort(ctx context.Context, document Document, collection string, sort []string, db *mongo.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, nil, sort, db)
}

// GetDocument : returns a document from a collection
func GetDocument(ctx context.Context, document Document, collection string, db *mongo.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, nil, []string{}, db)
}

// CountBySelector : returns the count of documents from a collection
func CountBySelector(ctx context.Context, collection string, selector any, db *mongo.Database) (int64, error) {
	count, err := db.Collection(collection).CountDocuments(ctx, selector)
	log.DatabaseError(err, collection)

	return count, err
}

// Count : returns the count of documents from a collection
func Count(ctx context.Context, collection string, db *mongo.Database) (int64, error) {
	return CountBySelector(ctx, collection, nil, db)
}

// PipeAll : returns all the results of a pipeline.
func PipeAll(ctx context.Context, result any, collection string, pipeline any, db *mongo.Database) error {
	cursor, err := db.Collection(collection).Aggregate(ctx, pipeline)
	if err == nil {
		err = cursor.All(ctx, result)
	}
	log.DatabaseError(err, collection)

	return err
}

// PipeOne : returns the first result of a pipeline.
func PipeOne(ctx context.Context, result any, collection string, pipeline any, db *mongo.Database) error {
	cursor, err := db.Collection(collection).Aggregate(ctx, pipeline)
	if err == nil {
		defer cursor.Close(ctx)

		if cursor.Next(ctx) {
			err = cursor.Decode(result)
		} else if err = cursor.Err(); err == nil {
			err = mongo.ErrNoDocuments
		}
	}
	log.DatabaseError(err, collection)

	return err
}
