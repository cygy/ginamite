package authentication

// User : struct representing the details of a user.
type User struct {
	ID           string
	Username     string
	EmailAddress string
	Locale       string
	PrivateKey   string
	IsEmailValid bool
}
