package authentication

// User holds the details of a user.
type User struct {
	ID           string
	Username     string
	EmailAddress string
	Locale       string
	PrivateKey   string
	IsEmailValid bool
}

// IsSaved reports whether the user is saved, i.e. has an ID.
func (user *User) IsSaved() bool {
	return len(user.ID) > 0
}
