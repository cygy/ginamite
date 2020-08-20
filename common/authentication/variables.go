package authentication

var (
	// BuildTokenFromID : returns the content of a auth token from its ID.
	BuildTokenFromID BuildTokenFunc

	// ExtraPropertiesForTokenWithID : returns some extra properties to include in the auth token.
	ExtraPropertiesForTokenWithID ExtraPropertiesForTokenWithIDFunc

	// ExtendTokenExpirationDateFromID : extends the expiration date of a auth token from its ID.
	ExtendTokenExpirationDateFromID ExtendTokenExpirationDateFunc
)
