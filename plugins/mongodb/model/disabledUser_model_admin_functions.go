package model

// URLParams : returns a struct containing the parameters to use into a url.
func (user AdminDisabledUser) URLParams() map[string]string {
	return map[string]string{
		"id": user.ID.Hex(),
	}
}
