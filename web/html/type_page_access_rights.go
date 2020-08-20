package html

// AccessRight rights to access a page.
type AccessRight string

const (
	// AccessRightAll concerns all the users.
	AccessRightAll = "all"
	// AccessRightSignedIn concerns only the signed in users.
	AccessRightSignedIn = "signedin"
	// AccessRightNotSignedIn concerns only the non-signed in users.
	AccessRightNotSignedIn = "notsignedin"
	// AccessRightAttributes concerns only the users with some attributes.
	AccessRightAttributes = "attributes"
)
