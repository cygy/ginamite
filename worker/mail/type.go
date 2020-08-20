package mail

// Message : struct representing the details of a mail message.
type Message struct {
	Subject   string
	Recipient string
	Template  string
	HTMLVars  map[string]interface{}
	TextVars  map[string]interface{}
	Locale    string
}

// IsEmailAddressValidFunc : function to know if an e-mail can be sent to an email address.
type IsEmailAddressValidFunc func(emailAddress string) bool
