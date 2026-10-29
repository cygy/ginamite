package mail

import "context"

// Message : struct representing the details of a mail message.
type Message struct {
	Subject   string
	Recipient string
	Template  string
	HTMLVars  map[string]any
	TextVars  map[string]any
	Locale    string
}

// IsEmailAddressValidFunc : function to know if an e-mail can be sent to an email address.
type IsEmailAddressValidFunc func(ctx context.Context, emailAddress string) (bool, error)

func (f IsEmailAddressValidFunc) check(ctx context.Context, emailAddress string) bool {
	isValid, _ := f(ctx, emailAddress)
	return isValid
}
