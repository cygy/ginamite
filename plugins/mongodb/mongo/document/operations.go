package document

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/bson"
	"github.com/cygy/ginamite/plugins/mongodb/mongo/store"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SaveDocument inserts a document into a collection and returns its ID, generating one if the document has none.
func SaveDocument(ctx context.Context, document Document, collection string, db store.Database) (string, error) {
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

// UpdateDocument applies the update to the document documentID of a collection.
// It returns an error which is, and wraps, errors.ErrNotFound and mongo.ErrNoDocuments if no document matches.
func UpdateDocument(ctx context.Context, documentID bson.ObjectID, collection string, update any, db store.Database) error {
	result, err := db.Collection(collection).UpdateByID(ctx, documentID, update)
	if err == nil && result.MatchedCount == 0 {
		err = notFound(mongo.ErrNoDocuments)
	}
	log.DatabaseError(err, collection)
	return err
}

// UpdateDocumentsBySelector applies the update to the documents of a collection matching the selector
// and returns the number of documents matched.
func UpdateDocumentsBySelector(ctx context.Context, collection string, selector any, update any, db store.Database) (int64, error) {
	result, err := db.Collection(collection).UpdateMany(ctx, selector, update)
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.MatchedCount, nil
}

// UpsertDocumentBySelector applies the update to the first document of a collection matching the selector,
// inserting a document if there is none, and returns the number of documents matched.
func UpsertDocumentBySelector(ctx context.Context, collection string, selector any, update any, db store.Database) (int64, error) {
	result, err := db.Collection(collection).UpdateOne(ctx, selector, update, options.UpdateOne().SetUpsert(true))
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.MatchedCount, nil
}

// DeleteDocument deletes the document documentID from a collection.
// It returns an error which is, and wraps, errors.ErrNotFound and mongo.ErrNoDocuments if no document matches.
func DeleteDocument(ctx context.Context, documentID bson.ObjectID, collection string, db store.Database) error {
	result, err := db.Collection(collection).DeleteOne(ctx, bson.D{{Key: "_id", Value: documentID}})
	if err == nil && result.DeletedCount == 0 {
		err = notFound(mongo.ErrNoDocuments)
	}
	log.DatabaseError(err, collection)
	return err
}

// DeleteDocumentsBySelector deletes the documents of a collection matching the selector
// and returns the number of documents deleted.
func DeleteDocumentsBySelector(ctx context.Context, collection string, selector any, db store.Database) (int64, error) {
	result, err := db.Collection(collection).DeleteMany(ctx, selector)
	log.DatabaseError(err, collection)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetDocumentByID loads the document with the hexadecimal id from a collection into document.
// It returns a "not found" error, see errors.IsNotFound, if there is no such document.
func GetDocumentByID(ctx context.Context, document Document, collection string, id string, db store.Database) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	err = db.Collection(collection).FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(document)
	log.DatabaseError(err, collection)

	if stderrors.Is(err, mongo.ErrNoDocuments) || (err == nil && !document.IsSaved()) {
		return errors.ErrNotFound
	}

	return err
}

// GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit decodes the documents of a collection matching the selector
// into result, which must be a pointer to a slice. Only the fields are returned if not empty, the documents are
// sorted by the sort fields ("name", or "-name" for a descending order), and a negative offset or limit is ignored.
// The other GetDocuments functions are shortcuts for it.
func GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx context.Context, result any, collection string, selector any, fields bson.M, sort []string, offset, limit int64, db store.Database) error {
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

// GetDocumentsBySelectorAndSortWithOffsetAndLimit returns the documents matching the selector, sorted, from offset, up to limit.
func GetDocumentsBySelectorAndSortWithOffsetAndLimit(ctx context.Context, result any, collection string, selector any, sort []string, offset, limit int64, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, sort, offset, limit, db)
}

// GetDocumentsBySelectorAndSort returns the documents matching the selector, sorted.
func GetDocumentsBySelectorAndSort(ctx context.Context, result any, collection string, selector any, sort []string, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, sort, -1, -1, db)
}

// GetDocumentsBySelector returns the documents matching the selector.
func GetDocumentsBySelector(ctx context.Context, result any, collection string, selector any, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, []string{}, -1, -1, db)
}

// GetDocumentsBySelectorWithFields returns the fields of the documents matching the selector.
func GetDocumentsBySelectorWithFields(ctx context.Context, result any, collection string, selector any, fields bson.M, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, fields, []string{}, -1, -1, db)
}

// GetDocumentsBySelectorWithOffsetAndLimit returns the documents matching the selector, from offset, up to limit.
func GetDocumentsBySelectorWithOffsetAndLimit(ctx context.Context, result any, collection string, selector any, offset, limit int64, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, selector, bson.M{}, []string{}, offset, limit, db)
}

// GetDocumentsBySort returns all the documents of a collection, sorted.
func GetDocumentsBySort(ctx context.Context, result any, collection string, sort []string, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, sort, -1, -1, db)
}

// GetDocumentsBySortWithOffsetAndLimit returns the documents of a collection, sorted, from offset, up to limit.
func GetDocumentsBySortWithOffsetAndLimit(ctx context.Context, result any, collection string, sort []string, offset, limit int64, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, sort, offset, limit, db)
}

// GetDocuments returns all the documents of a collection.
func GetDocuments(ctx context.Context, result any, collection string, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, []string{}, -1, -1, db)
}

// GetDocumentsWithOffsetAndLimit returns the documents of a collection from offset, up to limit.
func GetDocumentsWithOffsetAndLimit(ctx context.Context, result any, collection string, offset, limit int64, db store.Database) error {
	return GetDocumentsBySelectorAndSortWithFieldsOffsetAndLimit(ctx, result, collection, nil, bson.M{}, []string{}, offset, limit, db)
}

// GetDocumentBySelectorAndSort loads the first document matching the selector, once sorted, into document.
func GetDocumentBySelectorAndSort(ctx context.Context, document Document, collection string, selector any, sort []string, db store.Database) error {
	opts := options.FindOne()

	if len(sort) > 0 {
		opts.SetSort(sortKeys(sort...))
	}

	err := notFound(db.Collection(collection).FindOne(ctx, selector, opts).Decode(document))
	log.DatabaseError(err, collection)

	return err
}

// GetDocumentBySelector loads the first document matching the selector into document.
func GetDocumentBySelector(ctx context.Context, document Document, collection string, selector any, db store.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, selector, []string{}, db)
}

// GetDocumentBySort loads the first document of a collection, once sorted, into document.
func GetDocumentBySort(ctx context.Context, document Document, collection string, sort []string, db store.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, nil, sort, db)
}

// GetDocument loads the first document of a collection into document.
func GetDocument(ctx context.Context, document Document, collection string, db store.Database) error {
	return GetDocumentBySelectorAndSort(ctx, document, collection, nil, []string{}, db)
}

// CountBySelector returns the number of documents of a collection matching the selector.
func CountBySelector(ctx context.Context, collection string, selector any, db store.Database) (int64, error) {
	count, err := db.Collection(collection).CountDocuments(ctx, selector)
	log.DatabaseError(err, collection)

	return count, err
}

// Count returns the number of documents of a collection.
func Count(ctx context.Context, collection string, db store.Database) (int64, error) {
	return CountBySelector(ctx, collection, nil, db)
}

// PipeAll decodes all the results of an aggregation pipeline into result, which must be a pointer to a slice.
func PipeAll(ctx context.Context, result any, collection string, pipeline any, db store.Database) error {
	cursor, err := db.Collection(collection).Aggregate(ctx, pipeline)
	if err == nil {
		err = cursor.All(ctx, result)
	}
	log.DatabaseError(err, collection)

	return err
}

// PipeOne decodes the first result of an aggregation pipeline into result.
// It returns an error which is, and wraps, errors.ErrNotFound and mongo.ErrNoDocuments if the pipeline has no result.
func PipeOne(ctx context.Context, result any, collection string, pipeline any, db store.Database) error {
	cursor, err := db.Collection(collection).Aggregate(ctx, pipeline)
	if err == nil {
		defer func() {
			if closeErr := cursor.Close(ctx); closeErr != nil {
				log.DatabaseError(closeErr, collection)
			}
		}()

		if cursor.Next(ctx) {
			err = cursor.Decode(result)
		} else if err = cursor.Err(); err == nil {
			err = notFound(mongo.ErrNoDocuments)
		}
	}
	log.DatabaseError(err, collection)

	return err
}

// notFound makes the error of the driver for a missing document also match errors.ErrNotFound, which is what the
// code independent of the database tests with errors.IsNotFound. Any other error is returned unchanged.
func notFound(err error) error {
	if stderrors.Is(err, mongo.ErrNoDocuments) && !stderrors.Is(err, errors.ErrNotFound) {
		return fmt.Errorf("%w: %w", errors.ErrNotFound, err)
	}

	return err
}
