package model

// HasLocation : returns true if the token has a defined location.
func (token *AdminAuthToken) HasLocation() bool {
	return token.Location != nil && !token.Location.IsEmpty()
}

// URLParams : returns a struct containing the parameters to use into a url.
func (token *AdminAuthToken) URLParams() map[string]string {
	return map[string]string{
		"id": token.ID.Hex(),
	}
}
