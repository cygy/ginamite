package model

// URLParams : returns a struct containing the parameters to use into a url.
func (contact AdminContact) URLParams() map[string]string {
	return map[string]string{
		"id": contact.ID.Hex(),
	}
}

// URLParams : returns a struct containing the parameters to use into a url.
func (contact AdminSummarizedContact) URLParams() map[string]string {
	return map[string]string{
		"id": contact.ID.Hex(),
	}
}
