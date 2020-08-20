package authentication

// IsSaved : returns true if the struct 'User' is saved.
func (user *User) IsSaved() bool {
	return len(user.ID) > 0
}
