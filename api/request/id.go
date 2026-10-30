package request

// IsValidID reports whether id is a well-formed identifier of a stored entity, such as a notification.
//
// The API server replaces it with the Handlers.IsValidID function of the database plugin when it starts, because the
// format of the identifiers depends on the database. By default every identifier is valid.
var IsValidID = func(id string) bool { return true }
