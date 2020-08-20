package document

import (
	"github.com/cygy/ginamite/common/errors"
	"github.com/cygy/ginamite/common/log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// SaveDocument : inserts a document into a collection.
func SaveDocument(document Document, collection string, db *mgo.Database) (string, error) {
	objectID := document.GetID()
	if len(objectID) == 0 {
		objectID = bson.NewObjectId().Hex()
		document.SetID(objectID)
	}

	err := db.C(collection).Insert(document)
	log.DatabaseError(err, collection)

	return objectID, err
}

// UpdateDocument : updates a document into a collection.
func UpdateDocument(documentID bson.ObjectId, collection string, update interface{}, db *mgo.Database) error {
	err := db.C(collection).UpdateId(documentID, update)
	log.DatabaseError(err, collection)
	return err
}

// UpdateDocumentsBySelector : updates documents into a collection.
func UpdateDocumentsBySelector(collection string, selector interface{}, update interface{}, db *mgo.Database) (int, error) {
	info, err := db.C(collection).UpdateAll(selector, update)
	log.DatabaseError(err, collection)
	return info.Updated, err
}

// UpsertDocumentBySelector : updates documents into a collection.
func UpsertDocumentBySelector(collection string, selector interface{}, update interface{}, db *mgo.Database) (int, error) {
	info, err := db.C(collection).Upsert(selector, update)
	log.DatabaseError(err, collection)
	return info.Updated, err
}

// DeleteDocument : deletes a document from a collection.
func DeleteDocument(documentID bson.ObjectId, collection string, db *mgo.Database) error {
	err := db.C(collection).RemoveId(documentID)
	log.DatabaseError(err, collection)
	return err
}

// DeleteDocumentsBySelector : deletes documents into a collection.
func DeleteDocumentsBySelector(collection string, selector interface{}, db *mgo.Database) (int, error) {
	info, err := db.C(collection).RemoveAll(selector)
	log.DatabaseError(err, collection)
	return info.Removed, err
}

// GetDocumentByID : returns a document from a collection.
func GetDocumentByID(document Document, collection string, id string, db *mgo.Database) error {
	if !bson.IsObjectIdHex(id) {
		return errors.NotFound()
	}

	err := db.C(collection).FindId(bson.ObjectIdHex(id)).One(document)
	log.DatabaseError(err, collection)

	if !document.IsSaved() {
		return errors.NotFound()
	}

	return err
}

// GetDocumentsBySelectorAndSortWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySelectorAndSortWithOffsetAndLimit(result interface{}, collection string, selector interface{}, sort []string, offset, limit int, db *mgo.Database) error {
	query := db.C(collection).Find(selector)

	if len(sort) > 0 {
		query = query.Sort(sort...)
	}

	if offset > -1 {
		query = query.Skip(offset)
	}

	if limit > -1 {
		query = query.Limit(limit)
	}

	err := query.All(result)
	log.DatabaseError(err, collection)

	return err
}

// GetDocumentsBySelectorAndSort : returns the documents from a collection
func GetDocumentsBySelectorAndSort(result interface{}, collection string, selector interface{}, sort []string, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, selector, sort, -1, -1, db)
}

// GetDocumentsBySelector : returns the documents from a collection
func GetDocumentsBySelector(result interface{}, collection string, selector interface{}, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, selector, []string{}, -1, -1, db)
}

// GetDocumentsBySelectorWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySelectorWithOffsetAndLimit(result interface{}, collection string, selector interface{}, offset, limit int, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, selector, []string{}, offset, limit, db)
}

// GetDocumentsBySort : returns the documents from a collection
func GetDocumentsBySort(result interface{}, collection string, sort []string, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, nil, sort, -1, -1, db)
}

// GetDocumentsBySortWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsBySortWithOffsetAndLimit(result interface{}, collection string, sort []string, offset, limit int, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, nil, sort, offset, limit, db)
}

// GetDocuments : returns the documents from a collection
func GetDocuments(result interface{}, collection string, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, nil, []string{}, -1, -1, db)
}

// GetDocumentsWithOffsetAndLimit : returns the documents from a collection
func GetDocumentsWithOffsetAndLimit(result interface{}, collection string, offset, limit int, db *mgo.Database) error {
	return GetDocumentsBySelectorAndSortWithOffsetAndLimit(result, collection, nil, []string{}, offset, limit, db)
}

// GetDocumentBySelectorAndSort : returns a document from a collection
func GetDocumentBySelectorAndSort(document Document, collection string, selector interface{}, sort []string, db *mgo.Database) error {
	query := db.C(collection).Find(selector)

	if len(sort) > 0 {
		query = query.Sort(sort...)
	}

	err := query.One(document)
	log.DatabaseError(err, collection)

	if !document.IsSaved() {
		return errors.NotFound()
	}

	return err
}

// GetDocumentBySelector : returns a document from a collection
func GetDocumentBySelector(document Document, collection string, selector interface{}, db *mgo.Database) error {
	return GetDocumentBySelectorAndSort(document, collection, selector, []string{}, db)
}

// GetDocumentBySort : returns a document from a collection
func GetDocumentBySort(document Document, collection string, sort []string, db *mgo.Database) error {
	return GetDocumentBySelectorAndSort(document, collection, nil, sort, db)
}

// GetDocument : returns a document from a collection
func GetDocument(document Document, collection string, db *mgo.Database) error {
	return GetDocumentBySelectorAndSort(document, collection, nil, []string{}, db)
}

// CountBySelector : returns the count of documents from a collection
func CountBySelector(collection string, selector interface{}, db *mgo.Database) (int, error) {
	query := db.C(collection).Find(selector)
	count, err := query.Count()
	log.DatabaseError(err, collection)

	return count, err
}

// Count : returns the count of documents from a collection
func Count(collection string, db *mgo.Database) (int, error) {
	return CountBySelector(collection, nil, db)
}

// PipeAll : returns all the results of a pipeline.
func PipeAll(result interface{}, collection string, pipeline interface{}, db *mgo.Database) error {
	pipe := db.C(collection).Pipe(pipeline)
	err := pipe.All(result)
	log.DatabaseError(err, collection)

	return err
}

// PipeOne : returns the first result of a pipeline.
func PipeOne(result interface{}, collection string, pipeline interface{}, db *mgo.Database) error {
	pipe := db.C(collection).Pipe(pipeline)
	err := pipe.One(result)
	log.DatabaseError(err, collection)

	return err
}
