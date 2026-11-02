// Package model defines the documents stored in MongoDB by the application and the queries to read and
// update them.
//
// Each document type implements document.Document and lives in its own file, with the file named after the
// type. The views exposed by the API sit next to it: Admin* for the administrators, Own* for the owner of the
// document and Public* for everyone. All the queries take a store.Database so that they can be tested with
// the stub of the store/mock package.
package model
