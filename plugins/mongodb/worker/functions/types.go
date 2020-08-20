package functions

import "github.com/globalsign/mgo"

// IndexToCreate : struct representing the index of a collection.
type IndexToCreate struct {
	Collection string
	Index      mgo.Index
}

// IndexToDelete : struct representing the index of a collection.
type IndexToDelete struct {
	Collection string
	IndexName  string
}

// CollectionToCreate : struct representing a collection.
type CollectionToCreate struct {
	Name string
	Info mgo.CollectionInfo
}
